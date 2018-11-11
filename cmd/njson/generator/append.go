package generator

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"github.com/alxarch/meta"
)

// WriteAppender writes an AppendJSON method for a type
func (g *Generator) WriteAppender(typName string) (err error) {
	code := g.Appender(typName)
	if err = code.Err(); err != nil {
		return
	}
	g.Import(code.Imports...)
	_, err = g.buffer.Write(code.Code)
	return
}

// Appender returns the AppendJSON method code for a type
func (g *Generator) Appender(typName string) (c meta.Code) {
	typ := g.LookupType(typName)
	if typ == nil {
		return c.Errorf("Type %s not found", typName)
	}
	receiverName := strings.ToLower(typName[:1])
	method := g.AppendMethod()
	return g.Code(`
		func (%[1]s *%[2]s) %[3]s(out []byte) ([]byte, error) {
			if v := %[1]s; v != nil {
				%[4]s
			} else {
				out = append(out, "null"...)
			}
			return out, nil
		}
	`, receiverName, typ, method, g.TypeAppender(typ, nil))

}

// OmiterType returns the interface for an omiter
func (g *Generator) OmiterType() (string, *types.Interface) {
	if g.omiter == nil {
		g.omiter = typOmiter
	}
	return g.omiter.Method(0).Name(), g.omiter
}

// TypeOmiter returns the code block to check if a value should be omited
func (g *Generator) TypeOmiter(typ types.Type, block meta.Code) meta.Code {
	cond := ""
	if method, omiter := g.OmiterType(); types.Implements(typ, omiter) {
		cond = `!v.` + method + `()`
	} else {
		switch t := typ.Underlying().(type) {
		case *types.Pointer:
			cond = `v != nil`
		case *types.Map:
			cond = `len(v) > 0`
		case *types.Slice:
			cond = `len(v) > 0`
		case *types.Struct:
			return block.Error(typeError{typ})

		case *types.Basic:
			switch t.Kind() {
			case types.Bool:
				cond = `v`
			case types.String:
				cond = `len(v) > 0`
			default:
				if t.Info()&types.IsNumeric != 0 {
					cond = `v != 0`
				} else {
					return block.Error(fmt.Errorf("Omit basic error %s", typeError{typ}))
				}
			}
		case *types.Interface:
			if t.NumMethods() == 0 {
				cond = `v != nil`
			}
		}

	}
	if cond == "" {
		return block
	}
	return g.Code(`if %s {
		%s
	}`, cond, block)

}

// EnsureReversePath returns a code block to check the path to an embedded field does not contain any nil pointers
func (g *Generator) EnsureReversePath(path meta.FieldPath, code meta.Code) meta.Code {
	cond := []string{}
	for i, p := range path {
		if _, ok := p.Type().Underlying().(*types.Pointer); ok {
			cond = append(cond, fmt.Sprintf("v%s != nil", path[:i+1]))
		}
	}

	if len(cond) > 0 {
		return g.Code(`
		if %s {
			v := v%s
			%s
		}`, strings.Join(cond, " && "), path, code)
	}
	return g.Code(`
	{
		v := v%s
		%s
	}`, path, code)
}

// StructAppender returns the AppendJSON block code for a struct
func (g *Generator) StructAppender(fields meta.Fields) (c meta.Code) {
	c = c.Println(`more := 0`)
	sortedFields := []meta.Field{}
	for name := range fields {
		sortedFields = append(sortedFields, fields[name]...)
	}
	sort.SliceStable(sortedFields, func(i, j int) bool {
		return sortedFields[i].Tag < sortedFields[j].Tag
	})
	used := make(map[string]bool)
	for _, field := range sortedFields {
		field = field.WithTag(g.TagKey())
		name, tag, ok := g.parseField(field.Var, field.Tag)
		if !ok {
			continue
		}
		if tag.Name != "" {
			name = tag.Name
		}
		if used[name] {
			continue
		}
		used[name] = true
		var cf meta.Code
		cf = g.TypeAppender(field.Type(), tag.Params)
		cf = g.Code(`
				out = append(out, "{,"[more])
				more = 1
				out = append(out, '"')
				out = append(out, "%s"...)
				out = append(out, '"', ':')
				{
					%s
				}`, name, cf)
		if tag.Params.Has(paramOmitempty) {
			cf = g.TypeOmiter(field.Type(), cf)
		}
		cf = g.EnsureReversePath(field.Path, cf)
		c = c.Append(cf)

	}
	c = c.Println(`
	out = append(out, "{}"[more:]...)`)
	return
}

// TypeAppender returnds the AppendJSON block code for a type
func (g *Generator) TypeAppender(typ types.Type, params meta.Params) (c meta.Code) {
	switch {
	case types.Implements(typ, typJSONAppender):
		return g.Code(`
		var err error
		if out, err = v.AppendJSON(out); err != nil {
			return out, err
		}
		`)
	case types.Implements(typ, typJSONMarshaler):
		return g.Code(`
		data, err := v.MarshalJSON()
		if err != nil {
			return out, err
		}
		out = append(out, data...)
		`)
	case types.Implements(typ, typTextMarshaler):
		return g.Code(`
		data, err := v.MarshalText()
		if err != nil {
			return out, err
		}
		out = append(out, '"')
		out = append(out, data...)
		out = append(out, '"')
		`)

	}
	switch t := typ.Underlying().(type) {
	case *types.Pointer:
		return g.Code(`
			if v == nil {
				out = append(out, "null"...)
			} else {
				%s
			}
		`, g.TypeAppender(t.Elem(), params))
	case *types.Map:
		return g.Code(`
			{
				more := 0
				for k, v := range v {
					out = append(out, "{,"[more])
					more = 1
					out = append(out, '"')
					{
						v := k
						%s
					}
					out = append(out, '"', ':')
					%s
				}
				out = append(out, "{}"[more:]...)
			}
		`, g.TypeAppender(t.Key(), nil), g.TypeAppender(t.Elem(), params))
	case *types.Slice:
		return g.Code(`
			out = append(out, '[')
			for i, v := range v {
				if i > 0 {
					out = append(out, ',')
				}
				%s
			}
			out = append(out, ']')
		`, g.TypeAppender(t.Elem(), params))
	case *types.Struct:
		fields := meta.NewFields(t, true)
		return g.StructAppender(fields)
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return c.Println(`if v { out = append(out, "true"...) } else { out = append(out, "false"...) }`)
		case types.String:
			if params.Has("raw") {
				return c.Println(`
					out = append(out, '"')
					out = append(out, v...)
					out = append(out, '"')`)
			}
			return c.Println(`
				out = append(out, '"')
				out = strjson.AppendEscaped(out, v, false)
				out = append(out, '"')`).Import(strjsonPkg)
		default:
			if info := t.Info(); info&types.IsFloat != 0 {
				return c.Println(`out = numjson.AppendFloat(out, float64(v), 64)`).Import(numjsonPkg)
			} else if info&types.IsUnsigned != 0 {
				return c.Println(`out = strconv.AppendUint(out, uint64(v), 10)`).Import(strconvPkg)
			} else if info&types.IsInteger != 0 {
				return c.Println(`out = strconv.AppendInt(out, int64(v), 10)`).Import(strconvPkg)
			}
			return c.Error(typeError{typ})
		}
	case *types.Interface:
		if t.Empty() {
			return c.Println(`
				// Fallback to json.Marshal for empty interface
				if data, err := json.Marshal(v); err == nil {
					out = append(out, data)
				} else {
					return out, err
				}`).Import(jsonPkg)
		}
		return c.Error(typeError{typ})
	default:
		return c.Error(typeError{typ})

	}
}

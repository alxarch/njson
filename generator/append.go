package generator

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/alxarch/njson"
	"github.com/alxarch/njson/njsonutil"

	"github.com/alxarch/meta"
)

func (g *Generator) WriteAppender(typName string) (err error) {
	code := g.Appender(typName)
	if err = code.Err(); err != nil {
		return
	}
	g.Import(code.Imports...)
	_, err = g.buffer.Write(code.Code)
	return
}

func (g *Generator) Appender(typName string) (c meta.Code) {
	typ := g.LookupType(typName)
	if typ == nil {
		return c.Errorf("Type %s not found", typName)
	}
	receiverName := strings.ToLower(typName[:1])
	method := g.AppenderMethodName()
	return g.Code(`
		func (%[1]s *%[2]s) %[3]s(out []byte) ([]byte, error) {
			if v := %[1]s; v != nil {
				%[4]s
			} else {
				out = append(out, "null"...)
			}
			return out, nil
		}
	`, receiverName, typ, method, g.TypeAppender(typ))
	return

}

var (
	typJSONAppender    = reflect.TypeOf((*njson.Appender)(nil)).Elem()
	methodAppendJSON   = typJSONAppender.Method(0)
	appenderMethodName = methodAppendJSON.Name
)

func (g *Generator) AppenderMethodName() string {
	return njsonutil.TaggedMethodName(appenderMethodName, g.TagKey())
}

func (g *Generator) OmiterType() (string, *types.Interface) {
	name := njsonutil.TaggedMethodName("Omit", g.TagKey())
	results := []types.Type{types.Typ[types.Bool]}
	return name, meta.MakeInterface(name, nil, results, false)

}
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
				}
				return block.Error(typeError{typ})
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

func (g *Generator) EnsureReversePath(path meta.FieldPath, code meta.Code) meta.Code {
	cond := []string{}
	for i, p := range path {
		if _, ok := p.Type().Underlying().(*types.Pointer); ok {
			cond = append(cond, fmt.Sprintf("v%s != nil", path[:i+1]))
		}
	}

	if len(cond) > 0 {
		return g.Code(`if %s {
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

func (g *Generator) StructAppender(fields meta.Fields) (c meta.Code) {
	c = c.Println(`more := 0`)
	for name := range fields {
		used := make(map[string]bool)
		for _, field := range fields[name] {
			field = field.WithTag(g.TagKey())
			name, tag, ok := g.parseField(field.Var, field.Tag)
			if !ok || used[name] {
				continue
			}
			used[name] = true
			cf := g.TypeAppender(field.Type())
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
	}
	c = c.Println(`
	if more == 1 {
		out = append(out, '}')
	}`)
	return
}
func (g *Generator) TypeAppender(typ types.Type) (c meta.Code) {
	switch t := typ.Underlying().(type) {
	case *types.Pointer:
		return g.Code(`
			if v == nil {
				out = append(out, "null"...)
			} else {
				%s
			}
		`, g.TypeAppender(t.Elem()))
	case *types.Map:
		return g.Code(`
			out = append(out, '{')
			{
				first := true
				for k, v := range v {
					if first {
						first := false
					} else {
						out = append(out, ',')
					}
					out = append(out, '"')
					%s
					out = append(out, '"', ':')
					%s
				}
			}
			out = append(out, '}')
		`, g.TypeAppender(t.Key()), g.TypeAppender(t.Elem()))
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
		`, g.TypeAppender(t.Elem()))
	case *types.Struct:
		fields := meta.NewFields(t, true)
		return g.StructAppender(fields)
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return c.Println(`if v { out = append(out, "true"...) } else { out = append(out, "false") }`)
		case types.String:
			return c.Println(`
				out = append(out, '"')
				out = njson.AppendEscaped(out, v)
				out = append(out, '"')`).Import(njsonPkg)
		default:
			if info := t.Info(); info&types.IsFloat != 0 {
				return c.Println(`out = strconv.AppendFloat(out, float64(v), 'f', 6, 10)`).Import(strconvPkg)
			} else if info&types.IsUnsigned != 0 {
				return c.Println(`out = strconv.AppendUint(out, uint64(v), 10)`).Import(strconvPkg)
			} else if info&types.IsInteger != 0 {
				return c.Println(`out = strconv.AppendInt(out, int64(v), 10)`).Import(strconvPkg)
			} else {
				return c.Error(typeError{typ})
			}
		}
	case *types.Interface:
		if t.NumMethods() == 0 {
			return c.Println(`
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

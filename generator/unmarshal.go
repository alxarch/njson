package generator

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/alxarch/meta"
)

// Unmarshaler generates an unmarshaler method for a type
func (g *Generator) Unmarshaler(typeName string) (code meta.Code) {
	typ := g.LookupType(typeName)
	if typ == nil {
		return code.Errorf("Type %s not found", typeName)
	}
	receiverName := strings.ToLower(typeName[:1])
	return g.Code(`
		func (%[1]s *%[2]s) %[3]s(node njson.Node) error {
			typ := node.Type()
			if !typ.IsValue() {
				return node.TypeError(njson.TypeAnyValue)
			}
			if typ == njson.TypeNull {
				return nil
			}
			{
				r := %[1]s
				n := node
				{
					%[4]s
				}

			}
			return nil
		}
	`, receiverName, typ, g.UnmarshalMethod(), g.TypeUnmarshaler(typ)).Import(njsonPkg)
}

// WriteUnmarshaler writes an unmarshaler method for a type in the generator's buffer.
func (g *Generator) WriteUnmarshaler(typeName string) (err error) {
	code := g.Unmarshaler(typeName).Format()
	if err = code.Err(); err != nil {
		return
	}

	g.Import(code.Imports...)
	_, err = g.buffer.Write(code.Code)
	return
}

// TypeUnmarshaler returns the code block for unmarshaling a type.
func (g *Generator) TypeUnmarshaler(t types.Type) (code meta.Code) {
	if t == nil {
		return code.Error(typeError{t})
	}

	pt := types.NewPointer(t)
	switch {
	case types.Implements(pt, typNodeJSONUnmarshaler):
		return g.NodeJSONUnmarshaler(t)
	case types.Implements(pt, typJSONUnmarshaler):
		return g.JSONUnmarshaler(t)
	case types.Implements(pt, typTextUnmarshaler):
		return g.TextUnmarshaler(t)
	}

	switch typ := t.Underlying().(type) {
	case *types.Map:
		return g.MapUnmarshaler(t, typ)
	case *types.Struct:
		return g.StructUnmarshaler(typ)
	case *types.Slice:
		return g.SliceUnmarshaler(t, typ)
	case *types.Pointer:
		return g.PointerUnmarshaler(t, typ)
	case *types.Basic:
		return g.BasicUnmarshaler(t, typ)
	case *types.Interface:
		if typ.Empty() {
			return g.InterfaceUnmarshaler(t, typ)
		}
		return code.Error(typeError{t})
	default:
		return code.Error(typeError{t})
	}
}

// SliceUnmarshaler generates the code block to unmarshal a slice.
func (g *Generator) SliceUnmarshaler(T types.Type, t *types.Slice) meta.Code {
	return g.Code(`
switch n.Type() {
case njson.TypeArray:
	// Ensure slice is big enough
	values := n.Values()
	size := values.Len()
	
	if cap(*r) < size {
		*r = make([]%s, len(*r) + size)
	} else {
		*r = (*r)[:size]
	}
	
	s := *r
	for i := 0; values.Next() && 0 <= i && i < len(s) ; i++ {
		r := &s[i]
		n := n.With(values.ID())
		%s
	}

case njson.TypeNull:
	*r = nil
default:
	return n.TypeError(njson.TypeArray|njson.TypeNull)
}
`, t.Elem(), g.TypeUnmarshaler(t.Elem())).Import(njsonPkg)
}

// PointerUnmarshaler generates the code block to unmarshal a pointer type.
func (g *Generator) PointerUnmarshaler(T types.Type, t *types.Pointer) (code meta.Code) {
	return g.Code(`
switch {
case n.Type() == njson.TypeNull:
	*r = nil
case *r == nil:
	*r = new(%s)
	fallthrough
default:
	r := *r
	%s
}
`, t.Elem(), g.TypeUnmarshaler(t.Elem()))
}

// InterfaceUnmarshaler generates the code block to unmarshal an empty interface.
func (g *Generator) InterfaceUnmarshaler(t types.Type, b *types.Interface) (code meta.Code) {
	return code.Import(njsonPkg).Println(`if x, ok := n.ToInterface(); ok { *r = x } else { return n.TypeError(njson.AnyValue) }`)
}

// BasicUnmarshaler generates the code block to unmarshal a basic type.
func (g *Generator) BasicUnmarshaler(t types.Type, b *types.Basic) (code meta.Code) {
	var c string
	switch b.Kind() {
	case types.Bool:
		c = "if b, ok := n.ToBool(); ok { *r = %s(b) } else { return n.TypeError(njson.TypeBoolean) }"
	case types.String:
		return g.Code("*r = %s(n.Unescaped())", t)
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		c = "if i, ok := n.ToInt(); ok { *r = %s(i) } else { return n.TypeError(njson.TypeNumber) }"
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
		c = ("if u, ok := n.ToUint(); ok { *r = %s(u) } else { return n.TypeError(njson.TypeNumber) }")
	case types.Float32, types.Float64:
		c = "if f, ok := n.ToFloat(); ok { *r = %s(f) } else { return n.TypeError(njson.TypeNumber) }"
	default:
		return code.Error(typeError{t})
	}
	return g.Code(c, t).Import(njsonPkg)
}

// EnsurePath generates a code block to ensure the path to an embedded pointer to struct has no nils.
func (g *Generator) EnsurePath(path meta.FieldPath) (code meta.Code) {
	if last := len(path) - 1; last > 0 {
		for i := 0; i < last; i++ {
			f := &path[i]
			t := f.Type()
			if t == nil {
				return
			}
			if p, ok := t.(*types.Pointer); ok {
				r := g.Code("r%s", path[:i+1])
				code = g.Code("%[1]s\nif %[2]s == nil { %[2]s = new(%[3]s) }\n", code, r, p.Elem())
			}
		}
	}
	return
}

// NodeJSONUnmarshaler generates code to wrap the UnmarshalNodeJSON method of a value.
func (g *Generator) NodeJSONUnmarshaler(t types.Type) (code meta.Code) {
	return g.Code(`
	if err := v.%s(n); err != nil {
		return err
	}
	`, methodNodeUnmarshalJSON.Name())
}

// JSONUnmarshaler generates code to wrap the UnmarshalJSON method of a value.
func (g *Generator) JSONUnmarshaler(t types.Type) (code meta.Code) {
	return g.Code(`
	if err := n.WrapUnmarshalJSON(r); err != nil {
		return nil
	}
	`)
}

// TextUnmarshaler generates code to wrap the UnmarshalText method of a value.
func (g *Generator) TextUnmarshaler(t types.Type) (code meta.Code) {
	return g.Code(`
	return n.WrapUnmarshalText(r)
	`)
}

// MapUnmarshaler generates the code block to unmarshal a map.
func (g *Generator) MapUnmarshaler(t types.Type, m *types.Map) (code meta.Code) {
	// TODO: Enforce string, TextUnmarshaler key type
	typK := m.Key()
	// var codeK meta.Code
	// switch {
	// case meta.IsString(typK):
	// 	codeK = g.TypeUnmarshaler(typK)
	// case types.Implements(typK, typTextUnmarshaler):
	// 	codeK = g.TypeUnmarshaler(typK)
	// default:
	// 	return code.Errorf("Invalid key type %s", typK)
	// }
	typV := m.Elem()
	codeV := g.TypeUnmarshaler(typV)
	return g.Code(`
switch n.Type() {
case njson.TypeNull:
	*r = nil
case njson.TypeObject:
	values := n.Values()
	if *r == nil {
		*r = make(map[%[1]s]%[2]s, values.Len())
	}
	m := *r
	for values.Next() {
		var v %[2]s
		{
			n := n.With(values.ID())
			r := &v
			%[3]s
		}
		m[values.Key()] = v
	}
default:
	return n.TypeError(njson.TypeObject|njson.TypeNull)
}
`, typK, typV, codeV).Import(njsonPkg)
}

func (g *Generator) RawStringUnmarshaler(t types.Type) meta.Code {
	return g.Code(`*r = %s(n.Raw())`, t)
}

// StructUnmarshaler generates the code block to unmarshal a struct.
func (g *Generator) StructUnmarshaler(t *types.Struct) (code meta.Code) {
	fields := meta.NewFields(t, true)
	tagKey := g.TagKey()
	used := make(map[string]bool)
	for name := range fields {
		for _, field := range fields[name] {
			field = field.WithTag(tagKey)
			if field.Name() == "_" {
				continue
			}
			if !CanUnmarshal(field.Type()) {
				continue
			}
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
			if tag.Params.Has("raw") && meta.IsString(field.Type()) {
				cf = g.RawStringUnmarshaler(field.Type())
			} else {
				cf = g.TypeUnmarshaler(field.Type())
			}
			code = g.Code(`%s
				case %s:
					n := n.With(values.ID())
					%s{
						r := &r%s
						%s
					}
					`, code, fmt.Sprintf("`%s`", name), g.EnsurePath(field.Path), field.Path, cf)
			if code.Err() != nil {
				return
			}
		}
	}
	return g.Code(`
		if n.Type() != njson.TypeObject {
			return n.TypeError(njson.TypeObject)
		}
		for values := n.Values(); values.Next(); {
			switch values.Key() {
				%s
			}
		}`, code).Import(njsonPkg)

}

// CanUnmarshal returns if can be unmarshaled
func CanUnmarshal(t types.Type) bool {
	if t == nil {
		return false
	}
	switch typ := t.Underlying().(type) {
	case *types.Map:
		return true
	case *types.Struct:
		return typ.NumFields() > 0
	case *types.Slice:
		return CanUnmarshal(typ.Elem())
	case *types.Pointer:
		return CanUnmarshal(typ.Elem())
	case *types.Basic:
		switch typ.Kind() {
		case types.Bool,
			types.String,
			types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64:
			return true
		default:
			return false
		}
	case *types.Interface:
		return typ.Empty()
	case *types.Chan, *types.Tuple, *types.Signature:
		return false
	default:
		return false
	}

}

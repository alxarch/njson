package generator

// import (
// 	"go/types"

// 	"github.com/alxarch/njson"
// )

// // StructField describes a struct's field.
// type StructField struct {
// 	*types.Var
// 	NameJSON  string
// 	OmitEmpty bool
// 	Path      FieldPath
// }

// func (f *StructField) String() string {
// 	return f.Var.String() + " " + f.Path.String()
// }

// // StructFields is a map of a struct's fields
// type StructFields map[string]StructField

// // Add adds a field to a field map handling duplicates.
// func (fields StructFields) Add(f *types.Var, name string, omitempty bool, path FieldPath) {
// 	name = string(njson.AppendEscaped(nil, name))
// 	_, duplicate := fields[name]
// 	if duplicate && ComparePaths(fields[name].Path, path) == -1 {
// 		// keep existing
// 		return
// 	}
// 	fields[name] = StructField{f, name, omitempty, path.Copy()}
// }

// // FieldIndex is a part of the path of a field in a struct.
// type FieldIndex struct {
// 	Index int
// 	Type  types.Type
// 	Name  string
// }

// func (i FieldIndex) String() string {
// 	return i.Name
// }

// // FieldPath is the path of a field in a struct.
// type FieldPath []FieldIndex

// func fieldTypeName(t types.Type) string {
// 	switch t := t.(type) {
// 	case *types.Named:
// 		return t.Obj().Name()
// 	case *types.Pointer:
// 		return fieldTypeName(t.Elem())
// 	default:
// 		return t.String()
// 	}
// }

// func (p FieldPath) String() string {
// 	buf := make([]byte, 0, len(p)*16)
// 	for i := range p {
// 		buf = append(buf, '.')
// 		buf = append(buf, p[i].Name...)
// 	}
// 	return string(buf)
// }

// // Copy creates a copy of a path.
// func (p FieldPath) Copy() FieldPath {
// 	cp := make([]FieldIndex, len(p))
// 	copy(cp, p)
// 	return cp
// }

// // ComparePaths compares the paths of two fields.
// func ComparePaths(a, b FieldPath) int {
// 	if len(a) < len(b) {
// 		return -1
// 	}
// 	if len(b) < len(a) {
// 		return 1
// 	}
// 	for i := range a {
// 		if a[i].Index < b[i].Index {
// 			return -1
// 		}
// 		if b[i].Index < a[i].Index {
// 			return 1
// 		}
// 	}
// 	return 0
// }

// // MergeFields merges a struct's fields to a field map.
// func (g *Generator) MergeFields(fields StructFields, s *types.Struct, path FieldPath) error {
// 	if s == nil {
// 		return nil
// 	}

// 	// depth := len(path)

// 	for i := 0; i < s.NumFields(); i++ {
// 		field := s.Field(i)
// 		// name, omitempty, tagged, skip := g.parseField(field, s.Tag(i))
// 		// if skip {
// 		// 	continue
// 		// }

// 		// path = append(path[:depth], FieldIndex{i, field.Type(), field.Name()})
// 		// if !tagged && field.Anonymous() {
// 		// 	t := meta.Resolve(field.Type())
// 		// 	if ptr, isPointer := t.(*types.Pointer); isPointer {
// 		// 		t = ptr.Elem()
// 		// 	}
// 		// 	name = g.JSONFieldName(fieldTypeName(t))
// 		// 	tt := meta.Resolve(t)
// 		// 	if tt, ok := tt.(*types.Struct); ok {
// 		// 		// embedded struct
// 		// 		if err := g.MergeFields(fields, tt, path); err != nil {
// 		// 			return err
// 		// 		}
// 		// 		continue
// 		// 	}
// 		// }
// 		if CanUnmarshal(field.Type()) {
// 			// 	fields.Add(field, name, omitempty, path)
// 		}

// 	}
// 	return nil
// }

// // StructFields creates a fields map for a struct.
// func (g *Generator) StructFields(s *types.Struct, path FieldPath) (StructFields, error) {
// 	size := 0
// 	if s != nil {
// 		size = s.NumFields()
// 	}
// 	fields := StructFields(make(map[string]StructField, size))
// 	err := g.MergeFields(fields, s, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return fields, nil
// }

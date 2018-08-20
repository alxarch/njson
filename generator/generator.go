package generator

/*
Package generator is a code generator for njson.Unmarsaler

It parses a package dir and can generate (t *T)UnmarshalNodeJSON(*njson.Node) error methods.
*/

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/alxarch/njson"
)

// TODO: handle json.Unmarshaler
// TODO: handle njson.Unmarshaler
// TODO: handle encoding.TextUnmarshaler
// TODO: handle flag combinations (ie if tag key is not json don't use UnmarshalJSON)
// TODO: generate AppendJSON([]byte) ([]byte, error) methods

// Generator is a source code generator for njson unmarshal methods.
type Generator struct {
	options
	pkg     *types.Package
	buffer  bytes.Buffer
	imports map[string]*types.Package
	files   []*ast.File
	info    types.Info
}

// ResolveType returns the underlying unnamed type.
func ResolveType(typ types.Type) types.Type {
	for typ != nil {
		if _, ok := typ.(*types.Named); ok {
			typ = typ.Underlying()
		} else {
			break
		}
	}
	return typ
}

// LookupType looks up a named type in the package's definitions.
func (g *Generator) LookupType(name string) (t *types.Named) {
	for _, def := range g.info.Defs {
		if def == nil {
			continue
		}
		typ := def.Type()
		if typ == nil {
			continue
		}
		if typ, ok := typ.(*types.Named); ok {
			if obj := typ.Obj(); obj != nil && obj.Name() == name {
				return typ
			}
		}
	}
	return nil
}

// StructField describes a struct's field.
type StructField struct {
	*types.Var
	NameJSON  string
	OmitEmpty bool
	Path      FieldPath
}

func (f *StructField) String() string {
	return f.Var.String() + " " + f.Path.String()
}

// StructFields is a map of a struct's fields
type StructFields map[string]StructField

// Add adds a field to a field map handling duplicates.
func (fields StructFields) Add(f *types.Var, name string, omitempty bool, path FieldPath) {
	name = string(njson.EscapeString(nil, name))
	_, duplicate := fields[name]
	if duplicate && ComparePaths(fields[name].Path, path) == -1 {
		// keep existing
		return
	}
	fields[name] = StructField{f, name, omitempty, path.Copy()}
}

// FieldIndex is a part of the path of a field in a struct.
type FieldIndex struct {
	Index int
	Type  types.Type
	Name  string
}

func (i FieldIndex) String() string {
	return i.Name
}

// FieldPath is the path of a field in a struct.
type FieldPath []FieldIndex

func fieldTypeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Pointer:
		return fieldTypeName(t.Elem())
	default:
		return t.String()
	}
}

// TypeName resolves a type's local name in the scope of the generator's package.
func (g *Generator) TypeName(t types.Type) string {
	name, pkg := g.resolveTypeName(t)
	if pkg != nil {
		g.Import(pkg)
	}
	return name
}

func (g *Generator) resolveTypeName(t types.Type) (string, *types.Package) {
	switch t := t.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg == g.pkg {
			return t.Obj().Name(), nil
		}
		return t.String(), pkg
	case *types.Pointer:
		return g.resolveTypeName(t.Elem())
	default:
		return t.String(), nil
	}
}

func (p FieldPath) String() string {
	buf := make([]byte, 0, len(p)*16)
	for i := range p {
		buf = append(buf, '.')
		buf = append(buf, p[i].Name...)
	}
	return string(buf)
}

// Copy creates a copy of a path.
func (p FieldPath) Copy() FieldPath {
	cp := make([]FieldIndex, len(p))
	copy(cp, p)
	return cp
}

// ComparePaths compares the paths of two fields.
func ComparePaths(a, b FieldPath) int {
	if len(a) < len(b) {
		return -1
	}
	if len(b) < len(a) {
		return 1
	}
	for i := range a {
		if a[i].Index < b[i].Index {
			return -1
		}
		if b[i].Index < a[i].Index {
			return 1
		}
	}
	return 0
}

// MergeFields merges a struct's fields to a field map.
func (g *Generator) MergeFields(fields StructFields, s *types.Struct, path FieldPath) error {
	if s == nil {
		return nil
	}

	depth := len(path)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		name, omitempty, tagged, skip := g.parseField(field, s.Tag(i))
		if skip {
			continue
		}

		path = append(path[:depth], FieldIndex{i, field.Type(), field.Name()})
		if !tagged && field.Anonymous() {
			t := ResolveType(field.Type())
			if ptr, isPointer := t.(*types.Pointer); isPointer {
				t = ptr.Elem()
			}
			name = g.JSONFieldName(fieldTypeName(t))
			tt := ResolveType(t)
			if tt, ok := tt.(*types.Struct); ok {
				// embedded struct
				if err := g.MergeFields(fields, tt, path); err != nil {
					return err
				}
				continue
			}
		}
		if CanUnmarshal(field.Type()) {
			fields.Add(field, name, omitempty, path)
		}

	}
	return nil
}

// StructFields creates a fields map for a struct.
func (g *Generator) StructFields(s *types.Struct, path FieldPath) (StructFields, error) {
	size := 0
	if s != nil {
		size = s.NumFields()
	}
	fields := StructFields(make(map[string]StructField, size))
	err := g.MergeFields(fields, s, nil)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// SliceUnmarshaler generates the code block to unmarshal a slice.
func (g *Generator) SliceUnmarshaler(T types.Type, t *types.Slice) (code string, err error) {
	body, err := g.TypeUnmarshaler(t.Elem())
	if err != nil {
		return
	}
	typeName := g.TypeName(t.Elem())
	code = fmt.Sprintf(`
switch {
case n.IsArray():
	// Ensure slice is big enough
	if cap(*r) < n.Len() {
		*r = make([]%s, len(*r) + n.Len())
	} else {
		*r = (*r)[:n.Len()]
	}
	s := *r
	for i, n := 0, n.Value(); n != nil; n, i = n.Next(), i+1 {
		r := &s[i]
		%s
	}
case n.IsNull():
	*r = nil
default:
	return n.TypeError(njson.TypeArray|njson.TypeNull)
}
`, typeName, body)
	return
}

// PointerUnmarshaler generates the code block to unmarshal a pointer type.
func (g *Generator) PointerUnmarshaler(T types.Type, t *types.Pointer) (code string, err error) {
	body, err := g.TypeUnmarshaler(t.Elem())
	if err != nil {
		return
	}
	typeName := g.TypeName(t.Elem())
	code = fmt.Sprintf(`
switch {
case n.IsNull():
	*r = nil
case *r == nil:
	*r = new(%s)
	fallthrough
default:
	r := *r
	%s
}
`, typeName, body)
	return
}

// InterfaceUnmarshaler generates the code block to unmarshal an empty interface.
func (g *Generator) InterfaceUnmarshaler(t types.Type, b *types.Interface) (code string, err error) {
	return `if x, ok := n.ToInterface(); ok { *r = x } else { return n.TypeError(njson.AnyValue) }`, nil
}

// BasicUnmarshaler generates the code block to unmarshal a basic type.
func (g *Generator) BasicUnmarshaler(t types.Type, b *types.Basic) (code string, err error) {
	switch b.Kind() {
	case types.Bool:
		code = "if b, ok := n.ToBool(); ok { *r = %s(b) } else { return n.TypeError(njson.TypeBool) }"
	case types.String:
		code = "*r = %s(n.Unescaped())"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		code = "if i, ok := n.ToInt(); ok { *r = %s(i) } else { return n.TypeError(njson.TypeNumber) }"
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
		code = ("if u, ok := n.ToUint(); ok { *r = %s(u) } else { return n.TypeError(njson.TypeNumber) }")
	case types.Float32, types.Float64:
		code = "if f, ok := n.ToFloat(); ok { *r = %s(f) } else { return n.TypeError(njson.TypeNumber) }"
	default:
		return "", typeError{t}
	}
	return fmt.Sprintf(code, g.TypeName(t)), nil
}

// EnsurePath generates a code block to ensure the path to an embedded pointer to struct has no nils.
func (g *Generator) EnsurePath(path FieldPath) (code string) {
	if last := len(path) - 1; last > 0 {
		for i := 0; i < last; i++ {
			f := &path[i]
			t := ResolveType(f.Type)
			if t == nil {
				return
			}
			if _, ok := t.(*types.Pointer); ok {
				v := "r" + path[:i+1].String()
				typeName := g.TypeName(f.Type)
				code += fmt.Sprintf("if %[1]s == nil { %[1]s = new(%[2]s) }\n", v, typeName)
			}
		}
	}
	return
}

// MapUnmarshaler generates the code block to unmarshal a map.
func (g *Generator) MapUnmarshaler(t types.Type, m *types.Map) (code string, err error) {
	typK := g.TypeName(m.Key())
	typV := g.TypeName(m.Elem())
	bodyK, err := g.TypeUnmarshaler(m.Key())
	if err != nil {
		return
	}
	bodyV, err := g.TypeUnmarshaler(m.Elem())
	if err != nil {
		return
	}
	code = fmt.Sprintf(`
switch {
case n.IsNull():
	*r = nil
case !n.IsObject():
	return n.TypeError(njson.TypeObject|njson.TypeNull)
case *r == nil:
	*r = make(map[%[1]s]%[2]s, n.Len())
	fallthrough
default:
	m := *r
	for n := n.Value(); n != nil; n = n.Next() {
		var k %[1]s
		{
			r := &k
			%[3]s
		}
		var v %[2]s
		{
			r := &v
			%[4]s
		}
		m[k] = v
	}
}
`, typK, typV, bodyK, bodyV)
	return
}

// StructUnmarshaler generates the code block to unmarshal a struct.
func (g *Generator) StructUnmarshaler(t *types.Struct) (code string, err error) {
	fields, err := g.StructFields(t, nil)
	if err != nil {
		return "", err
	}
	code = fmt.Sprintf(`
if !n.IsObject() {
	return n.TypeError(njson.TypeObject)
}
for k := n.Value(); k != nil; k = k.Next() {
	n := k.Value()
	switch k.Unescaped() {`)
	for name, field := range fields {
		body, err := g.TypeUnmarshaler(field.Type())
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf(`
	case %s:
		%s{
			r := &r%s
			%s
		}
		`, fmt.Sprintf("`\"%s\"`", name), g.EnsurePath(field.Path), field.Path, body)
	}
	code += `
	}
}`

	return

}

// CanUnmarshal returns if can be unmarshaled
func CanUnmarshal(t types.Type) bool {
	tt := ResolveType(t)
	if tt == nil {
		return false
	}
	switch typ := tt.(type) {
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
		println("Unknown type: ", typ)
		return false
	}

}

type typeError struct {
	typ types.Type
}

func (e typeError) Error() string {
	return fmt.Sprintf("Unsupported type %s %#v", e.typ, e.typ)
}

// TypeUnmarshaler returns the code block for unmarshaling a type.
func (g *Generator) TypeUnmarshaler(t types.Type) (code string, err error) {
	typ := ResolveType(t)
	switch typ := typ.(type) {
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
		return "", typeError{t}
	default:
		return "", typeError{t}
	}
}

// AllStructs returns all structs from the package.
func (g *Generator) AllStructs() (all []string) {
	for _, def := range g.info.Defs {
		if def == nil {
			continue
		}
		typ := def.Type()
		if typ == nil {
			continue
		}
		if typ, ok := typ.(*types.Named); ok {
			obj := typ.Obj()
			if obj == nil || obj.Pkg() != g.pkg {
				continue
			}
			if t := ResolveType(typ); t != nil {
				if _, ok := t.(*types.Struct); ok {
					all = append(all, typ.Obj().Name())
				}
			}
		}
	}
	return

}

// UnmarshalMethodName is the default name for the unmarshal function
const UnmarshalMethodName = "UnmarshalNodeJSON"

// UnmarshalMethodName returns the name for the unmarshal method.
func (g *Generator) UnmarshalMethodName() (m string) {
	m = UnmarshalMethodName
	switch g.TagKey() {
	case "", DefaultTagKey:
	default:
		m += strcase.ToCamel(g.TagKey())
	}
	return m
}

// WriteUnmarshaler writes an unmarshaler method for a type in the generator's buffer.
func (g *Generator) WriteUnmarshaler(typeName string) (err error) {
	_, code, err := g.Unmarshaler(typeName)
	if err != nil {
		return
	}
	g.Import(njsonPkg)
	_, err = g.buffer.WriteString(code)
	return
}

// Unmarshaler generates an unmarshaler method for a type
func (g *Generator) Unmarshaler(typeName string) (typ *types.Named, code string, err error) {
	typ = g.LookupType(typeName)
	if typ == nil {
		return nil, "", fmt.Errorf("Type %s not found", typeName)
	}
	receiverName := strings.ToLower(typeName[:1])
	method := g.UnmarshalMethodName()
	body, err := g.TypeUnmarshaler(typ)
	if err != nil {
		return typ, "", err
	}
	code = fmt.Sprintf(`
		func (%[1]s *%[2]s) %[3]s(n *njson.Node) error {
			if !n.IsValue() {
				return n.TypeError(njson.TypeAnyValue)
			}
			r := %[1]s
			%[4]s
			return nil
		}
	`, receiverName, typeName, method, body)

	return
}

var injectPackages = map[string]string{
	"encoding/json": "encoding/json",
}

func newGenerator(path, targetPkg string) (g *Generator, err error) {
	fset := token.NewFileSet()
	mode := parser.ParseComments | parser.DeclarationErrors
	astPkgs, err := parser.ParseDir(fset, path, filterTestFiles, mode)
	if err != nil {
		return nil, err
	}
	pkg := astPkgs[targetPkg]
	if pkg == nil {
		return nil, fmt.Errorf("Target package %s not found in path", targetPkg)
	}
	g = new(Generator)
	for _, f := range pkg.Files {
		if !isGeneratedByNJSON(f) {
			g.files = append(g.files, f)
		}
	}
	g.logger = log.New(ioutil.Discard, "", 0)
	g.info = types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}
	config := types.Config{
		IgnoreFuncBodies: true,
		FakeImportC:      true,
		Importer:         importer.Default(),
	}
	g.pkg, err = config.Check(pkg.Name, fset, g.files, &g.info)
	if err != nil {
		return nil, err
	}
	return
}

func filterTestFiles(f os.FileInfo) bool {
	return !strings.HasSuffix(f.Name(), "_test.go")
}

// New creates a new Generator for a package named targetPkg and parses the specified path.
func New(path string, targetPkg string, options ...Option) (*Generator, error) {
	g, err := newGenerator(path, targetPkg)
	if err != nil {
		return nil, err
	}
	for _, opt := range options {
		opt(g)
	}
	return g, nil
}

// Reset resets the generator to start a new file.
func (g *Generator) Reset() {
	g.buffer.Reset()
	g.imports = nil
}

func inject(fset *token.FileSet, target, pkg string) (*ast.File, error) {
	src := fmt.Sprintf(`package %s
	import _ %q
	`, target, pkg)
	filename := fmt.Sprintf("njson/inject/%s.go", pkg)
	return parser.ParseFile(fset, filename, src, 0)
}

// Import adds packages to import in the generated file
func (g *Generator) Import(imports ...*types.Package) {
	if len(imports) == 0 {
		return
	}
	if g.imports == nil {
		g.imports = make(map[string]*types.Package, len(imports))
	}
	for _, pkg := range imports {
		g.imports[pkg.Path()] = pkg
	}
	return
}

// DumpTo writes the generated file without checking and formatting.
func (g *Generator) DumpTo(w io.Writer) error {
	if _, err := w.Write([]byte(g.Header())); err != nil {
		return err
	}
	if _, err := g.buffer.WriteTo(w); err != nil {
		return err
	}
	return nil
}

// PrintTo writes the generated file after checking and formatting.
func (g *Generator) PrintTo(w io.Writer) error {
	fset := token.NewFileSet()
	buf := new(bytes.Buffer)
	buf.WriteString(g.Header())
	g.buffer.WriteTo(buf)
	filename := fmt.Sprintf("%s_njson.go", g.pkg.Name())
	astFile, err := parser.ParseFile(fset, filename, buf.Bytes(), parser.ParseComments)
	if err != nil {
		return err
	}
	return printer.Fprint(w, fset, astFile)
}

const (
	njsonPkgPath = "github.com/alxarch/njson"
	njsonPkgName = "njson"
)

var njsonPkg = types.NewPackage(njsonPkgPath, njsonPkgName)

const headerComment = `// Code generated by njson; DO NOT EDIT.`

func isGeneratedByNJSON(f *ast.File) bool {
	return len(f.Comments) > 0 && len(f.Comments[0].List) > 0 && f.Comments[0].List[0].Text == headerComment
}

// Header returns the header code for the generated file.
func (g *Generator) Header() string {
	h := []string{}
	h = append(h, headerComment)
	h = append(h, fmt.Sprintf("package %s", g.pkg.Name()))

	for path, pkg := range g.imports {
		if filepath.Base(path) == pkg.Name() {
			h = append(h, fmt.Sprintf("import %q", path))
		} else {
			h = append(h, fmt.Sprintf("import %s %q", pkg.Name(), path))
		}
	}
	return strings.Join(h, "\n")

}

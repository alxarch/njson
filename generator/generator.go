package generator

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
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/alxarch/njson"
)

const DefaultTagKey = "json"

func ParseFieldTag(tag, key string) (name string, omitempty, ok bool) {
	if key == "" {
		key = DefaultTagKey
	}
	tag, ok = reflect.StructTag(tag).Lookup(key)
	if !ok {
		return
	}
	name = tag
	if ok = name != "-"; !ok {
		return
	}
	if i := strings.IndexByte(tag, ','); i > -1 {
		name = tag[:i]
		omitempty = strings.Index(tag[i:], "omitempty") > 0
	}
	return
}

type Generator struct {
	options
	pkg     *types.Package
	buffer  bytes.Buffer
	imports map[string]*types.Package
	files   []*ast.File
	info    types.Info
}

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

type StructField struct {
	*types.Var
	NameJSON  string
	OmitEmpty bool
	Path      FieldPath
}

func (f *StructField) String() string {
	return f.Var.String() + " " + f.Path.String()
}

type StructFields map[string]StructField

func (fields StructFields) Add(f *types.Var, name string, omitempty bool, path FieldPath) {
	name = string(njson.EscapeString(nil, name))
	_, duplicate := fields[name]
	if duplicate && ComparePaths(fields[name].Path, path) == -1 {
		// keep existing
		return
	}
	fields[name] = StructField{f, name, omitempty, path.Copy()}
}

type FieldIndex struct {
	Index int
	Type  types.Type
	Name  string
}

func (i FieldIndex) String() string {
	return i.Name
}

type FieldPath []FieldIndex

func TypeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Pointer:
		return TypeName(t.Elem())
	default:
		return t.String()
	}
}
func (g *Generator) TypeName(t types.Type) string {
	name, pkg := g.ResolveTypeName(t)
	if pkg != nil {
		g.Import(pkg)
	}
	return name
}
func (g *Generator) ResolveTypeName(t types.Type) (string, *types.Package) {
	switch t := t.(type) {
	case *types.Named:
		pkg := t.Obj().Pkg()
		if pkg == g.pkg {
			return t.Obj().Name(), nil
		}
		return t.String(), pkg
	case *types.Pointer:
		return g.ResolveTypeName(t.Elem())
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

func (p FieldPath) Copy() FieldPath {
	cp := make([]FieldIndex, len(p))
	copy(cp, p)
	return cp
}

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
			name = g.JSONFieldName(TypeName(t))
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

func (g *Generator) InterfaceUnmarshaler(t types.Type, b *types.Interface) (code string, err error) {
	return `if x, ok := n.ToInterface(); ok { *r = x } else { return n.TypeError(njson.AnyValue) }`, nil
}

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

func (g *Generator) MapUnmarshaller(t types.Type, m *types.Map) (code string, err error) {
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

func (g *Generator) TypeUnmarshaler(t types.Type) (code string, err error) {
	typ := ResolveType(t)
	switch typ := typ.(type) {
	case *types.Map:
		return g.MapUnmarshaller(t, typ)
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

const UnmarshalMethodName = "UnmarshalNodeJSON"

func (g *Generator) UnmarshalMethodName() (m string) {
	m = UnmarshalMethodName
	switch g.TagKey() {
	case "", DefaultTagKey:
	default:
		m += strcase.ToCamel(g.TagKey())
	}
	return m
}

func (g *Generator) WriteUnmarshaler(typeName string) (err error) {
	_, code, err := g.Unmarshaler(typeName)
	if err != nil {
		return
	}
	g.Import(njsonPkg)
	_, err = g.buffer.WriteString(code)
	return
}

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

func (g *Generator) DumpTo(w io.Writer) error {
	if err := g.WriteHeaderTo(w); err != nil {
		return err
	}
	if _, err := g.buffer.WriteTo(w); err != nil {
		return err
	}
	return nil
}

func (g *Generator) WriteFormattedTo(w io.Writer) error {
	fset := token.NewFileSet()
	buf := new(bytes.Buffer)
	g.WriteHeaderTo(buf)
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

func (g *Generator) WriteHeaderTo(w io.Writer) (err error) {
	fmt.Fprintln(w, headerComment)
	fmt.Fprintf(w, "package %s\n", g.pkg.Name())
	for path, pkg := range g.imports {
		if filepath.Base(path) == pkg.Name() {
			fmt.Fprintf(w, "import %q\n", path)
		} else {
			fmt.Fprintf(w, "import %s %q\n", pkg.Name(), path)
		}
	}
	return
}

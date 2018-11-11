package generator

/*
Package generator is a code generator for njson.Unmarsaler

It parses a package dir and can generate (t *T)UnmarshalNodeJSON(*njson.Node) error methods.
*/

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alxarch/meta"
)

// TODO: handle flag combinations (ie if tag key is not json don't use UnmarshalJSON)

// Generator is a source code generator for njson unmarshal methods.
type Generator struct {
	*meta.Package
	options
	test    bool
	buffer  bytes.Buffer
	imports map[string]*types.Package
}

type typeError struct {
	typ types.Type
}

func (e typeError) Error() string {
	return fmt.Sprintf("Unsupported type %s %#v", e.typ, e.typ)
}

// AllStructs returns all structs from the package.
func (g *Generator) AllStructs() (all []string) {
	types := g.Package.DefinedTypes(meta.IsStruct)
	for _, typ := range types {
		all = append(all, g.TypeString(typ))
	}
	return
}

// NewFromFile creates a new Generator for a package named targetPkg and parses the specified file.
func NewFromFile(filename string, src interface{}, options ...Option) (*Generator, error) {
	p := meta.NewParser(parser.ParseComments | parser.DeclarationErrors)
	name, err := p.ParseFile(filename, src)
	if err != nil {
		return nil, err
	}
	pkg, err := p.Package(name, name, func(f *ast.File) bool {
		return !IsGeneratedByNJSON(f)
	})
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	g.Package = pkg
	g.logger = log.New(ioutil.Discard, "", 0)
	for _, opt := range options {
		opt(g)
	}
	return g, nil
}

// NewFromDir creates a new Generator for a package named targetPkg and parses the specified path.
func NewFromDir(path, name string, options ...Option) (*Generator, error) {
	var filter func(os.FileInfo) bool
	if !strings.HasSuffix(name, "_test") {
		filter = meta.IgnoreTestFiles
	}
	p := meta.NewParser(parser.ParseComments | parser.DeclarationErrors)
	if err := p.ParseDir(path, filter); err != nil {
		return nil, err
	}
	pkg, err := p.Package(name, name, func(f *ast.File) bool {
		return !IsGeneratedByNJSON(f)
	})
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	g.Package = pkg
	g.logger = log.New(ioutil.Discard, "", 0)
	for _, opt := range options {
		opt(g)
	}
	return g, nil
}

// Filename returns the filename to write output.
func (g *Generator) Filename() (name string) {
	name = g.Name()
	if strings.HasSuffix(name, "_test") {
		name = strings.TrimSuffix(name, "_test")
		name = name + "_njson_test.go"
		return
	}
	name = name + "_njson.go"
	return

}

// Reset resets the generator to start a new file.
func (g *Generator) Reset() {
	g.buffer.Reset()
	g.imports = nil
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
		if pkg.Path() != g.Path() {
			g.imports[pkg.Path()] = pkg
		}
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
	filename := g.Filename()
	astFile, err := parser.ParseFile(fset, filename, buf.Bytes(), parser.ParseComments)
	if err != nil {
		return err
	}
	ast.SortImports(fset, astFile)
	return format.Node(w, fset, astFile)
}

const (
	njsonPkgPath = "github.com/alxarch/njson"
	njsonPkgName = "njson"
)

var (
	njsonPkg                = meta.MustImport(njsonPkgPath)
	typNodeJSONUnmarshaler  = njsonPkg.Scope().Lookup("Unmarshaler").Type().Underlying().(*types.Interface)
	methodNodeUnmarshalJSON = typNodeJSONUnmarshaler.Method(0)
	typJSONAppender         = njsonPkg.Scope().Lookup("Appender").Type().Underlying().(*types.Interface)
	methodAppendJSON        = typJSONAppender.Method(0)
	typNode                 = njsonPkg.Scope().Lookup("Node").Type()
	typNodePtr              = types.NewPointer(typNode)

	unjsonPkg  = meta.MustImport(njsonPkgPath + "/unjson")
	typOmiter  = unjsonPkg.Scope().Lookup("Omiter").Type().Underlying().(*types.Interface)
	methodOmit = typOmiter.Method(0)

	strjsonPkg = meta.MustImport(njsonPkgPath + "/strjson")
	numjsonPkg = meta.MustImport(njsonPkgPath + "/numjson")

	jsonPkg            = meta.MustImport("encoding/json")
	typJSONUnmarshaler = jsonPkg.Scope().Lookup("Unmarshaler").Type().Underlying().(*types.Interface)
	typJSONMarshaler   = jsonPkg.Scope().Lookup("Marshaler").Type().Underlying().(*types.Interface)

	encodingPkg        = meta.MustImport("encoding")
	typTextUnmarshaler = encodingPkg.Scope().Lookup("TextUnmarshaler").Type().Underlying().(*types.Interface)
	typTextMarshaler   = encodingPkg.Scope().Lookup("TextMarshaler").Type().Underlying().(*types.Interface)

	strconvPkg = meta.MustImport("strconv")
)

const (
	headerComment = `// Code generated by njson on %s; DO NOT EDIT.`
)

// IsGeneratedByNJSON checks if a file begins with the njson generated header comment.
func IsGeneratedByNJSON(f *ast.File) bool {
	return len(f.Comments) > 0 && strings.HasPrefix(f.Comments[0].Text(), "Code generated by njson")
}

// Header returns the header code for the generated file.
func (g *Generator) Header() string {
	h := []string{}
	now := time.Now().In(time.UTC)
	ts := now.Format(time.RFC1123)
	h = append(h, fmt.Sprintf(headerComment, ts))
	h = append(h, fmt.Sprintf("package %s", g.Name()))

	for path, pkg := range g.imports {
		if filepath.Base(path) == pkg.Name() {
			h = append(h, fmt.Sprintf("import %q", path))
		} else {
			h = append(h, fmt.Sprintf("import %s %q", pkg.Name(), path))
		}
	}
	return strings.Join(h, "\n")

}

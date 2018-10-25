package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"flag"

	"github.com/alxarch/njson/generator"
)

var (
	targetPath   = flag.String("p", ".", "Path to scan for .go files.")
	tagKey       = flag.String("k", "json", "Struct tag key.")
	onlyTagged   = flag.Bool("only-tagged", false, "Only use tagged fields.")
	onlyExported = flag.Bool("only-exported", false, "Only use exported fields.")
	// allStructs     = flag.Bool("all", false, "Generate methods for all defined structs in package.")
	forceOmitEmpty    = flag.Bool("omitempty", false, "Force omitempty on all fields.")
	generateAppend    = flag.Bool("append", false, "Generate AppendJSON methods.")
	generateUnmarshal = flag.Bool("unmarshal", false, "Generate UnmarshalNodeJSON methods.")
	matchFieldNames   = flag.String("match", ".*", "Regex for filtering by field name.")
	caseTransform     = flag.String("case", "none", "Field name case transformation for untagged fields (none|snake|camel|lower).")
	writeFile         = flag.Bool("w", false, `Write output to a file named "{pkgname}_njson.go".`)
	debug             = flag.Bool("d", false, "Debug mode.")
	// tests           = flag.Bool("tests", false, "Write test methods.")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [pkg] [type...]:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	logger := log.New(os.Stderr, "njson: ", log.LstdFlags)
	targetPkg := flag.Arg(0)
	if targetPkg == "" {
		absPath, err := filepath.Abs(*targetPath)
		if err != nil {
			logger.Fatalf("Failed to resolve path %q: %s", *targetPath, err)
		}
		targetPkg = path.Base(absPath)
		targetPkg = strings.TrimSuffix(targetPkg, path.Ext(targetPkg))
	}
	var options []generator.Option
	options = append(options, generator.Logger(logger))
	options = append(options, generator.TagKey(*tagKey))
	if *forceOmitEmpty {
		options = append(options, generator.ForceOmitEmpty(true))
	}
	if *onlyExported {
		options = append(options, generator.OnlyExported(true))
	}
	if *onlyTagged {
		options = append(options, generator.OnlyTagged(true))
	}
	if *caseTransform != "none" {
		options = append(options, generator.TransformFieldCase(*caseTransform))
	}
	if *matchFieldNames != ".*" {
		rx, err := regexp.Compile(*matchFieldNames)
		if err != nil {
			log.Fatalf("Invalid match field name regexp: %s", err)
		}
		options = append(options, generator.MatchFieldName(rx))
	}
	if *generateAppend == *generateUnmarshal {
		*generateAppend = true
		*generateUnmarshal = true
	}

	g, err := generator.NewFromDir(*targetPath, targetPkg, options...)
	if err != nil {
		logger.Fatal(err)
	}

	typeFilter := func(string) bool { return true }
	switch match := flag.Arg(1); match {
	case "", "*":
	default:
		if regexp.QuoteMeta(match) == match {
			// match is not regexp
			typeFilter = func(name string) bool { return name == match }
		} else if rx, err := regexp.Compile(match); err != nil {
			logger.Fatalln("Failed to compile match regex", err)
		} else {
			typeFilter = rx.MatchString
		}
	}
	types := []string{}
	for _, t := range g.AllStructs() {
		if typeFilter(t) {
			types = append(types, t)
		}
	}

	if len(types) == 0 {
		logger.Fatalf("No types found")
	}

	g.Reset()
	var t *generator.Generator
	// if *tests {
	// 	t = g.Test()
	// 	t.Reset()
	// }

	for _, typ := range types {
		if *generateUnmarshal {
			if err := g.WriteUnmarshaler(typ); err != nil {
				logger.Fatal(err)
			}
		}
		if *generateAppend {
			if err := g.WriteAppender(typ); err != nil {
				logger.Fatal(err)
			}
		}
		// if t != nil {
		// 	if err := t.WriteUnmarshalerTest(typ); err != nil {
		// 		logger.Fatal(err)
		// 	}
		// }
	}
	if *debug {
		g.DumpTo(os.Stderr)
		if t != nil {
			t.DumpTo(os.Stderr)
		}
		return
	}

	var out io.Writer = os.Stdout
	// var tout io.Writer = os.Stderr
	if *writeFile {
		filename := g.Filename()
		filename = filepath.Join(*targetPath, filename)
		f, err := os.Create(filename)
		defer f.Close()
		if err != nil {
			logger.Fatalf("Failed to open file %q for writing: %s", filename, err)
		}
		out = f
		// if t != nil && t != g {
		// 	filename := t.Filename()
		// 	filename = filepath.Join(*targetPath, filename)
		// 	f, err := os.Create(filename)
		// 	defer f.Close()
		// 	if err != nil {
		// 		logger.Fatalf("Failed to open file %q for writing: %s", filename, err)
		// 	}
		// 	tout = f
		// }

	}
	if err := g.PrintTo(out); err != nil {
		logger.Fatalf("Failed to write output: %s", err)
	}
	// if t != nil && t != g {
	// 	if err := t.PrintTo(tout); err != nil {
	// 		logger.Fatalf("Failed to write tests: %s", err)

	// 	}
	// }

}

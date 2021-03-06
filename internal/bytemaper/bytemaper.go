package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"text/template"
)

var (
	pkgName  = flag.String("pkg", "main", "Package name to use")
	fileName = flag.String("w", "", "Write output to file")
	prefix   = flag.String("prefix", "bytemap", "Var name prefix")
)

func trIs(fn func(rune) bool) func(byte) byte {
	return func(b byte) byte {
		if fn(rune(b)) {
			return 1
		}
		return 0
	}
}
func trTo(fn func(rune) rune) func(byte) byte {
	return func(b byte) byte {
		r := fn(rune(b))
		if utf8.RuneLen(r) == 1 {
			return byte(r)
		}
		return 0
	}
}

func nameOrDefault(name, defaultName string) string {
	if name == "" {
		name = defaultName
	}
	return name
}

func toISO3166(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c
	}
	if 'a' <= c && c <= 'z' {
		return byte(unicode.ToUpper(rune(c)))
	}
	return 'Z'
}

func toISO639(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return byte(unicode.ToLower(rune(c)))
	}
	if 'a' <= c && c <= 'z' {
		return c
	}
	return 'z'
}

func toJSON(c byte) byte {
	if c < utf8.RuneSelf {
		switch c {
		case '<', '>', '&':
			return 1 // HTML Unsafe
		case '\\', '/', '"':
			return '\\'
		case '\r':
			return 'r'
		case '\n':
			return 'n'
		case '\t':
			return 't'
		case '\f':
			return 'f'
		case '\b':
			return 'b'
		}
		if unicode.IsControl(rune(c)) {
			return 0 // Control character
		}
		return utf8.RuneSelf
	}
	return 0xff
}
func toHex(c byte) byte {
	switch {
	case 0 <= c && c <= 9:
		return c + '0'
	case 10 <= c && c <= 15:
		return c + 'a' - 10
	default:
		return 0xff
	}
}

func fromHex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'z':
		return 10 + c - 'a'
	case 'A' <= c && c <= 'Z':
		return 10 + c - 'A'
	default:
		return 0xff
	}
}
func main() {
	flag.Parse()
	w := os.Stdout
	if *fileName != "" {
		f, err := os.Create(*fileName)
		if err != nil {
			log.Fatalln("Failed to open file for writing", err)
		}
		w = f
	}
	masks := map[string]string{}
	var tr func(b byte) byte
	for _, arg := range flag.Args() {
		name := arg
		if i := strings.IndexByte(arg, ':'); i != -1 {
			name = arg[:i]
			if arg = arg[i+1:]; len(arg) >= 2 && arg[0] != '"' && arg[len(arg)-1] != '"' {
				arg = "\"" + arg + "\""
			}
			arg, _ = strconv.Unquote(arg)
		} else {
			arg = ""
		}
		switch name {
		case "ToJSON":
			tr = toJSON
			name = nameOrDefault(arg, name)
		case "ToISO3166":
			tr = toISO3166
			name = nameOrDefault(arg, name)
		case "ToISO639":
			tr = toISO639
			name = nameOrDefault(arg, name)
		case "ToHex":
			tr = toHex
			name = nameOrDefault(arg, name)
		case "FromHex":
			tr = fromHex
			name = nameOrDefault(arg, name)
		case "IsControl":
			tr = trIs(unicode.IsControl)
			name = nameOrDefault(arg, name)
		case "IsPunct":
			tr = trIs(unicode.IsPunct)
			name = nameOrDefault(arg, name)
		case "IsDigit":
			tr = trIs(unicode.IsDigit)
			name = nameOrDefault(arg, name)
		case "IsSpace":
			tr = trIs(unicode.IsSpace)
			name = nameOrDefault(arg, name)
		case "IsPrint":
			tr = trIs(unicode.IsPrint)
			name = nameOrDefault(arg, name)
		case "ToLower":
			tr = trTo(unicode.ToLower)
			name = nameOrDefault(arg, name)
		case "ToUpper":
			tr = trTo(unicode.ToUpper)
			name = nameOrDefault(arg, name)
		case "ToTitle":
			tr = trTo(unicode.ToTitle)
			name = nameOrDefault(arg, name)
		default:
			if name == "" {
				log.Fatalf("Invalid arg %q\n", arg)
			}
			tr = func(b byte) byte {
				if strings.IndexByte(arg, b) != -1 {
					return 1
				}
				return 0
			}

		}
		mask := make([]byte, 256)
		for i := 0; i < len(mask); i++ {
			mask[i] = tr(byte(i))
		}
		masks[name] = string(mask)
	}
	if err := tpl.Execute(w, map[string]interface{}{
		"PkgName": *pkgName,
		"Maps":    masks,
		"Prefix":  *prefix,
	}); err != nil {
		log.Fatalln("Failed to write file", err)
	}

}

var funcs = template.FuncMap{
	"hasprefix": strings.HasPrefix,
}
var tpl = template.Must(template.New("masks").Funcs(funcs).Parse(`// Code generated by bytemaper; DO NOT EDIT.
package {{.PkgName}}
{{- $prefix := .Prefix }}

const (
{{- range $name, $map := .Maps }}
	{{$prefix}}{{$name}} = {{ printf "%q" $map }}
{{- end }}
)

{{- range $name, $map := .Maps }}
{{- if hasprefix $name "Is" }}

func {{$name}}(c byte) bool {
	return {{$prefix}}{{$name}}[c] == 1
}
{{- else }}

func {{$name}}(c byte) byte {
	return {{$prefix}}{{$name}}[c]
}
{{- end }}
{{- end }}
`))

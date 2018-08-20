package generator

import (
	"go/types"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

type options struct {
	tagKey         string
	onlyExported   bool
	onlyTagged     bool
	forceOmitEmpty bool
	matchField     func(string) bool
	fieldName      func(string) string
	logger         *log.Logger
}
type Option func(g *Generator)

func (o *options) JSONFieldName(name string) string {
	if o.fieldName == nil {
		return name
	}
	return o.fieldName(name)
}
func TransformFieldCase(mode string) Option {
	var fieldNamer func(string) string
	switch mode {
	case "snake":
		fieldNamer = (strcase.ToSnake)
	case "lower":
		fieldNamer = (strings.ToLower)
	case "camel":
		fieldNamer = (strcase.ToLowerCamel)
	case "Camel":
		fieldNamer = (strcase.ToCamel)
	}
	return func(g *Generator) {
		g.fieldName = fieldNamer
	}
}

var matchAll = regexp.MustCompile(".*")

func MatchFieldName(rx *regexp.Regexp) Option {
	if rx == nil {
		rx = matchAll
	}
	return func(g *Generator) {
		g.matchField = rx.MatchString
	}
}

func OnlyTagged(on bool) Option {
	return func(g *Generator) {
		g.onlyTagged = on
	}
}
func OnlyExported(on bool) Option {
	return func(g *Generator) {
		g.onlyExported = on
	}
}
func (o *options) TagKey() string {
	if o.tagKey == "" {
		return DefaultTagKey
	}
	return o.tagKey
}
func (o *options) parseField(field *types.Var, tag string) (name string, omitempty, tagged, skip bool) {
	if skip = !o.MatchField(field.Name()); skip {
		return
	}
	if skip = o.onlyExported && !field.Exported(); skip {
		return
	}
	name, omitempty, tagged = ParseFieldTag(tag, o.TagKey())
	if skip = o.onlyTagged && !tagged; skip {
		return
	}
	if skip = name == "-"; skip {
		return
	}
	if name == "" {
		name = o.JSONFieldName(field.Name())
	}
	if o.forceOmitEmpty {
		omitempty = true
	}
	return
}
func (o *options) MatchField(name string) bool {
	if name == "_" {
		return false
	}
	if o.matchField == nil {
		return true
	}
	return o.matchField(name)
}
func TagKey(key string) Option {
	return func(g *Generator) {
		g.tagKey = key
	}
}

func Logger(logger *log.Logger) Option {
	if logger == nil {
		logger = log.New(ioutil.Discard, "", 0)
	}
	return func(g *Generator) {
		g.logger = logger
	}
}

func ForceOmitEmpty(on bool) Option {
	return func(g *Generator) {
		g.forceOmitEmpty = on
	}
}

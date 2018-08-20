package generator

import (
	"go/types"
	"io/ioutil"
	"log"
	"reflect"
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

// Option is a generator option
type Option func(g *Generator)

func (o *options) JSONFieldName(name string) string {
	if o.fieldName == nil {
		return name
	}
	return o.fieldName(name)
}

// TransformFieldCase sets a case transformation mode for field names when no tag based name is found.
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

// MatchFieldName sets a regex for matching struct field names
func MatchFieldName(rx *regexp.Regexp) Option {
	if rx == nil {
		rx = matchAll
	}
	return func(g *Generator) {
		g.matchField = rx.MatchString
	}
}

// OnlyTagged forces generator to ignore fields without a tag
func OnlyTagged(on bool) Option {
	return func(g *Generator) {
		g.onlyTagged = on
	}
}

// OnlyExported forces generator to ignore unexported fields
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

// TagKey sets the tag key to use when parsing struct fields
func TagKey(key string) Option {
	return func(g *Generator) {
		g.tagKey = key
	}
}

// Logger sets a logger for the generator error messages.
func Logger(logger *log.Logger) Option {
	if logger == nil {
		logger = log.New(ioutil.Discard, "", 0)
	}
	return func(g *Generator) {
		g.logger = logger
	}
}

// ForceOmitEmpty forces omitempty on all fields regardless of json tag.
func ForceOmitEmpty(on bool) Option {
	return func(g *Generator) {
		g.forceOmitEmpty = on
	}
}

// DefaultTagKey is the default tag key to use when parsing stuct fields.
const DefaultTagKey = "json"

// ParseFieldTag parses a field tag to get a json name and omitempty info.
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

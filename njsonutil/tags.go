package njsonutil

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

import "github.com/iancoleman/strcase"

func TaggedMethodName(methodName string, tag string) string {
	if tag != "" && tag != "json" {
		methodName += "Tag" + strcase.ToCamel(tag)
	}
	return methodName
}

type Tag struct {
	Name    string
	Missing bool
	Params  Params
}

type Params url.Values

func (p Params) Values() url.Values {
	return url.Values(p)
}

func (p Params) Defaults(other ...Params) Params {
	if p == nil {
		p = make(map[string][]string)
	}
	for i := 1; i < len(other); i++ {
		for k, v := range other[i] {
			if len(p[k]) == 0 && len(v) != 0 {
				p[k] = v
			}
		}
	}
	return p
}

func (p Params) Assign(other ...Params) Params {
	if p == nil {
		p = make(map[string][]string)
	}

	for i := 1; i < len(other); i++ {
		for k, v := range other[i] {
			if len(v) != 0 {
				p[k] = v
			}
		}
	}
	return p
}

func (p Params) Pull(key string) (v string) {
	if len(p[key]) > 0 {
		v = p[key][0]
		p[key] = p[key][1:]
	}
	return
}

func (p Params) Pop(key string) (v string) {
	if n := len(p[key]) - 1; n >= 0 {
		v = p[key][n]
		p[key] = p[key][:n]
	}
	return
}

func (p Params) Get(key string) string {
	return p.Values().Get(key)
}

func (p Params) Has(key string) bool {
	return len(p[key]) > 0
}

func (p Params) True(key string) (b bool) {
	if p.Get(key) == key {
		return true
	}
	b, _ = p.ToBool(key)
	return
}

func (p Params) ToBool(key string) (bool, error) {
	return strconv.ParseBool(p.Get(key))
}

func (p Params) ToDuration(key string) (time.Duration, error) {
	return time.ParseDuration(p.Get(key))
}

func (p Params) Time(key, format string) (t time.Time) {
	t, _ = p.ToTime(p.Get(key), format)
	return
}

func (p Params) ToTime(key, format string) (time.Time, error) {
	return time.Parse(p.Get(key), format)
}

func (p Params) Duration(key string) (d time.Duration) {
	d, _ = p.ToDuration(key)
	return
}

func (p Params) Int(key string) (i int) {
	i, _ = p.ToInt(key)
	return
}

func (p Params) ToInt(key string) (i int, err error) {
	return strconv.Atoi(p.Get(key))
}

func (p Params) Float(key string) (f float64) {
	f, _ = p.ToFloat(key)
	return
}

func (p Params) ToFloat(key string) (float64, error) {
	return strconv.ParseFloat(p.Get(key), 64)
}

func ParseTag(tag, key string) (t Tag) {
	name, ok := LookupTag(tag, key)
	if t.Missing = !ok; t.Missing {
		return
	}
	t.Name = name
	if i := strings.IndexByte(name, ','); i != -1 {
		t.Name = tag[:i]
		tag = tag[i+1:]
	} else {
		return
	}
	for len(tag) > 0 {
		params := url.Values{}
		if i := strings.IndexByte(tag, '='); i != -1 {
			k := tag[:i]
			tag = tag[i+1:]
			if j := strings.IndexByte(tag, ','); j != -1 {
				params.Add(k, tag[:j])
				tag = tag[j+1:]
			} else {
				params.Add(k, k)
			}
		} else if i := strings.IndexByte(tag, ','); i != -1 {
			k := tag[:i]
			tag = tag[i+1:]
			params.Add(k, k)
		} else {
			params.Add(tag, tag)
			tag = ""
		}
		if len(params) > 0 {
			t.Params = Params(params)
		}

	}
	return

}

func LookupTag(tag, key string) (string, bool) {
	return reflect.StructTag(tag).Lookup(key)
}

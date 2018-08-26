package genjson

import (
	"fmt"
	"go/types"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/alxarch/meta"
	"github.com/alxarch/njson/strjson"
)

var r = rand.NewSource(time.Now().UnixNano())

func GenerateJSONMap(t *types.Map, params meta.Params) (out string) {
	size := r.Int63() % 16
	out += "{"
	for i := int64(0); i < size; i++ {
		k := GenerateJSON(t.Key(), params)
		v := GenerateJSON(t.Elem(), params)
		if i > 0 {
			out += ","
		}
		out += k + ":" + v
	}
	out += "}"
	return
}

const (
	tagKey         = "genjson"
	strNull        = "null"
	paramOmit      = "omit"
	paramSize      = "size"
	paramMax       = "max"
	paramMin       = "min"
	paramTrue      = "true"
	paramFalse     = "false"
	paramPrecision = "precision"
	paramValue     = "value"
)

func GenerateJSONStruct(t *types.Struct, params meta.Params) (out string) {
	if params.Pop(strNull) == strNull {
		return strNull
	}
	return "{" + GenerateJSONStructFields(t, params) + "}"
}

func GenerateJSONStructFields(t *types.Struct, params meta.Params) (out string) {
	for i := 0; i < t.NumFields(); i++ {
		field := t.Field(i)
		if field.Name() == "_" {
			continue
		}
		tag, _ := meta.ParseTag(t.Tag(i), tagKey)
		tag.Params = tag.Params.Defaults(params)
		if tag.Params.Pop(paramOmit) == paramOmit {
			continue
		}
		if tag.Name == "" {
			if s, ok := meta.Embedded(field); ok {
				out += GenerateJSONStructFields(s, tag.Params)
				continue
			}
			tag.Name = field.Name()
			if tag.Name == "" {
				tag.Name = meta.Base(field.Type())
			}
		}
		if value := GenerateJSON(field.Type(), tag.Params); value != "" {
			if i > 0 {
				out += ","
			}
			out += fmt.Sprintf(`"%s":%s`, tag.Name, value)
		}
	}
	return
}

func GenerateJSONSlice(t *types.Slice, params meta.Params) (out string) {
	if params.Pop(strNull) == strNull {
		return strNull
	}
	size, err := strconv.Atoi(params.Pop(paramSize))
	if err != nil {
		size = int(rand.Int31n(16))
	}
	if size == 0 {
		return `[]`
	}
	out += "["
	for i := 0; i < size; i++ {
		if i > 0 {
			out += ","
		}
		out += GenerateJSON(t.Elem(), params)
	}
	out += "]"
	return
}

func GenerateJSONPointer(t *types.Pointer, params meta.Params) (out string) {
	if params.Pop(strNull) == strNull {
		return strNull
	}
	return GenerateJSON(t.Elem(), params)
}

func GenerateJSONValue(t *types.Basic, params meta.Params) (out string) {
	if out = params.Get("value"); out != "" {
		if t.Kind() == types.String {
			out = strconv.Quote(out)
		}
		return
	}
	switch t.Kind() {
	case types.String:
		size, err := strconv.Atoi(params.Pop(paramSize))
		if err != nil {
			size = rand.Intn(64)
		}
		data := make([]byte, size)
		rand.Read(data)
		s := strjson.Quoted(nil, string(data))
		return string(s)
	case types.Bool:
		return "true"
	default:
		if _, ok := meta.BasicInfo(t, types.IsNumeric); ok {
			min, err := params.ToFloat(paramMin)
			if err != nil {
				min = 500 - 1000*rand.Float64()
			}
			max, err := params.ToFloat(paramMax)
			if err != nil {
				max = min + 1000*rand.Float64()
			}
			var num float64
			switch {
			case min == max:
				num = min
			case min > max:
				max, min = min, max
				fallthrough
			default:
				num = min + rand.Float64()*(max-min)
			}
			if _, ok := meta.BasicInfo(t, types.IsFloat); ok {
				p := params.Int(paramPrecision)
				if p <= 0 {
					p = 6
				}
				return strconv.FormatFloat(num, 'f', p, 64)
			}
			if -1 < num && num < 1 {
				num *= 100
			}
			num = math.Trunc(num)
			if _, ok := meta.BasicInfo(t, types.IsUnsigned); ok {
				if num < 0 {
					num -= num
				}
			}
			return strconv.Itoa(int(num))
		}

	}
	return

}

func GenerateJSONAnyValue(params meta.Params) (out string) {
	return fmt.Sprintf("%d", rand.Int())
}

func GenerateJSON(t types.Type, params meta.Params) (out string) {
	if t = meta.Resolve(t); t == nil {
		return
	}
	switch typ := t.(type) {
	case *types.Map:
		return GenerateJSONMap(typ, params)
	case *types.Struct:
		return GenerateJSONStruct(typ, params)
	case *types.Slice:
		return GenerateJSONSlice(typ, params)
	case *types.Pointer:
		return GenerateJSONPointer(typ, params)
	case *types.Basic:
		return GenerateJSONValue(typ, params)
	case *types.Interface:
		if typ.Empty() {
			return GenerateJSONAnyValue(params)
		}
	}
	return ""
}

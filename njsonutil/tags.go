package njsonutil

import "github.com/iancoleman/strcase"

func TaggedMethodName(methodName string, tag string) string {
	if tag != "" && tag != "json" {
		methodName += "Tag" + strcase.ToCamel(tag)
	}
	return methodName
}

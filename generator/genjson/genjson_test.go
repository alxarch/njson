package genjson_test

import (
	"go/types"
	"testing"

	"github.com/alxarch/njson/generator/genjson"
)

func TestJSON(t *testing.T) {
	strukt := types.NewStruct([]*types.Var{
		types.NewVar(0, nil, "Foo", types.Typ[types.String]),
	}, nil)
	json := genjson.GenerateJSON(strukt, nil)
	t.Errorf(json)
}

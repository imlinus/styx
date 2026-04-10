package stdlib

import (
	"os"
	"strings"
	"styx/pkg/object"
)

func GetEnvModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// get(key)
	modulePairs[(&object.String{Value: "get"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "get"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "env.get expects 1 string argument"}
				}
				key := args[0].(*object.String).Value
				val, exists := os.LookupEnv(key)
				if !exists {
					return &object.Null{}
				}
				return &object.String{Value: val}
			},
		},
	}

	// all()
	modulePairs[(&object.String{Value: "all"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "all"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				pairs := make(map[string]object.HashPair)
				for _, e := range os.Environ() {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						keyObj := &object.String{Value: parts[0]}
						valObj := &object.String{Value: parts[1]}

						hashKey := keyObj.HashKey()
						pairs[hashKey] = object.HashPair{Key: keyObj, Value: valObj}
					}
				}
				return &object.Hash{Pairs: pairs}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

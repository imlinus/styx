package stdlib

import (
	"encoding/json"
	"styx/pkg/object"
)

func GetJsonModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// encode(obj)
	modulePairs[(&object.String{Value: "encode"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "encode"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Error{Message: "json.encode expects 1 argument"}
				}
				return &object.String{Value: object.ToJSON(args[0])}
			},
		},
	}

	// decode(str)
	modulePairs[(&object.String{Value: "decode"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "decode"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "json.decode expects 1 string argument"}
				}
				str := args[0].(*object.String).Value
				var data interface{}
				if err := json.Unmarshal([]byte(str), &data); err != nil {
					return &object.Error{Message: err.Error()}
				}
				return object.NativeToStyxObject(data)
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

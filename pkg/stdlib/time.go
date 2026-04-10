package stdlib

import (
	"styx/pkg/object"
	"time"
)

func GetTimeModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// now()
	modulePairs[(&object.String{Value: "now"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "now"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				return &object.Integer{Value: time.Now().Unix()}
			},
		},
	}

	// iso()
	modulePairs[(&object.String{Value: "iso"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "iso"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				return &object.String{Value: time.Now().UTC().Format(time.RFC3339)}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

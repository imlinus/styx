package stdlib

import (
	"fmt"
	"styx/pkg/object"
	"time"
)

func GetLogModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// info(...)
	modulePairs[(&object.String{Value: "info"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "info"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				var msg string
				for i, arg := range args {
					if arg.Type() == object.HASH_OBJ {
						msg += object.ToJSON(arg)
					} else {
						msg += arg.Inspect()
					}
					if i < len(args)-1 {
						msg += " "
					}
				}
				fmt.Printf("[%s] [INFO] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
				return &object.Null{}
			},
		},
	}

	// error(...)
	modulePairs[(&object.String{Value: "error"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "error"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				var msg string
				for i, arg := range args {
					if arg.Type() == object.HASH_OBJ {
						msg += object.ToJSON(arg)
					} else {
						msg += arg.Inspect()
					}
					if i < len(args)-1 {
						msg += " "
					}
				}
				fmt.Printf("[%s] [ERROR] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
				return &object.Null{}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

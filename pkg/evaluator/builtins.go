package evaluator

import (
	"fmt"
	"io"
	"styx/pkg/object"
)

var builtins = map[string]*object.Builtin{
	"load": {
		Fn: func(args ...object.Object) object.Object { return nil },
	},
	"delete": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return &object.Error{Message: "delete expects 2 arguments: hash, key"}
			}
			hash, ok := args[0].(*object.Hash)
			if !ok {
				return &object.Error{Message: "delete expects a hash as first argument"}
			}
			key := args[1].HashKey()
			if key == "" {
				return &object.Error{Message: "unusable as hash key"}
			}
			delete(hash.Pairs, key)
			return &object.Null{}
		},
	},
	"print": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Error{Message: "len expects 1 argument"}
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Hash:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				return &object.Error{Message: "len not supported for " + string(arg.Type())}
			}
		},
	},
}

func builtinPrint(w interface{}, args ...object.Object) object.Object {
	writer, ok := w.(io.Writer)
	if !ok {
		// Fallback to stdout
		for _, arg := range args {
			fmt.Println(arg.Inspect())
		}
		return NULL
	}

	for i, arg := range args {
		var output string
		if arg.Type() == object.HASH_OBJ {
			output = object.ToJSON(arg)
		} else {
			output = arg.Inspect()
		}

		io.WriteString(writer, output)
		if i < len(args)-1 {
			io.WriteString(writer, " ")
		}
	}
	io.WriteString(writer, "\n")
	return NULL
}

// Replaces the PHP class Database

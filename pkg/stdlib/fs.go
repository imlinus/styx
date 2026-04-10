package stdlib

import (
	"os"
	"path/filepath"
	"styx/pkg/object"
)

func GetFsModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// read(path)
	modulePairs[(&object.String{Value: "read"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "read"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "fs.read expects 1 string argument"}
				}
				path := args[0].(*object.String).Value
				content, err := os.ReadFile(path)
				if err != nil {
					return &object.Error{Message: "fs.read error: " + err.Error()}
				}
				return &object.String{Value: string(content)}
			},
		},
	}

	// write(path, content)
	modulePairs[(&object.String{Value: "write"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "write"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ {
					return &object.Error{Message: "fs.write expects 2 string arguments"}
				}
				path := args[0].(*object.String).Value
				content := args[1].(*object.String).Value
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return &object.Error{Message: "fs.write mkdir error: " + err.Error()}
				}

				err := os.WriteFile(path, []byte(content), 0644)
				if err != nil {
					return &object.Error{Message: "fs.write error: " + err.Error()}
				}
				return &object.Boolean{Value: true}
			},
		},
	}

	// exists(path)
	modulePairs[(&object.String{Value: "exists"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "exists"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "fs.exists expects 1 string argument"}
				}
				path := args[0].(*object.String).Value
				_, err := os.Stat(path)
				return &object.Boolean{Value: err == nil}
			},
		},
	}

	// is_dir(path)
	modulePairs[(&object.String{Value: "is_dir"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "is_dir"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "fs.is_dir expects 1 string argument"}
				}
				path := args[0].(*object.String).Value
				info, err := os.Stat(path)
				return &object.Boolean{Value: err == nil && info.IsDir()}
			},
		},
	}

	// ls(path)
	modulePairs[(&object.String{Value: "ls"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "ls"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "fs.ls expects 1 string argument"}
				}
				path := args[0].(*object.String).Value
				entries, err := os.ReadDir(path)
				if err != nil {
					return &object.Error{Message: "fs.ls error: " + err.Error()}
				}

				resultPairs := make(map[string]object.HashPair)
				for i, entry := range entries {
					key := &object.Integer{Value: int64(i)}
					resultPairs[key.HashKey()] = object.HashPair{Key: key, Value: &object.String{Value: entry.Name()}}
				}
				return &object.Hash{Pairs: resultPairs}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

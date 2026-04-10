package stdlib

import (
	"strings"
	"styx/pkg/object"
)

func GetStringModule() *object.Hash {
	modulePairs := make(map[string]object.HashPair)

	// replace(s, old, new)
	modulePairs[(&object.String{Value: "replace"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "replace"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ || args[2].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.replace expects 3 string arguments"}
				}
				s := args[0].(*object.String).Value
				old := args[1].(*object.String).Value
				new := args[2].(*object.String).Value
				return &object.String{Value: strings.ReplaceAll(s, old, new)}
			},
		},
	}

	// sub(s, start, [end])
	modulePairs[(&object.String{Value: "sub"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "sub"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.INTEGER_OBJ {
					return &object.Error{Message: "string.sub expects a string and at least 1 integer"}
				}
				s := args[0].(*object.String).Value
				start := int(args[1].(*object.Integer).Value)

				// Handle negative start (Lua style: -1 is last char)
				if start < 0 {
					start = len(s) + start
				}

				end := len(s)
				if len(args) == 3 && args[2].Type() == object.INTEGER_OBJ {
					end = int(args[2].(*object.Integer).Value)
					if end < 0 {
						end = len(s) + end + 1
					}
				}

				if start < 0 {
					start = 0
				}
				if end > len(s) {
					end = len(s)
				}
				if start > end {
					return &object.String{Value: ""}
				}

				return &object.String{Value: s[start:end]}
			},
		},
	}

	// lower(s)
	modulePairs[(&object.String{Value: "lower"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "lower"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.lower expects 1 string argument"}
				}
				return &object.String{Value: strings.ToLower(args[0].(*object.String).Value)}
			},
		},
	}

	// slug(s) - Basic slugify: non-alphanumeric -> dash
	modulePairs[(&object.String{Value: "slug"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "slug"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.slug expects 1 string argument"}
				}
				s := args[0].(*object.String).Value
				var out strings.Builder
				lastWasDash := false
				for _, r := range strings.ToLower(s) {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
						out.WriteRune(r)
						lastWasDash = false
					} else if !lastWasDash {
						out.WriteRune('-')
						lastWasDash = true
					}
				}
				return &object.String{Value: strings.Trim(out.String(), "-")}
			},
		},
	}

	// find(s, pattern)
	modulePairs[(&object.String{Value: "find"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "find"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.find expects 2 string arguments"}
				}
				s := args[0].(*object.String).Value
				pattern := args[1].(*object.String).Value
				index := strings.Index(s, pattern)
				if index == -1 {
					return &object.Null{}
				}
				return &object.Integer{Value: int64(index)}
			},
		},
	}

	// split(s, sep)
	modulePairs[(&object.String{Value: "split"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "split"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.split expects 2 string arguments"}
				}
				s := args[0].(*object.String).Value
				sep := args[1].(*object.String).Value
				parts := strings.Split(s, sep)

				resultPairs := make(map[string]object.HashPair)
				for i, part := range parts {
					key := &object.Integer{Value: int64(i)}
					resultPairs[key.HashKey()] = object.HashPair{Key: key, Value: &object.String{Value: part}}
				}
				return &object.Hash{Pairs: resultPairs}
			},
		},
	}

	// trim(s, [cutset])
	modulePairs[(&object.String{Value: "trim"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "trim"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) < 1 || args[0].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.trim expects at least 1 string argument"}
				}
				s := args[0].(*object.String).Value
				cutset := " \t\n\r"
				if len(args) == 2 && args[1].Type() == object.STRING_OBJ {
					cutset = args[1].(*object.String).Value
				}
				return &object.String{Value: strings.Trim(s, cutset)}
			},
		},
	}

	// starts_with(s, prefix)
	modulePairs[(&object.String{Value: "starts_with"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "starts_with"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.starts_with expects 2 string arguments"}
				}
				return &object.Boolean{Value: strings.HasPrefix(args[0].(*object.String).Value, args[1].(*object.String).Value)}
			},
		},
	}

	// ends_with(s, suffix)
	modulePairs[(&object.String{Value: "ends_with"}).HashKey()] = object.HashPair{
		Key: &object.String{Value: "ends_with"},
		Value: &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 || args[0].Type() != object.STRING_OBJ || args[1].Type() != object.STRING_OBJ {
					return &object.Error{Message: "string.ends_with expects 2 string arguments"}
				}
				return &object.Boolean{Value: strings.HasSuffix(args[0].(*object.String).Value, args[1].(*object.String).Value)}
			},
		},
	}

	return &object.Hash{Pairs: modulePairs}
}

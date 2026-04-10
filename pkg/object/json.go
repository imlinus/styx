package object

import (
	"encoding/json"
	"fmt"
)

func ToJSON(obj Object) string {
	native := ToNative(obj)
	bytes, err := json.Marshal(native)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return "{}"
	}
	return string(bytes)
}

func ToNative(obj Object) interface{} {
	switch o := obj.(type) {
	case *Integer:
		return o.Value
	case *Boolean:
		return o.Value
	case *String:
		return o.Value
	case *Null:
		return nil
	case *Hash:
		m := make(map[string]interface{})
		for _, pair := range o.Pairs {
			m[pair.Key.Inspect()] = ToNative(pair.Value)
		}
		return m
	default:
		return nil
	}
}

func NativeToStyxObject(val interface{}) Object {
	switch v := val.(type) {
	case string:
		return &String{Value: v}
	case float64:
		return &Integer{Value: int64(v)}
	case bool:
		return &Boolean{Value: v}
	case map[string]interface{}:
		pairs := make(map[string]HashPair)
		for k, val := range v {
			pairs[(&String{Value: k}).HashKey()] = HashPair{
				Key:   &String{Value: k},
				Value: NativeToStyxObject(val),
			}
		}
		return &Hash{Pairs: pairs}
	default:
		return &Null{}
	}
}

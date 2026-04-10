package object

import (
	"bytes"
	"fmt"
	"strings"
	"styx/pkg/ast"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	HASH_OBJ         = "HASH"
	ARRAY_OBJ        = "ARRAY"
	BREAK_OBJ        = "BREAK"
	CONTINUE_OBJ     = "CONTINUE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
	HashKey() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) HashKey() string  { return fmt.Sprintf("%s:%d", i.Type(), i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) HashKey() string  { return fmt.Sprintf("%s:%t", b.Type(), b.Value) }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return s.Value }
func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) HashKey() string  { return fmt.Sprintf("%s:%s", s.Type(), s.Value) }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) HashKey() string  { return string(n.Type()) }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) HashKey() string  { return "" }

type Error struct {
	Message string
}

func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) HashKey() string  { return "" }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("function")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) HashKey() string  { return "" }

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[string]HashPair
}

func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) HashKey() string  { return "" }

// BuiltinFunction defines a builtin function signature
type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) Type() ObjectType { return "BUILTIN" }
func (b *Builtin) HashKey() string  { return "" }

type Array struct {
	Elements []Object
}

func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) HashKey() string  { return "" }

type Break struct{}

func (b *Break) Inspect() string  { return "break" }
func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) HashKey() string  { return "" }

type Continue struct{}

func (c *Continue) Inspect() string  { return "continue" }
func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) HashKey() string  { return "" }

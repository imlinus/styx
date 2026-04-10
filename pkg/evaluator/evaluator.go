package evaluator

import (
	"fmt"
	"path/filepath"
	"styx/pkg/ast"
	"styx/pkg/lexer"
	"styx/pkg/object"
	"styx/pkg/parser"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

type Evaluator struct {
	Env        *object.Environment
	W          interface{} // ResponseWriter for server mode
	ScriptPath string      // Path to current script for relative imports
	ReadFile   func(string) ([]byte, error)
}

func New(env *object.Environment) *Evaluator {
	return &Evaluator{Env: env}
}

func (e *Evaluator) Eval(node ast.Node) object.Object {
	if node == nil {
		return nil
	}

	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return e.evalProgram(node)

	case *ast.ExpressionStatement:
		return e.Eval(node.Expression)

	case *ast.BlockStatement:
		return e.evalBlockStatement(node)

	case *ast.ReturnStatement:
		val := e.Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := e.Eval(node.Value)
		if isError(val) {
			return val
		}
		return e.Env.Define(node.Name.Value, val)

	case *ast.LoopStatement:
		return e.evalLoopStatement(node)

	case *ast.BreakStatement:
		return &object.Break{}

	case *ast.ContinueStatement:
		return &object.Continue{}

	case *ast.ImportStatement:
		if node.Path == nil {
			// Legacy stdlib import (already injected by server)
			return NULL
		}

		// Enhanced import: import {Name} from {Path}
		importPath := node.Path.Value
		if !filepath.IsAbs(importPath) && e.ScriptPath != "" {
			importPath = filepath.Join(filepath.Dir(e.ScriptPath), importPath)
		}

		// Use the builtinLoad logic but with a fresh environment
		if e.ReadFile == nil {
			return newError("%s", "ReadFile not implemented")
		}
		content, err := e.ReadFile(importPath)
		if err != nil {
			return newError("import error: %s", err.Error())
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			return newError("import parse error: %s", p.Errors()[0])
		}

		// Evaluate in a fresh environment that has access to globals
		importEnv := object.NewEnclosedEnvironment(e.Env)
		importEvaluator := &Evaluator{Env: importEnv, W: e.W, ScriptPath: importPath, ReadFile: e.ReadFile}
		importEvaluator.Eval(program)

		if node.Name != nil {
			// Export specific symbol
			if val, ok := importEnv.Get(node.Name.Value); ok {
				e.Env.Define(node.Name.Value, val)
			} else {
				return newError("symbol '%s' not found in %s", node.Name.Value, importPath)
			}
		} else {
			// Export all symbols (could be an option in the future)
			// For now, only named imports are fully supported by 'import ... from'
		}

		return NULL

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := e.Eval(node.Right)
		if isError(right) {
			return right
		}
		return e.evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := e.Eval(node.Left)
		if isError(left) {
			return left
		}

		right := e.Eval(node.Right)
		if isError(right) {
			return right
		}

		return e.evalInfixExpression(node.Operator, left, right)

	case *ast.AssignmentExpression:
		val := e.Eval(node.Value)
		if isError(val) {
			return val
		}

		if indexExp, ok := node.Left.(*ast.IndexExpression); ok {
			left := e.Eval(indexExp.Left)
			if isError(left) {
				return left
			}

			// For obj.key, the index identifier should be treated as a string
			var index object.Object
			if indexExp.Token.Type == lexer.DOT {
				if ident, ok := indexExp.Index.(*ast.Identifier); ok {
					index = &object.String{Value: ident.Value}
				} else {
					index = e.Eval(indexExp.Index)
				}
			} else {
				index = e.Eval(indexExp.Index)
			}

			if isError(index) {
				return index
			}

			return evalIndexAssignment(left, index, val)
		}

		return e.Env.Set(node.Left.String(), val)

	case *ast.IndexExpression:
		left := e.Eval(node.Left)
		if isError(left) {
			return left
		}

		var index object.Object
		if node.Token.Type == lexer.DOT {
			if ident, ok := node.Index.(*ast.Identifier); ok {
				index = &object.String{Value: ident.Value}
			} else {
				index = e.Eval(node.Index)
			}
		} else {
			index = e.Eval(node.Index)
		}

		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	case *ast.IfExpression:
		return e.evalIfExpression(node)

	case *ast.Identifier:
		return e.evalIdentifier(node)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: e.Env, Body: body}

	case *ast.CallExpression:
		function := e.Eval(node.Function)
		if isError(function) {
			return function
		}
		args := e.evalExpressions(node.Arguments)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return e.applyFunction(function, args)

	case *ast.HashLiteral:
		return e.evalHashLiteral(node)

	case *ast.Null:
		return NULL

	case *ast.TemplateLiteral:
		return e.evalTemplateLiteral(node)

	case *ast.ArrayLiteral:
		elements := e.evalExpressions(node.Elements)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	}

	return nil
}

func (e *Evaluator) evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = e.Eval(statement)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = e.Eval(statement)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ || rt == object.BREAK_OBJ || rt == object.CONTINUE_OBJ {
				return result
			}
		}
	}

	return result
}

func (e *Evaluator) evalLoopStatement(ls *ast.LoopStatement) object.Object {
	var result object.Object = NULL

	// Scenario 1: Iteration Loop (loop k, v in hash)
	if ls.Key != nil {
		itered := e.Eval(ls.Iter)
		if isError(itered) {
			return itered
		}

		if hash, ok := itered.(*object.Hash); ok {
			for _, pair := range hash.Pairs {
				oldEnv := e.Env
				e.Env = object.NewEnclosedEnvironment(oldEnv)
				e.Env.Define(ls.Key.Value, pair.Key)
				if ls.Value != nil {
					e.Env.Define(ls.Value.Value, pair.Value)
				}
				result = e.Eval(ls.Body)
				e.Env = oldEnv

				if result != nil {
					if result.Type() == object.BREAK_OBJ {
						return NULL
					}
					if result.Type() == object.CONTINUE_OBJ {
						continue
					}
					if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
						return result
					}
				}
			}
		} else if array, ok := itered.(*object.Array); ok {
			for i, val := range array.Elements {
				oldEnv := e.Env
				e.Env = object.NewEnclosedEnvironment(oldEnv)
				e.Env.Define(ls.Key.Value, &object.Integer{Value: int64(i)})
				if ls.Value != nil {
					e.Env.Define(ls.Value.Value, val)
				}
				result = e.Eval(ls.Body)
				e.Env = oldEnv

				if result != nil {
					if result.Type() == object.BREAK_OBJ {
						return NULL
					}
					if result.Type() == object.CONTINUE_OBJ {
						continue
					}
					if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
						return result
					}
				}
			}
		} else {
			return newError("loop..in expects a hash or array, got %s", itered.Type())
		}
		return NULL
	}

	// Scenario 2: Condition / Infinite Loop
	for {
		if ls.Condition != nil {
			condition := e.Eval(ls.Condition)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		result = e.Eval(ls.Body)
		if result != nil {
			if result.Type() == object.BREAK_OBJ {
				break
			}
			if result.Type() == object.CONTINUE_OBJ {
				continue
			}
			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result
			}
		}
	}

	return NULL
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "not":
		return e.evalNotOperatorExpression(right)
	case "-":
		return e.evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func (e *Evaluator) evalNotOperatorExpression(right object.Object) object.Object {
	if isTruthy(right) {
		return FALSE
	}
	return TRUE
}

func (e *Evaluator) evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch operator {
	case "and":
		if isTruthy(left) {
			return right
		}
		return left
	case "or":
		if isTruthy(left) {
			return left
		}
		return right
	case "..":
		return &object.String{Value: left.Inspect() + right.Inspect()}
	}

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "and":
		if isTruthy(left) {
			return right
		}
		return left
	case "or":
		if isTruthy(left) {
			return left
		}
		return right
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "..":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIfExpression(ie *ast.IfExpression) object.Object {
	condition := e.Eval(ie.Condition)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return e.Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return e.Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier) object.Object {
	if val, ok := e.Env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		// Inject ResponseWriter and Evaluator into builtins if needed
		return &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if node.Value == "print" && e.W != nil {
					return builtinPrint(e.W, args...)
				}
				if node.Value == "load" {
					return e.builtinLoad(args...)
				}
				return builtin.Fn(args...)
			},
		}
	}

	return newError("%s", "identifier not found: "+node.Value)
}

func (e *Evaluator) evalExpressions(exps []ast.Expression) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaluated := e.Eval(exp)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func (e *Evaluator) builtinLoad(args ...object.Object) object.Object {
	if len(args) != 1 || args[0].Type() != object.STRING_OBJ {
		return newError("load expects 1 string argument")
	}

	if e.ReadFile == nil {
		return newError("%s", "ReadFile not implemented")
	}
	path := args[0].(*object.String).Value
	content, err := e.ReadFile(path)
	if err != nil {
		return newError("load error: %s", err.Error())
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return newError("load parse error: %s", p.Errors()[0])
	}

	return e.Eval(program)
}

func (e *Evaluator) evalHashLiteral(node *ast.HashLiteral) object.Object {
	pairs := make(map[string]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		var key object.Object
		if ident, ok := keyNode.(*ast.Identifier); ok {
			key = &object.String{Value: ident.Value}
		} else {
			key = e.Eval(keyNode)
			if isError(key) {
				return key
			}
		}

		hashKey := key.HashKey()
		if hashKey == "" {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := e.Eval(valueNode)
		if isError(value) {
			return value
		}

		pairs[hashKey] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	case left.Type() == object.ARRAY_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("array index must be an integer, got %s", index.Type())
	}

	max := int64(len(arrayObject.Elements) - 1)
	if idx.Value < 0 || idx.Value > max {
		return NULL
	}

	return arrayObject.Elements[idx.Value]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	hashKey := index.HashKey()
	if hashKey == "" {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[hashKey]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalIndexAssignment(left, index, val object.Object) object.Object {
	switch {
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexAssignment(left, index, val)
	case left.Type() == object.ARRAY_OBJ:
		return evalArrayIndexAssignment(left, index, val)
	default:
		return newError("index assignment not supported: %s", left.Type())
	}
}

func evalHashIndexAssignment(left, index, val object.Object) object.Object {
	hashObject := left.(*object.Hash)
	hashKey := index.HashKey()
	if hashKey == "" {
		return newError("unusable as hash key: %s", index.Type())
	}

	hashObject.Pairs[hashKey] = object.HashPair{Key: index, Value: val}

	return val
}

func evalArrayIndexAssignment(left, index, val object.Object) object.Object {
	arrayObject := left.(*object.Array)
	idx, ok := index.(*object.Integer)
	if !ok {
		return newError("array index must be an integer, got %s", index.Type())
	}

	max := int64(len(arrayObject.Elements) - 1)
	if idx.Value < 0 || idx.Value > max {
		return newError("index out of bounds: %d", idx.Value)
	}

	arrayObject.Elements[idx.Value] = val
	return val
}

func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	if fn == nil {
		return newError("not a function: NIL")
	}

	switch fn := fn.(type) {

	case *object.Function:
		extendEnv := e.extendFunctionEnv(fn, args)
		evaluator := &Evaluator{Env: extendEnv, W: e.W}
		evaluated := evaluator.Eval(fn.Body)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func (e *Evaluator) extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Define(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Null:
		return false
	case *object.Boolean:
		return obj.Value
	default:
		return true
	}
}

func (e *Evaluator) evalTemplateLiteral(node *ast.TemplateLiteral) object.Object {
	var result string

	for _, part := range node.Parts {
		evaluated := e.Eval(part)
		if isError(evaluated) {
			return evaluated
		}
		result += evaluated.Inspect()
	}

	return &object.String{Value: result}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// evalGlobal is used for the compatibility with the old API if needed.
func EvalGlobal(node ast.Node, env *object.Environment) object.Object {
	e := &Evaluator{Env: env}
	return e.Eval(node)
}

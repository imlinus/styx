package object

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object)}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	// Walk up scopes to find where this variable was defined
	if _, ok := e.store[name]; !ok && e.outer != nil {
		if _, ok := e.outer.store[name]; ok {
			return e.outer.Set(name, val)
		}
	}

	e.store[name] = val
	return val
}

func (e *Environment) Define(name string, val Object) Object {
	e.store[name] = val
	return val
}

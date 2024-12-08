package object

type (
	// Environment is used to bind values to identifiers.
	Environment struct {
		// store maps identifiers to objects
		store map[string]Object
		outer *Environment
	}
)

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (env *Environment) Get(ident string) (Object, bool) {
	// lookup the inner scope
	obj, ok := env.store[ident]
	// maybe the identifier is in outer scope
	if !ok && env.outer != nil {
		obj, ok = env.outer.Get(ident)
	}
	return obj, ok
}

func (env *Environment) Set(ident string, obj Object) Object {
	env.store[ident] = obj
	return obj
}

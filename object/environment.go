package object

type (
	// Environment is used to bind values to identifiers.
	Environment struct {
		// store maps identifiers to objects
		store map[string]Object
	}
)

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (env *Environment) Get(ident string) (Object, bool) {
	obj, ok := env.store[ident]
	return obj, ok
}

func (env *Environment) Set(ident string, obj Object) Object {
	env.store[ident] = obj
	return obj
}

package symbols

type Environment struct {
	store map[string][]Symbol
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string][]Symbol)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (env *Environment) Set(ident string, symbol Symbol) {
	env.store[ident] = append(env.store[ident], symbol)
}

func (env *Environment) Get(ident string) ([]Symbol, bool) {
	// check current env
	if symbols, ok := env.store[ident]; ok {
		return symbols, true
	}

	// check outer scope
	if env.outer != nil {
		if symbols, ok := env.outer.Get(ident); ok {
			return symbols, true
		}
	}

	return nil, false
}

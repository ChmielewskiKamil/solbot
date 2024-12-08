package symbols

type Environment struct {
	store map[string]Symbol
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Symbol)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (env *Environment) Set(ident string, symbol Symbol) {
	env.store[ident] = symbol
}

func (env *Environment) Get(ident string) (Symbol, bool) {
	symbol, ok := env.store[ident]
	if !ok && env.outer != nil {
		symbol, ok = env.outer.Get(ident)
	}
	return symbol, ok
}

package symbols

import "fmt"

type Environment struct {
	store        map[string][]Symbol     // Mapping between symbol's name and a struct holding all info about that symbol.
	outer        *Environment            // Access to outer env for symbol lookups.
	inner        *Environment            // Access to inner envs such as contract env which contains functions which themselves have inner envs.
	perSymbolEnv map[Symbol]*Environment // With access to symbol's env, lookups like this are possible: for each contract in a file, find all public state vars.
}

func NewEnvironment() *Environment {
	s := make(map[string][]Symbol)
	e := make(map[Symbol]*Environment)
	return &Environment{store: s, outer: nil, perSymbolEnv: e}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	// The new inner scope
	env := NewEnvironment()
	// Outer of the inner scope
	env.outer = outer
	// Set the new env as inner of the old one
	outer.inner = env

	return env
}

func (env *Environment) Set(ident string, symbol Symbol) {
	env.store[ident] = append(env.store[ident], symbol)
	env.perSymbolEnv[symbol] = env
}

// Get looks up a symbol in the env provided the identifier. It returns an array
// of symbols since during the symbol discovery (phase 1) there could have been
// many symbols with the same name/signature (e.g. fn overrides) which must be
// resolved in the resolution phase (phase 2). Get returns false if the symbol
// is not found.
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

// GetInnerEnvOfSymbol returns the inner environment of the queried symbol.
// If there is no such environment it returns an error. Can be used like this:
//
// 1. Get all symbols of contract type in the most outer env.
//
// 2. Use this function (GetInnerEnvOfSymbol) to get access to env of
// particular contract.
//
// 3. With this you can access for example all functions or state variables.
func (env *Environment) GetInnerEnvOfSymbol(s Symbol) (error, *Environment) {
	e := env.perSymbolEnv[s]
	if e == nil {
		return fmt.Errorf("The symbol's env is nil. Symbol is located at: %s", s.Location()), nil
	}

	if e.inner == nil {
		return fmt.Errorf("The symbol's inner env is nil. Symbol is located at: %s", s.Location()), nil
	}

	return nil, e.inner
}

// Returns an array of all the symbols with a specific symbol type from an environment.
func GetAllSymbolsByType[T any](env *Environment) []T {
	var results []T

	for _, symbols := range env.store {
		for _, symbol := range symbols {
			if typedSymbol, ok := symbol.(T); ok {
				results = append(results, typedSymbol)
			}
		}
	}
	return results
}

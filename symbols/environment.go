package symbols

import "fmt"

type Environment struct {
	store     map[string][]Symbol // Mapping between symbol's name and a struct holding all info about that symbol.
	outer     *Environment        // Access to outer env for symbol lookups. Can be nil.
	inner     *Environment        // Access to inner envs such as contract env which contains functions which themselves have inner envs. Can be nil.
	scopeName string              // Name of the env scope; FileName.sol for files; contract name for contracts etc.
	scopeType ReferenceScopeType  // Type of the scope to be used for references
}

func NewEnvironment(scopeName string, scopeType ReferenceScopeType) *Environment {
	s := make(map[string][]Symbol)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment,
	scopeName string, scopeType ReferenceScopeType) *Environment {
	// The new inner scope
	env := NewEnvironment(scopeName, scopeType)
	// Outer of the inner scope
	env.outer = outer
	// Set the new env as inner of the old one
	outer.inner = env

	return env
}

func (env *Environment) Set(ident string, symbol Symbol) {
	env.store[ident] = append(env.store[ident], symbol)
	symbol.SetOuterEnv(env)
}

// TODO: The comment below might no longer be correct. There is no resolution
// phase except for reference resolution. Each lookup must check that the returned
// symbol from the array matches the expected type + param types in case of functions
// with the same name and different accepted param types.
//
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

func (env *Environment) GetCurrentScopeName() string {
	return env.scopeName
}

func (env *Environment) GetCurrentScopeType() ReferenceScopeType {
	return env.scopeType
}

func GetInnerEnv(s Symbol) (*Environment, error) {
	e := s.GetInnerEnv()
	if e == nil {
		return nil, fmt.Errorf("The symbol does not have INNER env set (it's nil). Symbol is located at: %s", s.Location())
	}

	return e, nil
}

func GetOuterEnv(s Symbol) (*Environment, error) {
	e := s.GetOuterEnv()
	if e == nil {
		return nil, fmt.Errorf("The symbol does not have OUTER env set (it's nil). Symbol is located at: %s", s.Location())
	}

	return e, nil
}

// Returns an array of all the symbols with a specific symbol type from an environment.
// This can return 0 elements if nothing was found. WARNING: check returned array's length.
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

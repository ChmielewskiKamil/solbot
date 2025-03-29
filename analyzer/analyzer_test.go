package analyzer

import (
	"reflect"
	"solbot/symbols"
	"testing"
)

func Test_DiscoverSymbols_getSymbolsByType(t *testing.T) {
	testContractPath := "testdata/foundry/src/003_SimpleCounter_WithEvents.sol"
	analyzer := Analyzer{}
	analyzer.Init(testContractPath)
	checkParserErrors(t, &analyzer)
	checkAnalyzerErrors(t, &analyzer)

	analyzer.AnalyzeCurrentFile()

	env := analyzer.GetCurrentFileEnv()

	if env == nil {
		t.Fatalf("Currently analyzed file's env is nil.")
	}

	expectedFileDecls := []struct {
		declName   string
		symbolType symbols.Symbol
	}{
		{"OutsideOfContract", &symbols.Event{}},
		{"Counter", &symbols.Contract{}},
	}

	// Iterate through expected top-level symbols and check their types
	for _, expected := range expectedFileDecls {
		sym, found := env.Get(expected.declName)
		if !found {
			t.Fatalf("Symbol: '%s' not found.", expected.declName)
		}

		// Type assertion to ensure the symbol is of the expected type
		if reflect.TypeOf(sym[0]) != reflect.TypeOf(expected.symbolType) {
			t.Fatalf("Symbol '%s' has unexpected type. Got: %T, Expected: %T", expected.declName, sym, expected.symbolType)
		}
	}

	expectedContractDecls := []struct {
		declName   string
		symbolType symbols.Symbol
	}{
		{"InsideOfContract", &symbols.Event{}},
		{"count", &symbols.StateVariable{}},
		{"increment", &symbols.Function{}},
		{"decrement", &symbols.Function{}},
		{"reset", &symbols.Function{}},
	}

	// Iterate through all contracts discovered in the file
	contracts := symbols.GetAllSymbolsByType[*symbols.Contract](env)
	if len(contracts) != 1 {
		t.Fatalf("Expected %d contracts, but found %d.", 1, len(contracts))
	}

	for _, contract := range contracts {
		if contract.Name != expectedFileDecls[1].declName {
			t.Fatalf("Unexpected contract name: '%s'. Expected: '%s'.", contract.Name, expectedFileDecls[0].declName)
		}

		// Retrieve the environment specific to this contract
		err, contractEnv := env.GetInnerEnvOfSymbol(contract)
		if err != nil {
			t.Fatalf("Cannot access contract's env: %s", err)
		}

		if contractEnv == nil {
			t.Fatalf("Environment for contract '%s' is nil.", contract.Name)
		}

		// Check for declarations within the contract environment
		for _, decl := range expectedContractDecls {
			sym, found := contractEnv.Get(decl.declName)
			if !found {
				t.Fatalf("Symbol: '%s' not found in contract '%s'.", decl.declName, contract.Name)
			}

			// Type assertion to ensure the symbol is of the expected type
			if reflect.TypeOf(sym[0]) != reflect.TypeOf(decl.symbolType) {
				t.Fatalf("Symbol '%s' has unexpected type. Got: %T, Expected: %T", decl.declName, sym[0], decl.symbolType)
			}
		}
	}
}

func checkParserErrors(t *testing.T, a *Analyzer) {
	errors := a.GetParserErrors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("Parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("Parser error: %s", err.Msg)
	}
	t.FailNow()
}

func checkAnalyzerErrors(t *testing.T, a *Analyzer) {
	errors := a.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("Analyzer has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("Analyzer error: %s At location: %s", err.Msg, err.Loc)
	}
	t.FailNow()
}

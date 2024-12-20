package analyzer

import (
	"solbot/symbols"
	"testing"
)

func Test_DiscoverSymbols_GetSymbolsByName(t *testing.T) {
	testContractPath := "testdata/foundry/src/002_SimpleCounter.sol"
	analyzer := Analyzer{}
	analyzer.Init(testContractPath)
	checkParserErrors(t, &analyzer)

	analyzer.AnalyzeCurrentFile()

	env := analyzer.GetCurrentFileEnv()
	println("Current file env: ", env)
	if env == nil {
		t.Fatalf("Currently analyzed file's env is nil.")
	}

	expectedFileDecls := []struct {
		declName string
	}{
		{"Counter"},
	}

	_, found := env.Get(expectedFileDecls[0].declName)
	if !found {
		t.Fatalf("Symbol: '%s' not found.", expectedFileDecls[0].declName)
	}

	expectedContractDecls := []struct {
		declName string
	}{
		{"count"},
		{"increment"},
		{"decrement"},
		{"reset"},
	}

	// Iterate through all contracts discovered in the file
	contracts := symbols.GetAllSymbolsByType[*symbols.Contract](env)
	if len(contracts) != len(expectedFileDecls) {
		t.Fatalf("Expected %d contracts, but found %d.", len(expectedFileDecls), len(contracts))
	}

	for _, contract := range contracts {
		if contract.Name != expectedFileDecls[0].declName {
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
			_, found := contractEnv.Get(decl.declName)
			if !found {
				t.Fatalf("Symbol: '%s' not found in contract '%s'.", decl.declName, contract.Name)
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

package analyzer

import (
	"github.com/ChmielewskiKamil/solbot/symbols"
	"reflect"
	"testing"
)

func Test_DiscoverSymbols_getSymbolsByType(t *testing.T) {
	testContractPath := "testdata/foundry/src/003_SimpleCounter_WithEvents.sol"
	analyzer := Analyzer{}
	analyzer.Init(testContractPath)

	analyzer.AnalyzeCurrentFile()

	checkAnalyzerErrors(t, &analyzer)

	env := analyzer.GetCurrentFileEnv()

	if env == nil {
		t.Fatalf("Currently analyzed file's env is nil.")
	}

	expectedFileDecls := []struct {
		declName   string
		symbolType symbols.Symbol
	}{
		{"OutsideOfContract", &symbols.Event{}},
		{"OutsideOfContractUnused", &symbols.Event{}},
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
		{"InsideOfContractUnused", &symbols.Event{}},
		{"count", &symbols.StateVariable{}},
		{"increment", &symbols.Function{}},
		{"decrement", &symbols.Function{}},
		{"reset", &symbols.Function{}},
	}

	// Iterate through all contracts discovered in the file.
	contracts := symbols.GetAllSymbolsByType[*symbols.Contract](env)
	if len(contracts) != 1 {
		t.Fatalf("Expected %d contracts, but found %d.", 1, len(contracts))
	}

	for _, contract := range contracts {
		if contract.Name != expectedFileDecls[2].declName {
			t.Fatalf("Unexpected contract name: '%s'. Expected: '%s'.", contract.Name, expectedFileDecls[0].declName)
		}

		// Retrieve the environment specific to this contract
		contractEnv, err := symbols.GetInnerEnv(contract)
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

func Test_ResolveReferences(t *testing.T) {
	testContractPath := "testdata/foundry/src/003_SimpleCounter_WithEvents.sol"
	analyzer := Analyzer{}
	analyzer.Init(testContractPath)

	analyzer.AnalyzeCurrentFile()

	checkAnalyzerErrors(t, &analyzer)

	fileEnv := analyzer.GetCurrentFileEnv()

	if fileEnv == nil {
		t.Fatalf("Currently analyzed file's env is nil.")
	}

	// 1. In the file env get all events.
	// 2. In the contract get all events.
	// 3. Check the references: 2 of them should be filled, 2 should be empty

	fileEvents := symbols.GetAllSymbolsByType[*symbols.Event](fileEnv)
	if len(fileEvents) != 2 {
		t.Fatalf("Expected 2 file scope events, got: %d", len(fileEvents))
	}

	contracts := symbols.GetAllSymbolsByType[*symbols.Contract](fileEnv)
	if len(contracts) != 1 {
		t.Fatalf("Expected 1 contract, but found %d.", len(contracts))
	}

	contractEnv, err := symbols.GetInnerEnv(contracts[0])
	if err != nil {
		t.Fatalf("Error getting inner env of symbol: %s", err)
	}

	contractEvents := symbols.GetAllSymbolsByType[*symbols.Event](contractEnv)
	if len(contractEvents) != 2 {
		t.Fatalf("Expected 2 file scope events, got: %d", len(contractEvents))
	}

	events := []*symbols.Event{}
	events = append(events, fileEvents...)
	events = append(events, contractEvents...)

	if len(events) != 4 {
		t.Fatalf("Expected 4 events in total, got %d", len(events))
	}

	expectedReferences := []struct {
		ref *symbols.Reference
	}{
		{
			&symbols.Reference{
				Context: symbols.ReferenceContext{
					ScopeName: "decrement",
					ScopeType: 3, // Used in a function
					Usage:     4, // Event emission
				},
			},
		},
		{nil},
		{
			&symbols.Reference{
				Context: symbols.ReferenceContext{
					ScopeName: "increment",
					ScopeType: 3, // Used in a function
					Usage:     4, // Event emission
				},
			},
		},
		{
			&symbols.Reference{
				Context: symbols.ReferenceContext{
					ScopeName: "",
					ScopeType: 0,
					Usage:     0,
				},
			},
		},
		{nil},
	}

	for idx, event := range events {
		if event.References != nil {
			gotScopeName := event.References[0].Context.ScopeName
			expectedScopeName := expectedReferences[idx].ref.Context.ScopeName
			if gotScopeName != expectedScopeName {
				t.Fatalf("Event '%s' got incorrect reference scope NAME. Got: %s, expected: %s.",
					event.Name, gotScopeName, expectedScopeName)
			}

			gotScopeType := event.References[0].Context.ScopeType.String()
			expectedScopeType := expectedReferences[idx].ref.Context.ScopeType.String()
			if gotScopeType != expectedScopeType {
				t.Fatalf("Event '%s' got incorrect reference scope TYPE. Got: %s, expected: %s.",
					event.Name, gotScopeType, expectedScopeType)
			}

			gotUsage := event.References[0].Context.Usage.String()
			expectedUsage := expectedReferences[idx].ref.Context.Usage.String()
			if gotUsage != expectedUsage {
				t.Fatalf("Event '%s' got incorrect reference scope USAGE. Got: %s, expected: %s.",
					event.Name, gotUsage, expectedUsage)
			}
		} else {
			if expectedReferences[idx].ref == nil && len(event.References) != 0 {
				t.Fatalf("Event '%s' got unexpected reference. Expected none.", event.Name)
			}
		}
	}
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

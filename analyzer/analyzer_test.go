package analyzer

import (
	"solbot/parser"
	"solbot/symbols"
	"solbot/token"
	"testing"
)

func Test_PopulateSymbols_GetSymbolsByName(t *testing.T) {
	testContractPath := "testdata/foundry/src/002_SimpleCounter.sol"
	p := parser.Parser{}
	sourceFile, err := token.NewSourceFile(testContractPath, "")
	if err != nil {
		t.Fatalf("Could not create source file: %s", err)
	}

	p.Init(sourceFile)

	file := p.ParseFile()
	checkParserErrors(t, &p)

	if file == nil {
		t.Fatalf("ParseFile() returned nil")
	}

	env := symbols.NewEnvironment()
	discoverSymbols(file, env, nil)

	tests := []struct {
		expectedSymbolName string
	}{
		{"increment"},
		{"decrement"},
		{"reset"},
	}

	for _, tt := range tests {
		_, found := env.Get(tt.expectedSymbolName)
		if !found {
			t.Fatalf("Symbol: '%s' not found.", tt.expectedSymbolName)
		}
	}
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("Parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("Parser error: %s", err.Msg)
	}
	t.FailNow()
}

package analyzer

import (
	"testing"
)

func Test_DiscoverSymbols_GetSymbolsByName(t *testing.T) {
	testContractPath := "testdata/foundry/src/002_SimpleCounter.sol"
	analyzer := Analyzer{}
	analyzer.Init(testContractPath)
	checkParserErrors(t, &analyzer)

	analyzer.AnalyzeCurrentFile()

	tests := []struct {
		expectedSymbolName string
	}{
		{"Counter"},
		{"increment"},
		{"decrement"},
		{"reset"},
	}

	for _, tt := range tests {
		env := analyzer.GetCurrentFileEnv()
		if env == nil {
			t.Fatalf("Currently analyzed file's env is nil.")
		}
		_, found := env.Get(tt.expectedSymbolName)
		if !found {
			t.Fatalf("Symbol: '%s' not found.", tt.expectedSymbolName)
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

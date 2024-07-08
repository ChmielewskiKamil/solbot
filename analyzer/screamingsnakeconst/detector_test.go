package screamingsnakeconst

import (
	"solbot/parser"
	"solbot/reporter"
	"solbot/token"
	"testing"
)

func Test_DetectSnakeCaseConst(t *testing.T) {
	src := `address owner = 0x12345;          // no match
	   bool constant IS_OWNER = true;            // no match
	   bool constant isOwner = false;            // match
	   bool constant is_owner = false;           // match
	   uint256 balance = 100;                    // no match
	   address constant router = 0x1337;         // match
	   bool isOwner = true;                      // no match
	   uint16 constant ONE_hundred_IS_100 = 100; // match
	   uint256 constant DENOMINATOR = 1_000_000; // no match
	   `

	p := parser.Parser{}

	handle := token.NewFile("test.sol", src)

	p.Init(handle)

	file := p.ParseFile()
	if file == nil {
		t.Fatalf("Parsed file is nil")
	}
	if len(file.Declarations) == 0 {
		t.Fatalf("Parsed file has no declarations")
	}

	d := Detector{}

	finding := d.Detect(file)
	if finding == nil {
		t.Fatalf("Expected a finding, got nil")
	}

	numResults := 4

	if len(finding.Locations) != numResults {
		t.Errorf("Expected %d findings, got %d", numResults, len(finding.Locations))
	}

	finding.CalculatePositions(handle)

	expectedLocations := []reporter.Location{
		{Position: token.Position{Line: 3, Column: 19}, Context: "isOwner"},
		{Position: token.Position{Line: 4, Column: 19}, Context: "is_owner"},
		{Position: token.Position{Line: 6, Column: 22}, Context: "router"},
		{Position: token.Position{Line: 8, Column: 21}, Context: "ONE_hundred_IS_100"},
	}

	for i, loc := range finding.Locations {
		if loc.Position.Line != expectedLocations[i].Position.Line {
			t.Errorf("Expected line %d, got %d", expectedLocations[i].Position.Line, loc.Position.Line)
		}

		if loc.Position.Column != expectedLocations[i].Position.Column {
			t.Errorf("Expected column %d, got %d", expectedLocations[i].Position.Column, loc.Position.Column)
		}

		if loc.Context != expectedLocations[i].Context {
			t.Errorf("Expected context %s, got %s", expectedLocations[i].Context, loc.Context)
		}
	}
}

func Test_ShouldReturnNilIfNoVariables(t *testing.T) {
	src := `
    function foo() public {}
    `

	p := parser.Parser{}

	handle := token.NewFile("test.sol", src)

	p.Init(handle)

	file := p.ParseFile()
	d := Detector{}

	finding := d.Detect(file)

	if finding != nil {
		t.Fatalf("Expected nil, got a finding")
	}
}

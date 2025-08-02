package screamingsnakeconst

import (
	"github.com/ChmielewskiKamil/solbot/parser"
	"strings"
	"testing"
)

func Test_DetectSnakeCaseConst(t *testing.T) {
	src := `contract Test {
       address owner = 0x12345;                  // no match
	   bool constant IS_OWNER = true;            // no match
	   bool constant isOwner = false;            // match
	   bool constant is_owner = false;           // match
	   uint256 balance = 100;                    // no match
	   address constant router = 0x1337;         // match
	   bool isOwner = true;                      // no match
	   uint16 constant ONE_hundred_IS_100 = 100; // match
	   uint256 constant DENOMINATOR = 1_000_000; // no match
       }
	   `

	file, err := parser.ParseFile("test.sol", strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if file == nil {
		t.Fatalf("Parsed file is nil")
	}

	if len(file.Declarations) != 1 {
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

	// TODO: CalculatePositions expects source file but with the new ParseFile
	// API, the source file is not accessible.

	// finding.CalculatePositions(file)

	// expectedLocations := []reporter.Location{
	// 	{Position: token.Position{Line: 4, Column: 19}, Context: "isOwner"},
	// 	{Position: token.Position{Line: 5, Column: 19}, Context: "is_owner"},
	// 	{Position: token.Position{Line: 7, Column: 22}, Context: "router"},
	// 	{Position: token.Position{Line: 9, Column: 21}, Context: "ONE_hundred_IS_100"},
	// }
	//
	// for i, loc := range finding.Locations {
	// 	if loc.Position.Line != expectedLocations[i].Position.Line {
	// 		t.Errorf("Expected line %d, got %d", expectedLocations[i].Position.Line, loc.Position.Line)
	// 	}
	//
	// 	if loc.Position.Column != expectedLocations[i].Position.Column {
	// 		t.Errorf("Expected column %d, got %d", expectedLocations[i].Position.Column, loc.Position.Column)
	// 	}
	//
	// 	if loc.Context != expectedLocations[i].Context {
	// 		t.Errorf("Expected context %s, got %s", expectedLocations[i].Context, loc.Context)
	// 	}
	// }
}

func Test_ShouldReturnNilIfNoVariables(t *testing.T) {
	src := `
    function foo() public {}
    `

	file, err := parser.ParseFile("test.sol", strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	d := Detector{}

	finding := d.Detect(file)

	if finding != nil {
		t.Fatalf("Expected nil, got a finding")
	}
}

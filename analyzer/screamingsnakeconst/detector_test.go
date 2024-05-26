package screamingsnakeconst

import (
	"solparsor/parser"
	"testing"
)

func Test_DetectSnakeCaseConst(t *testing.T) {
	src := `
    address owner = 0x12345;                  // no match
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

	p.Init(src)

	file := p.ParseFile()
	d := Detector{}

	finding := d.Detect(file)
	if finding == nil {
		t.Fatalf("Expected a finding, got nil")
	}

	numResults := 4

	if len(finding.Locations) != numResults {
		t.Errorf("Expected %d findings, got %d", numResults, len(finding.Locations))
	}
}

func Test_ShouldReturnNilIfNoVariables(t *testing.T) {
	src := `
    function foo() public {
    }
    `

	p := parser.Parser{}

	p.Init(src)

	file := p.ParseFile()
	d := Detector{}

	finding := d.Detect(file)

	if finding != nil {
		t.Fatalf("Expected nil, got a finding")
	}
}

package main

import (
	"solbot/analyzer/screamingsnakeconst"
	"solbot/parser"
	"solbot/reporter"
)

func main() {
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
	d := screamingsnakeconst.Detector{}

	finding := d.Detect(file)

	list := []reporter.Finding{}
	if finding != nil {
		list = append(list, *finding)
	}
	if len(list) == 0 {
		panic("No finding")
	}

	err := reporter.GenerateReport(list, "./report.md")
	if err != nil {
		panic(err)
	}
}

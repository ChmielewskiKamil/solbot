package analyzer

import (
	"solbot/analyzer/screamingsnakeconst"
	"solbot/ast"
	"solbot/reporter"
	"solbot/symbols"
)

type Detector interface {
	Detect(node ast.Node) *reporter.Finding
}

func GetAllDetectors() *[]Detector {
	return &[]Detector{
		&screamingsnakeconst.Detector{},
	}
}

func AnalyzeFile(file *ast.File) []reporter.Finding {
	var findings []reporter.Finding

	detectors := *GetAllDetectors()

	for _, detector := range detectors {
		finding := detector.Detect(file)
		if finding != nil {
			findings = append(findings, *finding)
		}
	}

	return findings
}

func PopulateSymbols(node ast.Node, env *symbols.Environment) {}

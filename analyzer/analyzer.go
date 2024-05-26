package analyzer

import (
	"solparsor/analyzer/screamingsnakeconst"
	"solparsor/ast"
	"solparsor/reporter"
)

type Detector interface {
	Detect(node ast.Node) *reporter.Finding
}

func GetAllDetectors() *[]Detector {
	return &[]Detector{
		&screamingsnakeconst.Detector{},
	}
}

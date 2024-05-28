package analyzer

import (
	"solbot/analyzer/screamingsnakeconst"
	"solbot/ast"
	"solbot/reporter"
)

type Detector interface {
	Detect(node ast.Node) *reporter.Finding
}

func GetAllDetectors() *[]Detector {
	return &[]Detector{
		&screamingsnakeconst.Detector{},
	}
}

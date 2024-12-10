package analyzer

import (
	"solbot/analyzer/screamingsnakeconst"
	"solbot/ast"
	"solbot/reporter"
	"solbot/symbols"
	"solbot/token"
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
	globalEnv := symbols.NewEnvironment()

	// Phase 1: Get all declarations first to avoid unknown symbol errors if
	// the symbols are defined later in a file or somewhere else.
	discoverSymbols(file, globalEnv, nil)

	// Phase 2: Populate all definitions and references. Resolve overrides and
	// inheritance structure.
	resolveDefinitions(file, globalEnv)

	// Phase 3: The environment is populated with context at this point.
	// Diagnose issues with the code. Run detectors.
	findings := detectIssues(file, globalEnv)

	return findings
}

////////////////////////////////////////////////////////////////////
//                            PHASE 1			                  //
////////////////////////////////////////////////////////////////////

func discoverSymbols(node ast.Node,
	env *symbols.Environment, src *token.SourceFile) {
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			discoverSymbols(decl, env, n.SourceFile)
		}
	case *ast.FunctionDeclaration:
		populateFunctionDeclaration(n, env, src)
	}
}

func populateFunctionDeclaration(
	node *ast.FunctionDeclaration,
	env *symbols.Environment,
	src *token.SourceFile) {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: src,
		Offset:     node.Pos,
		AstNode:    node,
	}

	fnSymbol := &symbols.FunctionDeclaration{
		BaseSymbol: baseSymbol,
	}

	env.Set(node.Name.Value, fnSymbol)
}

////////////////////////////////////////////////////////////////////
//                            PHASE 2			                  //
////////////////////////////////////////////////////////////////////

func resolveDefinitions(node ast.Node, env *symbols.Environment) {}

func detectIssues(node ast.Node, env *symbols.Environment) []reporter.Finding {
	var findings []reporter.Finding

	// detectors := *GetAllDetectors()
	//
	// for _, detector := range detectors {
	// 	finding := detector.Detect(file)
	// 	if finding != nil {
	// 		findings = append(findings, *finding)
	// 	}
	// }

	return findings
}

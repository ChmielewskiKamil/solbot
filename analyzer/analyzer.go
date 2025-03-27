package analyzer

import (
	"fmt"
	"solbot/analyzer/screamingsnakeconst"
	"solbot/ast"
	"solbot/parser"
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

type Analyzer struct {
	// All findings found during the analysis.
	findings []reporter.Finding

	analysisErrors ErrorList

	currentFile    *ast.File            // The currently analysed file; returned from parser.ParseFile.
	currentFileEnv *symbols.Environment // The environment of the currently analyred file.
	parser         *parser.Parser       // Parser is used to parse newly encountered files.
}

func (a *Analyzer) Init(filePathToAnalyze string) error {
	a.parser = &parser.Parser{}
	a.analysisErrors = ErrorList{}

	sourceFile, err := token.NewSourceFile(filePathToAnalyze, "")
	if err != nil {
		return err
	}

	a.parser.Init(sourceFile)
	file := a.parser.ParseFile()
	if file == nil {
		return fmt.Errorf("Could not parse file. Check parses errors.")
	}

	a.currentFile = file
	return nil
}

func (a *Analyzer) AnalyzeCurrentFile() {
	a.AnalyzeFile(a.currentFile)
}

func (a *Analyzer) AnalyzeFile(file *ast.File) {
	fileEnv := symbols.NewEnvironment()
	a.currentFileEnv = fileEnv

	// Phase 1: Get all declarations first to avoid unknown symbol errors if
	// the symbols are defined later in a file or somewhere else.
	a.discoverSymbols(file, fileEnv)

	// Phase 2: Populate all references. Since all declarations in all scopes
	// should be known at this time, connect them to the place they are used.
	a.resolveReferences(file, fileEnv)

	// Phase 3: The environment is populated with context at this point.
	// Diagnose issues with the code. Run detectors.
	findings := a.detectIssues(file, fileEnv)

	for _, finding := range findings {
		a.findings = append(a.findings, finding)
	}
}

func (a *Analyzer) GetFindings() []reporter.Finding {
	return a.findings
}

func (a *Analyzer) GetCurrentFileEnv() *symbols.Environment {
	return a.currentFileEnv
}

func (a *Analyzer) GetParserErrors() parser.ErrorList {
	return a.parser.Errors()
}

////////////////////////////////////////////////////////////////////
//                            PHASE 1			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) discoverSymbols(node ast.Node, env *symbols.Environment) {
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			a.discoverSymbols(decl, env)
		}
	case *ast.ContractDeclaration:
		// Populate file's env with contract's declaration
		a.populateContractDeclaration(n, env)

		// Create contract specific env with file's env as outer.
		contractEnv := symbols.NewEnclosedEnvironment(env)
		for _, decl := range n.Body.Declarations {
			a.discoverSymbols(decl, contractEnv)
		}
	case *ast.FunctionDeclaration:
		// Contract's env is used as opposed to function's env because at the
		// discovery phase we only care about fn signature and we don't analyze
		// function's body in the context of function's parameters.
		a.populateFunctionDeclaration(n, env)
	case *ast.StateVariableDeclaration:
		a.populateStateVariableDeclaration(n, env)
	}
}

func (a *Analyzer) populateContractDeclaration(
	node *ast.ContractDeclaration, env *symbols.Environment) {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: a.currentFile.SourceFile,
		Offset:     node.Name.Pos,
		AstNode:    node,
	}

	contractSymbol := &symbols.Contract{
		BaseSymbol: baseSymbol,
	}

	env.Set(node.Name.Value, contractSymbol)
}

func (a *Analyzer) populateFunctionDeclaration(
	node *ast.FunctionDeclaration, env *symbols.Environment) {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: a.currentFile.SourceFile,
		Offset:     node.Name.Pos,
		AstNode:    node,
	}

	fnSymbol := &symbols.Function{
		BaseSymbol: baseSymbol,
	}

	env.Set(node.Name.Value, fnSymbol)
}

func (a *Analyzer) populateStateVariableDeclaration(
	node *ast.StateVariableDeclaration, env *symbols.Environment) {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: a.currentFile.SourceFile,
		Offset:     node.Name.Pos,
		AstNode:    node,
	}

	stateVarSymbol := &symbols.StateVariable{
		BaseSymbol: baseSymbol,
	}

	env.Set(node.Name.Value, stateVarSymbol)
}

////////////////////////////////////////////////////////////////////
//                            PHASE 2			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) resolveReferences(node ast.Node, env *symbols.Environment) {}

////////////////////////////////////////////////////////////////////
//                            PHASE 3			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) detectIssues(node ast.Node, env *symbols.Environment) []reporter.Finding {
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

////////////////////////////////////////////////////////////////////
//                            Helpers			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) GetNodeLocation(node ast.Node, pos token.Pos) string {
	if node == nil {
		return "Unknown location: nil node"
	}

	sourceFile := a.currentFile.SourceFile
	if sourceFile == nil {
		return fmt.Sprintf("Unknown location for node at position %d", pos)
	}

	line, column := sourceFile.GetLineAndColumn(pos)

	return fmt.Sprintf("%s:%d:%d", sourceFile.RelativePathFromProjectRoot(), line, column)
}

// Errors returns the combined list of errors encountered during the analysis.
// It includes errors from all phases.
func (a *Analyzer) Errors() ErrorList {
	return a.analysisErrors
}

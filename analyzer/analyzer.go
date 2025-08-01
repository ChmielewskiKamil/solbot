package analyzer

import (
	"fmt"
	"os"
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
}

func (a *Analyzer) Init(filePathToAnalyze string) error {
	f, err := os.Open(filePathToAnalyze)
	if err != nil {
		return fmt.Errorf("Could not open the path to analyze %s: %w", filePathToAnalyze, err)
	}
	defer f.Close()

	file, err := parser.ParseFile(filePathToAnalyze, f)
	if err != nil {
		return fmt.Errorf("Error while parsing the file %s: %w", filePathToAnalyze, err)
	}

	a.analysisErrors = ErrorList{}

	a.currentFile = file
	return nil
}

func (a *Analyzer) AnalyzeCurrentFile() {
	a.AnalyzeFile(a.currentFile)
}

func (a *Analyzer) AnalyzeFile(file *ast.File) {
	fileEnv := symbols.NewEnvironment(file.Name, symbols.FILE)
	a.currentFileEnv = fileEnv

	// Phase 1: Get all declarations first to avoid unknown symbol errors if
	// the symbols are defined later in a file or somewhere else (inheritance).
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

////////////////////////////////////////////////////////////////////
//                            PHASE 1			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) discoverSymbols(node ast.Node, outer *symbols.Environment) {
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			a.discoverSymbols(decl, outer)
		}
	case *ast.ContractDeclaration:
		// Populate file's env with contract's declaration
		contractSymbol := a.discoverContractDeclaration(n, outer)

		// Create contract specific env with file's env as outer.
		contractEnv := symbols.NewEnclosedEnvironment(outer, n.Name.Value, symbols.CONTRACT)

		contractSymbol.SetInnerEnv(contractEnv)

		for _, decl := range n.Body.Declarations {
			a.discoverSymbols(decl, contractEnv)
		}
	case *ast.FunctionDeclaration:
		// Add the function declaration to the current env (most often its the
		// contract's env). Function body can be discovered in the context of
		// function's inner ENV.
		functionSymbol := a.discoverFunctionDeclaration(n, outer)
		functionEnv := symbols.NewEnclosedEnvironment(outer, n.Name.Value, symbols.FUNCTION)

		functionSymbol.SetInnerEnv(functionEnv)

		// Statements in the function's body can be analyzed in the context of
		// the function's inner env.
		a.discoverSymbols(n.Body, functionEnv)
	case *ast.StateVariableDeclaration:
		a.discoverStateVariableDeclaration(n, outer)
	case *ast.EventDeclaration:
		// Event declaration can be present in the Contract as well as outside
		a.discoverEventDeclaration(n, outer)
	}
}

func (a *Analyzer) discoverContractDeclaration(
	node *ast.ContractDeclaration, env *symbols.Environment) *symbols.Contract {
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

	return contractSymbol
}

func (a *Analyzer) discoverFunctionDeclaration(
	node *ast.FunctionDeclaration, env *symbols.Environment) *symbols.Function {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: a.currentFile.SourceFile,
		Offset:     node.Name.Pos,
		AstNode:    node,
	}

	fnSymbol := &symbols.Function{
		BaseSymbol: baseSymbol,
		Visibility: node.Visibility,
		Mutability: node.Mutability,
		Virtual:    node.Virtual,
	}

	// TODO: Finish populating function symbol with: Parameters, Results

	env.Set(node.Name.Value, fnSymbol)

	return fnSymbol
}

func (a *Analyzer) discoverStateVariableDeclaration(
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

func (a *Analyzer) discoverEventDeclaration(
	node *ast.EventDeclaration, env *symbols.Environment) {
	baseSymbol := symbols.BaseSymbol{
		Name:       node.Name.Value,
		SourceFile: a.currentFile.SourceFile,
		Offset:     node.Name.Pos,
		AstNode:    node,
	}

	eventSymbol := &symbols.Event{
		BaseSymbol:  baseSymbol,
		IsAnonymous: node.IsAnonymous,
	}

	for _, param := range node.Params.List {
		if param != nil {
			eventParamSymbol := &symbols.EventParam{
				BaseSymbol: symbols.BaseSymbol{
					Name:       param.Name.Value,
					SourceFile: a.currentFile.SourceFile,
					// TODO: Should an LSP point at the type of param as the
					// location or at the name of the param?
					Offset:  param.Name.Pos,
					AstNode: param,
				},
				IsIndexed: param.IsIndexed,
			}
			eventSymbol.Parameters = append(eventSymbol.Parameters, eventParamSymbol)
		}
	}
	env.Set(node.Name.Value, eventSymbol)
}

////////////////////////////////////////////////////////////////////
//                            PHASE 2			                  //
////////////////////////////////////////////////////////////////////

func (a *Analyzer) resolveReferences(node ast.Node, env *symbols.Environment) {
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			a.resolveReferences(decl, env)
		}
	case *ast.ContractDeclaration:
		a.resolveContractDeclaration(n, env)
	case *ast.FunctionDeclaration:
		a.resolveFunctionDeclaration(n, env)
	case *ast.BlockStatement:
		a.resolveBlockStatement(n, env)
	}
}

func (a *Analyzer) resolveContractDeclaration(contractNode *ast.ContractDeclaration, env *symbols.Environment) {
	// Find ENV and resolve in its context.
	contractSymbol, found := env.Get(contractNode.Name.Value)
	if !found {
		a.analysisErrors.Add(a.GetNodeLocation(contractNode, contractNode.Name.Pos),
			"Reference resolution error: No symbol with this name found for contract '"+
				contractNode.Name.Value+"'.")
	}

	// Each contract should have a unique name. If there are more, there is
	// an issue.
	if len(contractSymbol) != 1 {
		a.analysisErrors.Add(a.GetNodeLocation(contractNode, contractNode.Name.Pos),
			"Reference resolution error: Found multiple symbols with the same name for contract '"+
				contractNode.Name.Value+"'.")
	}

	// Safe to access 0th element since we check the array length before.
	contractEnv, err := symbols.GetInnerEnv(contractSymbol[0])
	if err != nil {
		a.analysisErrors.Add(
			a.GetNodeLocation(contractNode, contractNode.Name.Pos),
			"Reference resolution error: "+err.Error(),
		)
	}

	for _, decl := range contractNode.Body.Declarations {
		a.resolveReferences(decl, contractEnv)
	}
}

func (a *Analyzer) resolveFunctionDeclaration(fnNode *ast.FunctionDeclaration, env *symbols.Environment) {
	functionSymbol, found := env.Get(fnNode.Name.Value)
	if !found {
		a.analysisErrors.Add(a.GetNodeLocation(fnNode, fnNode.Name.Pos),
			"Reference resolution error: No symbol with this name found for function '"+
				fnNode.Name.Value+"'.")
	}

	functionEnv, err := symbols.GetInnerEnv(functionSymbol[0])
	if err != nil {
		a.analysisErrors.Add(
			a.GetNodeLocation(fnNode, fnNode.Name.Pos),
			"Reference resolution error: "+err.Error(),
		)
	}

	a.resolveReferences(fnNode.Body, functionEnv)
}

func (a *Analyzer) resolveBlockStatement(blockNode *ast.BlockStatement, env *symbols.Environment) {
	for _, statement := range blockNode.Statements {
		a.resolveStatement(statement, env)
	}
}

func (a *Analyzer) resolveStatement(statement ast.Statement, env *symbols.Environment) {
	switch stmt := statement.(type) {
	case *ast.EmitStatement:
		a.resolveEmitStatement(stmt, env)
	}
}

func (a *Analyzer) resolveEmitStatement(stmt *ast.EmitStatement, env *symbols.Environment) {
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		a.analysisErrors.Add(a.GetNodeLocation(stmt, stmt.Pos),
			"Reference resolution error: expected call expression in an emit statement.")
	}

	ident, ok := call.Ident.(*ast.Identifier)
	if !ok {
		a.analysisErrors.Add(a.GetNodeLocation(call, call.Pos),
			"Reference resolution error: emit statement must refer to an event identifier.")
	}

	matchingSymbols, found := env.Get(ident.Value)
	if !found {
		a.analysisErrors.Add(a.GetNodeLocation(ident, ident.Start()),
			"Reference resolution error: No symbol found for event '"+
				ident.Value+"'.")
	}

	// TODO: Validate arguments match parameters.
	// TODO: Handle the situations when many symbols match

	eventSymbol, ok := matchingSymbols[0].(*symbols.Event)
	if !ok {
		a.analysisErrors.Add(a.GetNodeLocation(ident, ident.Start()),
			"Reference resolution error: symbols found with name '"+
				ident.Value+"' does not match the Event type.")
	}

	ref := &symbols.Reference{
		SourceFile: a.currentFile.SourceFile,
		Offset:     ident.Pos,
		Context: symbols.ReferenceContext{
			ScopeName: env.GetCurrentScopeName(),
			ScopeType: env.GetCurrentScopeType(),
			Usage:     symbols.EMIT,
		},
		AstNode: stmt,
	}

	eventSymbol.References = append(eventSymbol.References, ref)
}

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

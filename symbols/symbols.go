package symbols

import (
	"fmt"
	"solbot/ast"
	"solbot/token"
)

// Symbol represents any identifiable entity in the Solidity source code, such as contracts,
// functions, or variables. The purpose of this interface is to standardize access to essential
// metadata, like the location of the symbol in the source file. This makes it easier to generalize
// operations across different symbol types during analysis.
type Symbol interface {
	Location() string // Prints the symbol's location in the format: path/from/project/root/file.sol:Line:Column
}

type BaseSymbol struct {
	Name       string            // symbol name e.g. "Vault", "add", "balanceOf", "x", "Ownable"
	SourceFile *token.SourceFile // Pointer to the source file were symbol was declared.
	Offset     token.Pos         // Offset to the symbol name.
	References []Reference       // Places where the symbol was used.
	AstNode    ast.Node          // Pointer to ast node.
}

func (bs *BaseSymbol) Location() string {
	if bs.SourceFile != nil {
		loc := ""
		loc += bs.SourceFile.RelativePathFromProjectRoot()
		loc += ":"

		line, column := bs.SourceFile.GetLineAndColumn(bs.Offset)
		loc += fmt.Sprintf("%d:%d", line, column)

		return loc
	}

	return fmt.Sprintf("Missing location of symbol: %s. No source file info.", bs.Name)
}

type (
	Contract struct {
		BaseSymbol
	}

	Function struct {
		BaseSymbol
		Parameters []*Param
		Results    []*Param
		Visibility ast.Visibility
		Mutability ast.Mutability
		Virtual    bool
	}

	Param struct {
		BaseSymbol
		// TODO: What about the type?
		DataLocation ast.DataLocation
	}

	StateVariable struct {
		BaseSymbol
	}

	Event struct {
		BaseSymbol
		Parameters  []*EventParam
		IsAnonymous bool
	}

	EventParam struct {
		BaseSymbol
		// TODO: What about the type?
		IsIndexed bool
	}
)

////////////////////////////////////////////////////////////////////
//                          References		                      //
////////////////////////////////////////////////////////////////////

// References are resolved in the second phase of the analysis. They can be
// analyzed to undersand where a symbol is used and how.
type Reference struct {
	SourceFile *token.SourceFile // Pointer to the source file were symbol reference was found.
	Offset     token.Pos         // Offset to the place where symbol was referenced in the source file.
	Context    ReferenceContext  // Info about usage and scope e.g. state var is written to in function foo()
	AstNode    *ast.Node         // Pointer to ast node.
}

type ReferenceContext struct {
	ScopeName string
	ScopeType ReferenceScopeType
	Usage     ReferenceUsageType // How the reference was used: "call", "read", "write".
}

type ReferenceUsageType int

const (
	_ ReferenceUsageType = iota
	READ
	WRITE
	CALL
)

func (u ReferenceUsageType) String() string {
	switch u {
	case READ:
		return "READ"
	case WRITE:
		return "WRITE"
	case CALL:
		return "CALL"
	default:
		return "UNKNOWN"
	}
}

type ReferenceScopeType int

const (
	_ ReferenceScopeType = iota
	CONTRACT
	FUNCTION
	CONSTRUCTOR
)

func (s ReferenceScopeType) String() string {
	switch s {
	case CONTRACT:
		return "CONTRACT"
	case FUNCTION:
		return "FUNCTION"
	case CONSTRUCTOR:
		return "CONSTRUCTOR"
	default:
		return "UNKNOWN"
	}
}

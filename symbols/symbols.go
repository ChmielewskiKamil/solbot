package symbols

import (
	"solbot/ast"
	"solbot/token"
)

type Symbol interface {
	Name() string                   // symbol name e.g. "Vault", "add", "balanceOf", "x", "Ownable"
	DeclarationLocation() token.Pos // offset from the file beginning to the symbol
	String() string                 // Pretty print of symbol name and exact position in specific file.
}

type Reference struct {
	FilePathFromProjectRoot string             // Path from project root to file where reference was found.
	Offset                  token.Pos          // Offset to the place where symbol was referenced in the source file.
	Usage                   ReferenceUsageType // How the reference was used: "call", "read", "write".
	AstNode                 *ast.Node          // Pointer to ast node.
}

type BaseSymbol struct {
	Name                    string      // Symbol name
	FilePathFromProjectRoot string      // Path from project root to file where symbol was declared.
	Offset                  token.Pos   // Offset to the symbol name.
	References              []Reference // Places where the symbol was used.
	AstNode                 *ast.Node   // Pointer to ast node.
}

type Contract struct {
	BaseSymbol
}

type FunctionDeclaration struct {
	BaseSymbol
}

type ReferenceUsageType int

const (
	_ ReferenceUsageType = iota
	READ
	WRITE
	CALL
)

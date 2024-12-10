package symbols

import (
	"fmt"
	"solbot/ast"
	"solbot/token"
)

type Symbol interface {
	Location() string // Prints the symbol's location in the format: path/from/project/root/file.sol:Line:Column
}

type BaseSymbol struct {
	Name       string            // symbol name e.g. "Vault", "add", "balanceOf", "x", "Ownable"	SourceFile *token.SourceFile // Pointer to the source file were symbol was declared
	SourceFile *token.SourceFile // Pointer to the source file were symbol was declared.
	Offset     token.Pos         // Offset to the symbol name.
	References []Reference       // Places where the symbol was used.
	AstNode    *ast.Node         // Pointer to ast node.
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

type Reference struct {
	SourceFile *token.SourceFile  // Pointer to the source file were symbol reference was found.
	Offset     token.Pos          // Offset to the place where symbol was referenced in the source file.
	Usage      ReferenceUsageType // How the reference was used: "call", "read", "write".
	AstNode    *ast.Node          // Pointer to ast node.
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

package parser

import (
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseImportDecl parses an import declaration
// Format: import ModuleName
func (p *Parser) parseImportDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse "import" keyword
	if !p.curTokenIs(lexer.IMPORT) {
		p.addErrorf("expected import, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "import"

	// Parse module name
	if !p.curTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier after import, got %s", p.curToken.Type)
		return nil
	}
	moduleName := p.curToken.Literal
	p.nextToken() // consume module name

	return &ast.ImportDecl{
		Position: pos,
		Module:   moduleName,
	}
}


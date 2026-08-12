package parser

import (
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseImportDecl parses an import declaration
// Format: import ModuleName or import "path/to/module"
func (p *Parser) parseImportDecl() ast.Decl {
	pos := tokenPos(p.curToken)

	// Parse "import" keyword
	if !p.curTokenIs(lexer.IMPORT) {
		p.addErrorf("expected import, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken() // consume "import"

	// Check if it's a string literal (path-based import) or identifier (module name)
	if p.curTokenIs(lexer.STRING) {
		// Path-based import: import "path/to/module"
		path := p.curToken.Literal
		p.nextToken() // consume string
		return &ast.ImportDecl{
			Position: pos,
			Path:     path,
		}
	} else if p.curTokenIs(lexer.IDENT) {
		// Module name import: import ModuleName (from same directory)
		moduleName := p.curToken.Literal
		p.nextToken() // consume module name
		return &ast.ImportDecl{
			Position: pos,
			Module:   moduleName,
		}
	} else {
		p.addErrorf("expected identifier or string after import, got %s", p.curToken.Type)
		return nil
	}
}


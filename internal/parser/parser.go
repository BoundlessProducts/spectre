package parser

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// Parser represents the parser
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token
	
	// pendingDescription stores a description that was consumed but should be
	// associated with the next declaration. This handles descriptions on separate
	// lines before declarations.
	pendingDescription string
}

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances both curToken and peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseError represents a parse error with position information
type ParseError struct {
	Message  string
	Position ast.Position
}

func (e *ParseError) Error() string {
	if e.Position.Line > 0 {
		return fmt.Sprintf("%d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
	}
	return e.Message
}

// Errors returns the list of parse errors
func (p *Parser) Errors() []string {
	return p.errors
}

// ErrorsWithPosition returns parse errors with position information
func (p *Parser) ErrorsWithPosition() []*ParseError {
	// For now, return simple errors (we'll enhance this to include positions)
	errors := make([]*ParseError, len(p.errors))
	for i, msg := range p.errors {
		errors[i] = &ParseError{
			Message:  msg,
			Position: ast.Position{}, // Will be populated from token positions
		}
	}
	return errors
}

// addError adds a parse error
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

// addErrorAt adds a parse error at a specific position
func (p *Parser) addErrorAt(pos ast.Position, msg string) {
	if pos.Line > 0 {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		p.errors = append(p.errors, msg)
	}
}

// expectPeek checks if the next token matches the expected type
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	pos := tokenPos(p.peekToken)
	p.addErrorfAt(pos, "expected next token to be %s, got %s instead", t, p.peekToken.Type)
	return false
}

// addErrorf adds a formatted parse error (uses current token position)
func (p *Parser) addErrorf(format string, args ...interface{}) {
	pos := tokenPos(p.curToken)
	p.addErrorfAt(pos, format, args...)
}

// addErrorfAt adds a formatted parse error at a specific position
func (p *Parser) addErrorfAt(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Always include position if available (Line > 0 means position is set)
	if pos.Line > 0 {
		p.errors = append(p.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		// Fallback: try to get position from current token
		if p.curToken.Position.Line > 0 {
			p.errors = append(p.errors, fmt.Sprintf("%d:%d: %s", p.curToken.Position.Line, p.curToken.Position.Column, msg))
		} else {
			p.errors = append(p.errors, msg)
		}
	}
}

// ParseFile parses a complete Spectre file and returns an AST File node
func (p *Parser) ParseFile() *ast.File {
	file := &ast.File{
		Position: ast.Position{
			Line:   1,
			Column: 1,
			Offset: 0,
		},
		Decls: []ast.Decl{},
	}

	// Parse all declarations until EOF
	for !p.curTokenIs(lexer.EOF) {
		// Handle descriptions: if description is followed by a declaration keyword,
		// store it as pending so the declaration parser can use it.
		if p.curTokenIs(lexer.DESCRIPTION) {
			// Check if this description is followed by a declaration keyword
			if p.peekToken.Type == lexer.STRING {
				// Temporarily consume description and string to see what follows
				p.nextToken() // consume "description"
				descText := p.curToken.Literal // store description text
				p.nextToken() // consume string
				
				// Check if next token is a declaration keyword
				if p.curTokenIs(lexer.VAR) || p.curTokenIs(lexer.CONST) || p.curTokenIs(lexer.FUN) ||
					p.curTokenIs(lexer.ACTION) || p.curTokenIs(lexer.INIT) || p.curTokenIs(lexer.INVARIANT) ||
					p.curTokenIs(lexer.TEMPORAL) || p.curTokenIs(lexer.MODULE) || p.curTokenIs(lexer.IMPORT) ||
					p.curTokenIs(lexer.TYPE) || p.curTokenIs(lexer.ENUM) {
					// Description is before a declaration - store it for the declaration parser
					p.pendingDescription = descText
					// Continue - the declaration parser will consume the declaration keyword
					// and use the pending description
				} else {
					// Standalone description - skip it
					p.pendingDescription = "" // Clear any pending description
					continue
				}
			} else {
				// Invalid description - skip it
				p.parseDescription()
				p.pendingDescription = "" // Clear any pending description
				continue
			}
		}

		var decl ast.Decl
		switch p.curToken.Type {
		case lexer.VAR:
			decl = p.parseVariableDecl()
		case lexer.CONST:
			decl = p.parseConstantDecl()
		case lexer.FUN:
			decl = p.parseFunctionDecl()
		case lexer.ACTION:
			decl = p.parseActionDecl()
		case lexer.INIT:
			decl = p.parseInitDecl()
		case lexer.INVARIANT:
			decl = p.parseInvariantDecl()
		case lexer.TEMPORAL:
			decl = p.parseTemporalDecl()
		case lexer.MODULE:
			decl = p.parseModuleDecl()
		case lexer.IMPORT:
			// Top-level imports are allowed (though typically they're in modules)
			decl = p.parseImportDecl()
		case lexer.TYPE:
			decl = p.parseTypeAliasDecl()
		case lexer.ENUM:
			decl = p.parseEnumDecl()
		default:
			// Skip unknown tokens (comments, whitespace, etc.)
			p.nextToken()
			continue
		}

		if decl != nil {
			file.Decls = append(file.Decls, decl)
		} else {
			// Error recovery: if parsing failed, sync to next declaration
			// This allows us to continue parsing even after errors
			if len(p.errors) > 0 {
				p.syncToDeclaration()
			}
		}
	}

	return file
}


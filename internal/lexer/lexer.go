package lexer

import (
	"unicode"
	"unicode/utf8"
)

// Lexer represents the lexer
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		l.position = l.readPosition
		l.readPosition++
	}
	
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

// readRuneAt reads a rune at the current read position without advancing
func (l *Lexer) readRuneAt() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken scans and returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	pos := Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.position,
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: ASSIGN, Literal: string(l.ch), Position: pos}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NEQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: NOT, Literal: string(l.ch), Position: pos}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LEQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Position: pos}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GEQ, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Position: pos}
		}
	case '+':
		tok = Token{Type: PLUS, Literal: string(l.ch), Position: pos}
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: MINUS, Literal: string(l.ch), Position: pos}
		}
	case '*':
		tok = Token{Type: ASTERISK, Literal: string(l.ch), Position: pos}
	case '/':
		if l.peekChar() == '/' {
			// Single-line comment
			l.skipSingleLineComment()
			return l.NextToken() // Return next token after comment
		} else if l.peekChar() == '*' {
			// Multi-line comment
			if err := l.skipMultiLineComment(); err != nil {
				tok = Token{Type: ILLEGAL, Literal: err.Error(), Position: pos}
				return tok
			}
			return l.NextToken() // Return next token after comment
		} else {
			tok = Token{Type: SLASH, Literal: string(l.ch), Position: pos}
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Position: pos}
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR, Literal: string(ch) + string(l.ch), Position: pos}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Position: pos}
		}
	case '\'':
		tok = Token{Type: PRIME, Literal: string(l.ch), Position: pos}
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Position: pos}
	case '.':
		if l.peekChar() == '.' && l.input[l.readPosition+1] == '.' {
			l.readChar()
			l.readChar()
			tok = Token{Type: ELLIPSIS, Literal: "...", Position: pos}
		} else {
			tok = Token{Type: DOT, Literal: string(l.ch), Position: pos}
		}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: string(l.ch), Position: pos}
	case ':':
		tok = Token{Type: COLON, Literal: string(l.ch), Position: pos}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch), Position: pos}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch), Position: pos}
	case '{':
		tok = Token{Type: LBRACE, Literal: string(l.ch), Position: pos}
	case '}':
		tok = Token{Type: RBRACE, Literal: string(l.ch), Position: pos}
	case '[':
		tok = Token{Type: LBRACKET, Literal: string(l.ch), Position: pos}
	case ']':
		tok = Token{Type: RBRACKET, Literal: string(l.ch), Position: pos}
	case 0:
		tok = Token{Type: EOF, Literal: "", Position: pos}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Position = pos
		return tok
	default:
		// Check for Unicode arrow character (→ U+2192) at current position
		if l.position < len(l.input) {
			r, size := utf8.DecodeRuneInString(l.input[l.position:])
			if r == '→' || r == 0x2192 {
				// Advance past the Unicode character
				tok = Token{Type: ARROW, Literal: string(r), Position: pos}
				l.position = l.position + size
				l.readPosition = l.position
				l.column++
				l.readChar() // Read the next character after the arrow
				return tok
			}
		}
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Position = pos
			return tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = l.readNumber()
			tok.Position = pos
			return tok
		} else {
			// For other Unicode characters, try to read as identifier
			if l.readPosition < len(l.input) {
				r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
				if r != utf8.RuneError && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
					// Read Unicode identifier
					tok.Literal = l.readUnicodeIdentifier()
					tok.Type = LookupIdent(tok.Literal)
					tok.Position = pos
					return tok
				}
			}
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Position: pos}
		}
	}

	l.readChar()
	return tok
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		// If next position is EOF, include current char and stop
		if l.readPosition >= len(l.input) {
			// Include current character in result
			l.position = l.readPosition
			l.ch = 0 // Set to EOF
			break
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

// readUnicodeIdentifier reads an identifier that may contain Unicode characters
func (l *Lexer) readUnicodeIdentifier() string {
	position := l.position
	for {
		if l.readPosition >= len(l.input) {
			break
		}
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		if r == utf8.RuneError {
			break
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		l.readPosition += size
		l.column++
	}
	return l.input[position:l.readPosition]
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(rune(ch))
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// readNumber reads a number (integer or float)
func (l *Lexer) readNumber() (TokenType, string) {
	position := l.position
	tokenType := INT

	for isDigit(l.ch) {
		// If next position is EOF, include current digit and stop
		if l.readPosition >= len(l.input) {
			l.position = l.readPosition
			l.ch = 0 // Set to EOF
			break
		}
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = FLOAT
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			// If next position is EOF, include current digit and stop
			if l.readPosition >= len(l.input) {
				l.position = l.readPosition
				l.ch = 0 // Set to EOF
				break
			}
			l.readChar()
		}
	}

	return tokenType, l.input[position:l.position]
}

// readString reads a string literal
func (l *Lexer) readString() string {
	position := l.position + 1 // skip opening quote
	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
		if l.ch == 0 {
			// Unterminated string - return what we have
			return l.input[position:]
		}
		if l.ch == '\\' {
			// Handle escape sequences
			l.readChar()
			if l.ch == 0 {
				break
			}
		}
	}
	str := l.input[position:l.position]
	l.readChar() // consume closing quote
	return str
}

// skipSingleLineComment skips a single-line comment (//)
func (l *Lexer) skipSingleLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipMultiLineComment skips a multi-line comment (/* */)
func (l *Lexer) skipMultiLineComment() error {
	l.readChar() // consume '*'
	for {
		if l.ch == 0 {
			return &LexerError{Message: "unterminated multi-line comment", Position: Position{Line: l.line, Column: l.column}}
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume '*'
			l.readChar() // consume '/'
			return nil
		}
		l.readChar()
	}
}

// LexerError represents a lexer error
type LexerError struct {
	Message  string
	Position Position
}

func (e *LexerError) Error() string {
	return e.Message
}


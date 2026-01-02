package lexer

// TokenType represents the type of a token
type TokenType string

const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"  // variable, function names
	INT    TokenType = "INT"    // integer literal
	FLOAT  TokenType = "FLOAT"  // float literal
	STRING TokenType = "STRING" // string literal
	BOOL   TokenType = "BOOL"   // true, false

	// Operators
	ASSIGN   TokenType = "="   // =
	EQ       TokenType = "=="  // ==
	NEQ      TokenType = "!="  // !=
	LT       TokenType = "<"   // <
	GT       TokenType = ">"   // >
	LEQ      TokenType = "<="  // <=
	GEQ      TokenType = ">="  // >=
	PLUS     TokenType = "+"   // +
	MINUS    TokenType = "-"   // -
	ASTERISK TokenType = "*"   // *
	SLASH    TokenType = "/"   // /
	AND      TokenType = "&&"  // &&
	OR       TokenType = "||"  // ||
	NOT      TokenType = "!"   // !
	PRIME    TokenType = "'"   // ' (next state)
	ARROW    TokenType = "->"  // -> (leads to)

	// Delimiters
	COMMA     TokenType = ","  // ,
	DOT       TokenType = "."  // .
	SEMICOLON TokenType = ";"  // ;
	COLON     TokenType = ":"  // :
	LPAREN    TokenType = "("  // (
	RPAREN    TokenType = ")"  // )
	LBRACE    TokenType = "{"  // {
	RBRACE    TokenType = "}"  // }
	LBRACKET  TokenType = "["  // [
	RBRACKET  TokenType = "]"  // ]
	ELLIPSIS  TokenType = "..." // ...

	// Keywords
	MODULE      TokenType = "MODULE"
	IMPORT      TokenType = "IMPORT"
	EXTENDS     TokenType = "EXTENDS"
	PUBLIC      TokenType = "PUBLIC"
	PRIVATE     TokenType = "PRIVATE"
	CONST       TokenType = "CONST"
	VAR         TokenType = "VAR"
	DESCRIPTION TokenType = "DESCRIPTION"
	INIT        TokenType = "INIT"
	ONEOF       TokenType = "ONEOF"
	ACTION      TokenType = "ACTION"
	FUN         TokenType = "FUN"
	INVARIANT   TokenType = "INVARIANT"
	TEMPORAL    TokenType = "TEMPORAL"
	REQUIRE     TokenType = "REQUIRE"
	ENSURE      TokenType = "ENSURE"
	IF          TokenType = "IF"
	ELSE        TokenType = "ELSE"
	THEN        TokenType = "THEN"
	LET         TokenType = "LET"
	RETURN      TokenType = "RETURN"
	TYPE        TokenType = "TYPE"
	ENUM        TokenType = "ENUM"
	WHEN        TokenType = "WHEN"
	SUPER       TokenType = "SUPER"
	WITH        TokenType = "WITH"

	// Temporal operators
	ALWAYS    TokenType = "ALWAYS"
	EVENTUALLY TokenType = "EVENTUALLY"
	UNTIL     TokenType = "UNTIL"
	WF        TokenType = "WF" // Weak fairness
	SF        TokenType = "SF" // Strong fairness
	NEXT      TokenType = "NEXT"

	// Type keywords
	SET   TokenType = "SET"
	MAP   TokenType = "MAP"
	LIST  TokenType = "LIST"
	OPTION TokenType = "OPTION"
)

// Token represents a token with its type, literal value, and position
type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

// Position represents the source position of a token
type Position struct {
	Line   int
	Column int
	Offset int
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	keywords := map[string]TokenType{
		"module":      MODULE,
		"import":      IMPORT,
		"extends":     EXTENDS,
		"public":      PUBLIC,
		"private":     PRIVATE,
		"const":       CONST,
		"var":         VAR,
		"description": DESCRIPTION,
		"init":        INIT,
		"oneOf":       ONEOF,
		"action":      ACTION,
		"fun":         FUN,
		"invariant":   INVARIANT,
		"temporal":    TEMPORAL,
		"require":     REQUIRE,
		"ensure":      ENSURE,
		"if":          IF,
		"else":        ELSE,
		"then":        THEN,
		"let":         LET,
		"return":      RETURN,
		"type":        TYPE,
		"enum":        ENUM,
		"when":        WHEN,
		"super":       SUPER,
		"with":        WITH,
		"always":      ALWAYS,
		"eventually":  EVENTUALLY,
		"until":       UNTIL,
		"WF":          WF,
		"SF":          SF,
		"next":        NEXT,
		"Set":         SET,
		"Map":         MAP,
		"List":        LIST,
		"Option":      OPTION,
		"true":        BOOL,
		"false":       BOOL,
	}

	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}


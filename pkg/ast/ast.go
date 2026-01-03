package ast

// Node represents any AST node
type Node interface {
	Pos() Position
}

// Position represents a source position
type Position struct {
	Line   int
	Column int
	Offset int
}

// File represents a complete Spectre file
type File struct {
	Position Position
	Decls    []Decl
}

// Decl represents a top-level declaration
type Decl interface {
	Node
	declNode()
}

// VariableDecl represents a variable declaration
type VariableDecl struct {
	Position    Position
	Description string
	Name        string
	Type        Type
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (v *VariableDecl) Pos() Position { return v.Position }
func (v *VariableDecl) declNode()     {}

// ConstantDecl represents a constant declaration
type ConstantDecl struct {
	Position    Position
	Description string
	Name        string
	Type        Type
	Value       Expr
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (c *ConstantDecl) Pos() Position { return c.Position }
func (c *ConstantDecl) declNode()      {}

// FunctionDecl represents a pure function declaration
type FunctionDecl struct {
	Position    Position
	Description string
	Name        string
	Parameters  []Parameter
	ReturnType  Type
	Body        *BlockStmt
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (f *FunctionDecl) Pos() Position { return f.Position }
func (f *FunctionDecl) declNode()      {}

// ActionDecl represents an action declaration
type ActionDecl struct {
	Position    Position
	Description string
	Name        string
	Parameters  []Parameter
	Guard       Expr // Optional guard condition (when clause)
	Body        *BlockStmt
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (a *ActionDecl) Pos() Position { return a.Position }
func (a *ActionDecl) declNode()     {}

// InitDecl represents an initial state declaration
type InitDecl struct {
	Position    Position
	Description string
	Body        *BlockStmt // Block of assignments, or nil if single expression
	Expression  Expr       // Single expression form, or nil if block form
}

func (i *InitDecl) Pos() Position { return i.Position }
func (i *InitDecl) declNode()     {}

// OneOfInitDecl represents a oneOf initial state declaration
type OneOfInitDecl struct {
	Position    Position
	Description string
	Options     []*BlockStmt // List of initial state options (each is a block of assignments)
}

func (o *OneOfInitDecl) Pos() Position { return o.Position }
func (o *OneOfInitDecl) declNode()     {}

// InvariantDecl represents an invariant declaration
type InvariantDecl struct {
	Position    Position
	Description string
	Name        string
	Condition   Expr // The invariant condition expression
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (i *InvariantDecl) Pos() Position { return i.Position }
func (i *InvariantDecl) declNode()      {}

// TemporalDecl represents a temporal property declaration
type TemporalDecl struct {
	Position    Position
	Description string
	Name        string
	Expression  Expr // The temporal expression (always, eventually, until, etc.)
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (t *TemporalDecl) Pos() Position { return t.Position }
func (t *TemporalDecl) declNode()     {}

// ModuleDecl represents a module declaration
type ModuleDecl struct {
	Position    Position
	Description string
	Name        string
	Extends     string // Name of parent module if extending, empty otherwise
	Decls       []Decl // All declarations within the module
}

func (m *ModuleDecl) Pos() Position { return m.Position }
func (m *ModuleDecl) declNode()     {}

// ImportDecl represents an import declaration
type ImportDecl struct {
	Position Position
	Module   string // Name of the module to import
}

func (i *ImportDecl) Pos() Position { return i.Position }
func (i *ImportDecl) declNode()     {}

// ModuleInstanceDecl represents a module instance declaration
type ModuleInstanceDecl struct {
	Position     Position
	Name         string // Name of the instance
	Module       string // Name of the module being instantiated
	Substitutions map[string]string // Map from original names to new names
}

func (m *ModuleInstanceDecl) Pos() Position { return m.Position }
func (m *ModuleInstanceDecl) declNode()     {}

// TypeAliasDecl represents a type alias declaration
// Format: type Name = Type
type TypeAliasDecl struct {
	Position    Position
	Description string
	Name        string
	Type        Type // The type being aliased
	Visibility  Visibility // Public or Private (defaults to Public)
}

func (t *TypeAliasDecl) Pos() Position { return t.Position }
func (t *TypeAliasDecl) declNode()     {}

// Visibility represents public/private visibility
type Visibility int

const (
	Public Visibility = iota
	Private
)

// Parameter represents a function parameter
type Parameter struct {
	Position Position
	Name     string
	Type     Type
}

// BlockStmt represents a block of statements
type BlockStmt struct {
	Position Position
	Statements []Stmt
}

// Stmt represents a statement
type Stmt interface {
	Node
	stmtNode()
}

// ReturnStmt represents a return statement
type ReturnStmt struct {
	Position Position
	Value    Expr
}

func (r *ReturnStmt) Pos() Position { return r.Position }
func (r *ReturnStmt) stmtNode()      {}

// ExprStmt represents an expression statement
type ExprStmt struct {
	Position Position
	Expr     Expr
}

func (e *ExprStmt) Pos() Position { return e.Position }
func (e *ExprStmt) stmtNode()      {}

// AssignStmt represents an assignment statement (e.g., counter = 0)
type AssignStmt struct {
	Position Position
	Left     Expr // Identifier or selector expression
	Right    Expr // Value being assigned
}

func (a *AssignStmt) Pos() Position { return a.Position }
func (a *AssignStmt) stmtNode()     {}

// RequireStmt represents a require (precondition) statement
type RequireStmt struct {
	Position Position
	Condition Expr
}

func (r *RequireStmt) Pos() Position { return r.Position }
func (r *RequireStmt) stmtNode()     {}

// EnsureStmt represents an ensure (postcondition) statement
type EnsureStmt struct {
	Position Position
	Condition Expr
}

func (e *EnsureStmt) Pos() Position { return e.Position }
func (e *EnsureStmt) stmtNode()      {}

// Type represents a type expression
type Type interface {
	Node
	typeNode()
}

// PrimitiveType represents a primitive type (int, bool, str, float)
type PrimitiveType struct {
	Position Position
	Name     string // "int", "bool", "str", "float"
}

func (p *PrimitiveType) Pos() Position { return p.Position }
func (p *PrimitiveType) typeNode()     {}

// RecordType represents a record/struct type
type RecordType struct {
	Position Position
	Fields   []Field
}

func (r *RecordType) Pos() Position { return r.Position }
func (r *RecordType) typeNode()     {}

// Field represents a field in a record
type Field struct {
	Position Position
	Name     string
	Type     Type
}

// SetType represents a Set type
type SetType struct {
	Position Position
	Element  Type
}

func (s *SetType) Pos() Position { return s.Position }
func (s *SetType) typeNode()     {}

// MapType represents a Map type
type MapType struct {
	Position Position
	Key      Type
	Value    Type
}

func (m *MapType) Pos() Position { return m.Position }
func (m *MapType) typeNode()     {}

// ListType represents a List/Array type
type ListType struct {
	Position Position
	Element  Type
}

func (l *ListType) Pos() Position { return l.Position }
func (l *ListType) typeNode()     {}

// EnumType represents an enum type
type EnumType struct {
	Position Position
	Name     string
	Values   []string
}

func (e *EnumType) Pos() Position { return e.Position }
func (e *EnumType) typeNode()     {}

// OptionType represents an Optional type
type OptionType struct {
	Position Position
	Element  Type
}

func (o *OptionType) Pos() Position { return o.Position }
func (o *OptionType) typeNode()    {}

// NamedType represents a named type (type alias)
type NamedType struct {
	Position Position
	Name     string
}

func (n *NamedType) Pos() Position { return n.Position }
func (n *NamedType) typeNode()     {}

// Expr represents an expression
type Expr interface {
	Node
	exprNode()
}

// Ident represents an identifier
type Ident struct {
	Position Position
	Name     string
	Prime    bool // true if this is a primed identifier (next state)
}

func (i *Ident) Pos() Position { return i.Position }
func (i *Ident) exprNode()     {}

// BasicLit represents a basic literal (int, float, string, bool)
type BasicLit struct {
	Position Position
	Kind     LitKind
	Value    string
}

type LitKind int

const (
	IntLit LitKind = iota
	FloatLit
	StringLit
	BoolLit
)

func (b *BasicLit) Pos() Position { return b.Position }
func (b *BasicLit) exprNode()     {}

// BinaryExpr represents a binary expression
type BinaryExpr struct {
	Position Position
	Op       BinaryOp
	Left     Expr
	Right    Expr
}

type BinaryOp int

const (
	Add BinaryOp = iota
	Sub
	Mul
	Div
	Eq
	Neq
	Lt
	Gt
	Leq
	Geq
	And
	Or
)

func (b *BinaryExpr) Pos() Position { return b.Position }
func (b *BinaryExpr) exprNode()      {}

// UnaryExpr represents a unary expression
type UnaryExpr struct {
	Position Position
	Op       UnaryOp
	Expr     Expr
}

type UnaryOp int

const (
	Not UnaryOp = iota
	Neg
)

func (u *UnaryExpr) Pos() Position { return u.Position }
func (u *UnaryExpr) exprNode()      {}

// CallExpr represents a function call
type CallExpr struct {
	Position Position
	Fun      Expr
	Args     []Expr
}

func (c *CallExpr) Pos() Position { return c.Position }
func (c *CallExpr) exprNode()      {}

// SelectorExpr represents a field/method selection (e.g., record.field)
type SelectorExpr struct {
	Position Position
	X        Expr
	Sel      string
}

func (s *SelectorExpr) Pos() Position { return s.Position }
func (s *SelectorExpr) exprNode()     {}

// IndexExpr represents an index expression (e.g., array[index])
type IndexExpr struct {
	Position Position
	X        Expr
	Index    Expr
}

func (i *IndexExpr) Pos() Position { return i.Position }
func (i *IndexExpr) exprNode()     {}

// ParenExpr represents a parenthesized expression
type ParenExpr struct {
	Position Position
	X        Expr
}

func (p *ParenExpr) Pos() Position { return p.Position }
func (p *ParenExpr) exprNode()      {}

// IfExpr represents an if-else expression
type IfExpr struct {
	Position Position
	Condition Expr
	Then      Expr
	Else      Expr
}

func (i *IfExpr) Pos() Position { return i.Position }
func (i *IfExpr) exprNode()      {}

// LetExpr represents a let binding expression
type LetExpr struct {
	Position Position
	Name     string
	Value    Expr
	Body     Expr
}

func (l *LetExpr) Pos() Position { return l.Position }
func (l *LetExpr) exprNode()      {}

// AlwaysExpr represents an always temporal expression
type AlwaysExpr struct {
	Position Position
	Expr     Expr
}

func (a *AlwaysExpr) Pos() Position { return a.Position }
func (a *AlwaysExpr) exprNode()     {}

// EventuallyExpr represents an eventually temporal expression
type EventuallyExpr struct {
	Position Position
	Expr     Expr
}

func (e *EventuallyExpr) Pos() Position { return e.Position }
func (e *EventuallyExpr) exprNode()     {}

// UntilExpr represents an until temporal expression
type UntilExpr struct {
	Position Position
	Left     Expr
	Right    Expr
}

func (u *UntilExpr) Pos() Position { return u.Position }
func (u *UntilExpr) exprNode()     {}

// WFExpr represents a weak fairness expression
type WFExpr struct {
	Position Position
	Target   Expr // Identifier (action name or variable name)
}

func (w *WFExpr) Pos() Position { return w.Position }
func (w *WFExpr) exprNode()     {}

// SFExpr represents a strong fairness expression
type SFExpr struct {
	Position Position
	Target   Expr // Identifier (action name or variable name)
}

func (s *SFExpr) Pos() Position { return s.Position }
func (s *SFExpr) exprNode()     {}

// LeadsToExpr represents a leads-to expression (→)
type LeadsToExpr struct {
	Position Position
	Left     Expr
	Right    Expr
}

func (l *LeadsToExpr) Pos() Position { return l.Position }
func (l *LeadsToExpr) exprNode()     {}

// SuperExpr represents a super call expression
type SuperExpr struct {
	Position Position
	Method   string // Name of the method/action to call
}

func (s *SuperExpr) Pos() Position { return s.Position }
func (s *SuperExpr) exprNode()     {}

// LambdaExpr represents a lambda expression (anonymous function)
// Format: param => expression or (param1, param2) => expression
type LambdaExpr struct {
	Position Position
	Params   []Parameter // Lambda parameters
	Body     Expr        // Lambda body expression
}

func (l *LambdaExpr) Pos() Position { return l.Position }
func (l *LambdaExpr) exprNode()     {}

// RecordLiteral represents a record literal value
// Format: { field1: value1, field2: value2, ... }
type RecordLiteral struct {
	Position Position
	Fields   []RecordField
}

// RecordField represents a field in a record literal
type RecordField struct {
	Name  string
	Value Expr
	Spread bool // true if this is a spread field: ...identifier
}

func (r *RecordLiteral) Pos() Position { return r.Position }
func (r *RecordLiteral) exprNode()      {}


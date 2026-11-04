# Appendix A: Language Grammar (EBNF)

The complete formal grammar for Spectre. Use this as the authoritative reference when the book's examples conflict with your intuition about syntax.

**Notation:**
- `::=` — definition
- `|` — alternative
- `[ X ]` — X is optional
- `{ X }` — zero or more X
- `( X | Y )` — grouping/choice
- `"word"` — literal keyword or symbol

---

## A.1 Top-level structure

A Spectre file is a sequence of declarations in any order.

```ebnf
File    ::= { TopDecl }

TopDecl ::= [ "description" STRING ]
            ( VarDecl
            | ConstDecl
            | ParamDecl
            | TypeAliasDecl
            | EnumDecl
            | InitDecl
            | ActionDecl
            | FunDecl
            | InvariantDecl
            | TemporalDecl
            | ModuleDecl
            | ImportDecl
            )
```

Example — a minimal file:
```spectre
var x: int
init { x = 0 }
action step { x' = x + 1 }
invariant positive { x >= 0 }
```

---

## A.2 Variable and constant declarations

```ebnf
VarDecl    ::= "var" Ident ":" TypeExpr

ConstDecl  ::= "const" Ident ":" TypeExpr "=" Expr

ParamDecl  ::= "param" Ident ":" TypeExpr
```

- `var` — a mutable state variable, changed by actions.
- `const` — a fixed value baked into the spec.
- `param` — a named constant whose value is supplied at run time via `--param NAME=value`.

```spectre
var balance: int
const LIMIT: int = 1000
param NUM_NODES: int          // set with --param NUM_NODES=5
```

---

## A.3 Types

```ebnf
TypeExpr ::= "int"
           | "bool"
           | "str"
           | "float"
           | "Set"  "<" TypeExpr ">"
           | "Map"  "<" TypeExpr "," TypeExpr ">"
           | "List" "<" TypeExpr ">"
           | "Option" "<" TypeExpr ">"
           | "{" FieldDecl { "," FieldDecl } "}"     (* record type *)
           | "(" TypeExpr { "," TypeExpr } ")"       (* tuple type *)
           | Ident                                   (* named type or enum *)

FieldDecl ::= Ident ":" TypeExpr
```

Examples:
```spectre
var count:   int
var active:  bool
var nodes:   Set<int>
var votes:   Map<int, bool>
var history: List<str>
var maybe:   Option<int>
```

---

## A.4 Enum declarations

```ebnf
EnumDecl    ::= "enum" Ident "{" EnumVariant { "," EnumVariant } "}"
EnumVariant ::= Ident
```

Variants are always referenced as `EnumName.Variant`.

```spectre
enum Phase { Preparing, Committed, Aborted }
var phase: Phase
init { phase = Phase.Preparing }
```

---

## A.5 Type aliases

```ebnf
TypeAliasDecl ::= "type" Ident "=" TypeExpr
```

```spectre
type NodeId = int
type VoteMap = Map<NodeId, bool>
var votes: VoteMap
```

---

## A.6 Init declaration

```ebnf
InitDecl  ::= "init" ( Block | OneOfBody | Expr )

Block     ::= "{" { AssignStmt } "}"

OneOfBody ::= "oneOf" "{" OneOfOpt { "," OneOfOpt } "}"
OneOfOpt  ::= Expr
            | "{" { Ident "=" Expr } "}"
```

- Single block: one deterministic initial state.
- `oneOf`: the verifier explores from every listed initial state.

```spectre
// Deterministic
init { balance = 0  frozen = false }

// Non-deterministic: explore from two starting states
init oneOf {
  { balance = 0, frozen = false },
  { balance = 100, frozen = false }
}
```

---

## A.7 Action declarations

```ebnf
ActionDecl ::= "action" Ident [ "(" ParamList ")" ] [ "when" Expr ] "{" { Stmt } "}"
ParamList  ::= Param { "," Param }
Param      ::= Ident ":" TypeExpr
```

The optional `when Expr` clause is a **declaration-level guard** — it is equivalent to placing `require Expr` as the first statement of the action body. The tutorial chapters use `require` exclusively; `when` is offered as a compact alternative for a single top-level precondition, most useful with parameterized actions.

```spectre
// These two forms are equivalent:
action deposit(amount: int) when amount > 0 {
  require balance + amount <= LIMIT
  balance' = balance + amount
}

action deposit(amount: int) {
  require amount > 0
  require balance + amount <= LIMIT
  balance' = balance + amount
}
```

---

## A.8 Statements

```ebnf
Stmt       ::= AssignStmt
             | RequireStmt
             | EnsureStmt
             | LetStmt
             | ReturnStmt
             | ExprStmt

AssignStmt ::= Ident "'" "=" Expr     (* next-state assignment *)

RequireStmt ::= "require" Expr        (* precondition — action disabled if false *)
EnsureStmt  ::= "ensure" Expr         (* postcondition — invariant on next state *)
LetStmt     ::= "let" Ident "=" Expr  (* local name binding *)
ReturnStmt  ::= "return" Expr         (* return value from fun *)
ExprStmt    ::= Expr
```

The prime `'` in `x' = e` marks a **next-state assignment**. All right-hand sides are evaluated in the pre-action state (see Appendix B).

```spectre
action transfer(amount: int) {
  require balance >= amount     // guard: evaluated in current state
  balance' = balance - amount   // next-state assignment
}
```

---

## A.9 Function declarations

```ebnf
FunDecl ::= "fun" Ident "(" [ ParamList ] ")" [ ":" TypeExpr ] "{" { Stmt } "}"
```

Functions are pure — no side effects, no state mutation. They can be called from expressions.

```spectre
fun clamp(v: int, lo: int, hi: int): int {
  if v < lo then lo else if v > hi then hi else v
}
```

---

## A.10 Invariant and temporal declarations

```ebnf
InvariantDecl ::= "invariant" Ident "{" Expr "}"

TemporalDecl  ::= "temporal" Ident "{" TemporalExpr "}"
```

```spectre
invariant solvent { balance >= 0 }
temporal  progress { WF(deposit) }
```

---

## A.11 Expressions

```ebnf
Expr ::= Expr BinOp Expr       (* binary operation *)
       | UnaryOp Expr           (* unary prefix *)
       | Expr "." Ident         (* field / method access *)
       | Expr "(" [ ArgList ] ")"  (* method or function call *)
       | Expr "[" Expr "]"      (* index access *)
       | Ident [ "'" ]          (* variable; primed = next-state read *)
       | INT | FLOAT | STRING | BOOL
       | "(" Expr ")"
       | IfExpr
       | LetExpr
       | LambdaExpr
       | SetLiteral
       | ListLiteral
       | RecordLiteral

BinOp    ::= "+" | "-" | "*" | "/" | "%"
           | "==" | "!=" | "<" | "<=" | ">" | ">="
           | "&&" | "||"
           | "in"

UnaryOp  ::= "!"          (* logical NOT *)
           | "-"          (* arithmetic negation *)

ArgList  ::= Expr { "," Expr }
```

The `in` operator checks membership: `x in S` is `true` if `x` is an element of set `S` (or a key in map `S`).

```spectre
invariant has_quorum { quorum_nodes.size() > NUM_NODES / 2 }
invariant leader_voted { leader in voted_set }
```

---

## A.12 If, let, and lambda expressions

```ebnf
IfExpr     ::= "if" Expr "then" Expr "else" Expr
             | "if" "(" Expr ")" "{" Expr "}" "else" Expr

LetExpr    ::= "let" Ident "=" Expr

LambdaExpr ::= Ident "=>" Expr
             | "(" ParamList ")" "=>" Expr
```

```spectre
let capped = if balance > LIMIT then LIMIT else balance
let double = x => x * 2
```

---

## A.13 Collection literals

```ebnf
SetLiteral    ::= "{" "}"
                | "{" Expr { "," Expr } "}"

ListLiteral   ::= "[" "]"
                | "[" Expr { "," Expr } "]"

RecordLiteral ::= "{" FieldInit { "," FieldInit } [ "," "..." Ident ] "}"
FieldInit     ::= Ident ":" Expr
```

There is no `MapLiteral` syntax. Maps are always initialized to `{}` (empty set literal) or built up via `.put(k, v)` calls. The parser disambiguates `{}` by the declared type of the variable being assigned: when the target type is `Map<K,V>`, the empty brace literal `{}` is interpreted as an empty map; when the target type is `Set<T>`, it is an empty set. Non-empty set literals `{e1, e2, ...}` are unambiguously sets — there is no key-value literal syntax for maps.

```spectre
var nodes:   Set<int>       = {1, 2, 3}  // SetLiteral (non-empty)
var history: List<str>      = ["start"]   // ListLiteral
var empty_s: Set<int>       = {}          // empty set (type drives interpretation)
var empty_m: Map<int, bool> = {}          // empty map (same syntax, type-driven)
```

---

## A.14 Temporal expressions

```ebnf
TemporalExpr ::= "always"      "(" TemporalExpr ")"
               | "eventually"  "(" TemporalExpr ")"
               | "next"        "(" TemporalExpr ")"
               | "WF"          "(" Ident ")"
               | "SF"          "(" Ident ")"
               | TemporalExpr "until" TemporalExpr
               | TemporalExpr ( "->" | "→" ) TemporalExpr
               | TemporalExpr "&&" TemporalExpr
               | TemporalExpr "||" TemporalExpr
               | "(" TemporalExpr ")"
               | Expr                                 (* state predicate *)
```

| Operator | Meaning |
|----------|---------|
| `always(P)` | P holds in every state of every execution |
| `eventually(P)` | P holds in some future state |
| `P until Q` | P holds until Q becomes true |
| `P -> Q` | Leads-to: whenever P holds, Q eventually holds |
| `WF(action)` | Weak fairness: if action is continually enabled, it fires |
| `SF(action)` | Strong fairness: if action is infinitely often enabled, it fires |

```spectre
temporal progress    { WF(increment) }
temporal completion  { eventually(phase == Phase.Committed) }
temporal stability   { always(balance >= 0) }
```

---

## A.15 Module system

```ebnf
ModuleDecl ::= "module" Ident [ "extends" Ident ] "{" { TopDecl } "}"
             | "module" Ident "=" Ident "with" "{" { Ident "=" Expr } "}"

ImportDecl ::= "import" Ident
```

- `extends` copies all declarations from the parent into the child scope (macro substitution — not runtime inheritance).
- `with` instantiates a parameterized module with concrete values.
- `import` brings a module name into scope so its members can be accessed as `ModuleName.member`.

```spectre
module Semaphore {
  var slots: int
  const MAX_SLOTS: int = 3
  action acquire { require slots > 0  slots' = slots - 1 }
  action release { require slots < MAX_SLOTS  slots' = slots + 1 }
  invariant valid { slots >= 0 && slots <= MAX_SLOTS }
}

module Mutex extends Semaphore {
  const MAX_SLOTS: int = 1    // override: binary semaphore
}
```

---

## A.16 Operator precedence

From lowest (loosest) to highest (tightest):

| Level | Operators | Associates |
|-------|-----------|------------|
| 1 (loosest) | `\|\|` (or) | left |
| 2 | `&&` (and) | left |
| 3 | `==`  `!=`  `in` | left |
| 4 | `<`  `<=`  `>`  `>=` | left |
| 5 | `+`  `-` | left |
| 6 | `*`  `/`  `%` | left |
| 7 | `!`  `-` (unary) | prefix |
| 8 | `.` (member)  `(` (call) | left |
| 9 (tightest) | `[` (index) | left |

When in doubt, add parentheses — the parser accepts them freely.

---

## A.17 Comments

```ebnf
LineComment  ::= "//" { any_char } newline
BlockComment ::= "/*" { any_char } "*/"
```

Comments have no effect on semantics and are stripped before parsing.

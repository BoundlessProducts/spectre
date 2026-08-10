# Spectre Language Definition

Formal grammar and operational semantics of the Spectre specification language.

---

## Table of Contents

1. [Lexical Structure](#1-lexical-structure)
2. [EBNF Grammar](#2-ebnf-grammar)
3. [Types](#3-types)
4. [Operator Precedence](#4-operator-precedence)
5. [Scoping and Visibility](#5-scoping-and-visibility)
6. [Operational Semantics](#6-operational-semantics)
7. [Temporal Logic Semantics](#7-temporal-logic-semantics)
8. [Type System](#8-type-system)
9. [Built-in Collection Operations](#9-built-in-collection-operations)

---

## 1. Lexical Structure

### 1.1 Character Encoding

Spectre source files are UTF-8 encoded. The Unicode arrow character `→` (U+2192) is accepted as an alias for the `->` leads-to operator.

### 1.2 Comments

```
line_comment   ::= "//" { any_char } newline
block_comment  ::= "/*" { any_char | block_comment } "*/"
```

Comments are stripped during lexing and produce no tokens.

### 1.3 Whitespace

Spaces, tabs, carriage returns, and newlines are insignificant except as token separators.

### 1.4 Tokens

#### Literals

| Token | Pattern | Notes |
|-------|---------|-------|
| `INT` | `[0-9]+` | Decimal integer |
| `FLOAT` | `[0-9]+ "." [0-9]+` | Decimal float |
| `STRING` | `'"' { char | escape } '"'` | Value stored without surrounding quotes |
| `BOOL` | `"true"` \| `"false"` | |

String escape sequences: `\"`, `\\`, `\n`, `\t`, `\r`.

#### Operators

| Token | Symbol | Meaning |
|-------|--------|---------|
| `ASSIGN` | `=` | Equality in expressions; assignment in `init`/`action` bodies |
| `EQ` | `==` | Equality (alternative to `=` in expressions) |
| `NEQ` | `!=` | Inequality |
| `LT` | `<` | Less than |
| `GT` | `>` | Greater than |
| `LEQ` | `<=` | Less than or equal |
| `GEQ` | `>=` | Greater than or equal |
| `PLUS` | `+` | Addition |
| `MINUS` | `-` | Subtraction / unary negation |
| `ASTERISK` | `*` | Multiplication |
| `SLASH` | `/` | Division |
| `AND` | `&&` | Logical conjunction |
| `OR` | `\|\|` | Logical disjunction |
| `NOT` | `!` | Logical negation (prefix) |
| `PRIME` | `'` | Next-state marker (postfix on identifiers) |
| `ARROW` | `->` or `→` | Leads-to (temporal) |
| `FATARROW` | `=>` | Lambda body separator |
| `ELLIPSIS` | `...` | Record spread |

#### Delimiters

`(` `)` `{` `}` `[` `]` `,` `.` `;` `:`

#### Keywords

| Keyword | Token |
|---------|-------|
| `module` | `MODULE` |
| `import` | `IMPORT` |
| `extends` | `EXTENDS` |
| `public` | `PUBLIC` |
| `private` | `PRIVATE` |
| `const` | `CONST` |
| `var` | `VAR` |
| `description` | `DESCRIPTION` |
| `init` | `INIT` |
| `oneOf` | `ONEOF` |
| `action` | `ACTION` |
| `fun` | `FUN` |
| `invariant` | `INVARIANT` |
| `temporal` | `TEMPORAL` |
| `require` | `REQUIRE` |
| `ensure` | `ENSURE` |
| `if` | `IF` |
| `else` | `ELSE` |
| `then` | `THEN` |
| `let` | `LET` |
| `return` | `RETURN` |
| `type` | `TYPE` |
| `enum` | `ENUM` |
| `when` | `WHEN` |
| `super` | `SUPER` |
| `with` | `WITH` |
| `always` | `ALWAYS` |
| `eventually` | `EVENTUALLY` |
| `until` | `UNTIL` |
| `WF` | `WF` |
| `SF` | `SF` |
| `next` | `NEXT` |
| `Set` | `SET` |
| `Map` | `MAP` |
| `List` | `LIST` |
| `Option` | `OPTION` |

---

## 2. EBNF Grammar

Notation: `[ X ]` = optional, `{ X }` = zero or more, `( X | Y )` = choice.

### 2.1 File

```ebnf
File          ::= { TopDecl }

TopDecl       ::= OptDesc ( VarDecl
                           | ConstDecl
                           | FunDecl
                           | ActionDecl
                           | InitDecl
                           | InvariantDecl
                           | TemporalDecl
                           | ModuleDecl
                           | ImportDecl
                           | TypeAliasDecl
                           | EnumDecl )

OptDesc       ::= [ "description" STRING ]
```

### 2.2 Declarations

```ebnf
VarDecl       ::= "var" Ident ":" Type { "," Ident ":" Type }

ConstDecl     ::= "const" Ident ":" Type "=" Expr

TypeAliasDecl ::= "type" Ident "=" TypeExpr

EnumDecl      ::= "enum" Ident "{" EnumVariant { "," EnumVariant } "}"
EnumVariant   ::= Ident

ImportDecl    ::= "import" Ident

ModuleDecl    ::= "module" Ident [ "extends" Ident ] "{" { ModuleDecl | TopDecl } "}"
               |  "module" Ident "=" Ident "with" "{" { Ident "=" Ident } "}"

InitDecl      ::= "init" ( "{" { AssignStmt } "}"
                          | OneOfBody
                          | Expr )

OneOfBody     ::= "oneOf" "{" OneOfOpt { "," OneOfOpt } "}"
OneOfOpt      ::= Expr
               |  "{" { Ident "=" Expr } "}"

ActionDecl    ::= "action" Ident [ "(" ParamList ")" ] [ "when" Expr ] "{" { Stmt } "}"
ParamList     ::= Param { "," Param }
Param         ::= Ident ":" Type

FunDecl       ::= "fun" Ident "(" [ ParamList ] ")" [ ":" Type ] "{" { Stmt } "}"

InvariantDecl ::= "invariant" Ident "{" Expr "}"

TemporalDecl  ::= "temporal" Ident "{" TemporalExpr "}"
```

### 2.3 Statements

```ebnf
Stmt          ::= AssignStmt
               |  RequireStmt
               |  EnsureStmt
               |  LetStmt
               |  ReturnStmt
               |  ExprStmt

AssignStmt    ::= PrimeIdent "=" Expr
PrimeIdent    ::= Ident "'"               (* next-state assignment *)

RequireStmt   ::= "require" Expr
EnsureStmt    ::= "ensure" Expr
LetStmt       ::= "let" Ident "=" Expr
ReturnStmt    ::= "return" Expr
ExprStmt      ::= Expr
```

### 2.4 Expressions

```ebnf
Expr          ::= BinExpr
               |  UnaryExpr
               |  PostfixExpr
               |  PrimaryExpr

BinExpr       ::= Expr BinOp Expr

BinOp         ::= "+" | "-" | "*" | "/"
               |  "=" | "==" | "!=" | "<" | ">" | "<=" | ">="
               |  "&&" | "||"

UnaryExpr     ::= ( "!" | "-" ) Expr

PostfixExpr   ::= Expr "." Ident                   (* field / method access *)
               |  Expr "[" Expr "]"                (* index *)
               |  Expr "(" [ ArgList ] ")"         (* call *)

PrimaryExpr   ::= Ident [ "'" ]                    (* primed = next-state *)
               |  INT | FLOAT | STRING | BOOL
               |  "(" Expr ")"
               |  IfExpr
               |  LetExpr
               |  LambdaExpr
               |  RecordLiteral
               |  SetLiteral
               |  ListLiteral
               |  SuperExpr

IfExpr        ::= "if" "(" Expr ")" ( "{" Expr "}" | "then" Expr ) "else" Expr
               |  "if" Expr "then" Expr "else" Expr

LetExpr       ::= "let" Ident "=" Expr

LambdaExpr    ::= Ident "=>" Expr
               |  "(" ParamList ")" "=>" Expr

RecordLiteral ::= "{" FieldInit { "," FieldInit } [ "," "..." Ident ] "}"
FieldInit     ::= Ident ":" Expr

SetLiteral    ::= "{" "}"
               |  "{" Expr { "," Expr } "}"

ListLiteral   ::= "[" "]"
               |  "[" Expr { "," Expr } "]"

SuperExpr     ::= "super" "." Ident "(" [ ArgList ] ")"

ArgList       ::= Expr { "," Expr }
```

### 2.5 Temporal Expressions

```ebnf
TemporalExpr  ::= "always" TemporalAtom
               |  "eventually" TemporalAtom
               |  "next" TemporalAtom
               |  "WF" "(" Ident ")"
               |  "SF" "(" Ident ")"
               |  TemporalAtom "until" TemporalAtom
               |  TemporalAtom ( "->" | "→" ) TemporalAtom
               |  TemporalAtom "&&" TemporalAtom
               |  TemporalAtom "||" TemporalAtom
               |  TemporalAtom

TemporalAtom  ::= "(" TemporalExpr ")"
               |  Expr                             (* state predicate *)
```

### 2.6 Types

```ebnf
TypeExpr      ::= "int"
               |  "bool"
               |  "str"
               |  "float"
               |  "Set" "<" TypeExpr ">"
               |  "Map" "<" TypeExpr "," TypeExpr ">"
               |  "List" "<" TypeExpr ">"
               |  "Option" "<" TypeExpr ">"
               |  "{" FieldDecl { "," FieldDecl } "}"   (* record type *)
               |  "(" TypeExpr { "," TypeExpr } ")"     (* tuple type *)
               |  Ident                                  (* named type / enum *)

FieldDecl     ::= Ident ":" TypeExpr
```

---

## 3. Types

### 3.1 Primitive Types

| Type | Values | Default |
|------|--------|---------|
| `int` | Arbitrary-precision integers | `0` |
| `bool` | `true`, `false` | `false` |
| `str` | Unicode strings | `""` |
| `float` | IEEE 754 double-precision | `0.0` |

### 3.2 Collection Types

| Type | Description |
|------|-------------|
| `Set<T>` | Unordered collection of unique values of type `T` |
| `List<T>` | Ordered sequence (may contain duplicates) |
| `Map<K, V>` | Finite function from keys `K` to values `V` |
| `Option<T>` | `Some(v)` for a value or `None` |

### 3.3 Structured Types

**Record type** — product of named fields:
```spectre
type Point = { x: int, y: int }
```

**Tuple type** — anonymous product:
```spectre
type Pair = (int, str)
```

**Enum type** — finite disjoint union of named variants:
```spectre
enum Color { Red, Green, Blue }
```

Enum values are referenced as `EnumName.Variant` (e.g., `Color.Red`).

### 3.4 Named Types

`type A = T` introduces `A` as an alias for type `T`. The alias is transparent — `A` and `T` are interchangeable.

---

## 4. Operator Precedence

Precedence from lowest (loosest binding) to highest (tightest binding):

| Level | Operators | Associativity |
|-------|-----------|---------------|
| 1 (lowest) | `&&` `\|\|` | Left |
| 2 | `=` `==` `!=` | Left |
| 3 | `<` `>` `<=` `>=` | Left |
| 4 | `+` `-` | Left |
| 5 | `*` `/` | Left |
| 6 | `!` `-` (unary) | Prefix |
| 7 | `(` (call), `.` (member) | Left |
| 8 (highest) | `[` (index) | Left |

**Notes:**
- The `=` token is used for equality in expressions (not assignment). Use `'` (prime) for next-state assignment: `x' = x + 1`.
- `==` is accepted as an alternative equality operator with the same precedence as `=`.
- `&&` and `||` share the same precedence level; use parentheses when mixing them.
- The `'` (prime) postfix is not an operator in the expression grammar; it is part of the identifier token in assignment positions.

---

## 5. Scoping and Visibility

### 5.1 Top-Level Scope

At file scope, the following names are in scope:
- All `var` declarations (state variables)
- All `const` declarations
- All `type` and `enum` declarations
- All `fun`, `action`, `invariant`, `temporal` declarations
- All imported module names

### 5.2 Action Scope

Inside an `action` body:
- Parameters are in scope as local read-only bindings.
- State variables are readable without qualification.
- `x'` (primed identifier) denotes the next-state value of variable `x` and may only appear on the left-hand side of an `AssignStmt` within `action` or `init` bodies.

### 5.3 Module Scope

Inside a `module` block, member names are in the module's namespace. Access from outside uses `ModuleName.memberName`.

**Visibility:**
- `public` (default): accessible outside the module.
- `private`: accessible only within the module.

Inherited members from `extends` are accessible as if declared in the child module. `super.memberName` refers to a parent member when shadowed by a child declaration.

### 5.4 Let Bindings

`let x = e` introduces `x` as a local binding within the enclosing block. `let` bindings are immutable and shadow outer names of the same identifier.

---

## 6. Operational Semantics

This section defines precisely what a Spectre specification *means* at runtime. The definitions are given in denotational style — each construct is mapped to a mathematical object — alongside plain-English notes and worked examples to build intuition.

### 6.1 State

A **state** `σ` is a total function from variable names to values:

```
σ : Var → Value
```

where `Value` is the union of all possible Spectre values.

The **initial states** are determined by the `init` declaration:
- `init { x = e₁; y = e₂; … }` — the unique initial state `σ₀` where `σ₀(x) = ⟦e₁⟧σ₀` for each assignment.
- `init oneOf { opt₁, opt₂, … }` — the set `{σ₀¹, σ₀², …}` of initial states, one per option.

**Notes:**

- A state is a *snapshot* — it captures the value of every `var` at one instant. Think of it as a row in a table where the columns are variable names. Two states are identical iff every variable has the same value in both.
- `const` declarations are **not** part of the state. They are fixed at specification time and live in a separate, immutable environment. The state only tracks things that can change.
- With `init oneOf`, the verifier starts a separate exploration from *each* initial option and checks all properties against all of them. A violation reachable from any one starting point is reported with that specific starting state shown in the trace.
- Init expressions should be closed (contain only literals and constants). Variable-to-variable dependencies in `init` are evaluated simultaneously, not sequentially.

**Example:**

```spectre
var counter: int
var status: str

init {
  counter = 0       // σ₀(counter) = 0
  status  = "idle"  // σ₀(status)  = "idle"
}
```

The single initial state is `{ counter → 0, status → "idle" }`.

### 6.2 Expression Evaluation

`⟦e⟧σ` denotes the value of expression `e` in state `σ`. All expressions are *pure* — they read the current state but never modify it.

```
⟦n⟧σ              = n                              (integer literal)
⟦s⟧σ              = s                              (string literal)
⟦b⟧σ              = b                              (bool literal)
⟦x⟧σ              = σ(x)                          (variable lookup)
⟦e₁ + e₂⟧σ       = ⟦e₁⟧σ + ⟦e₂⟧σ               (arithmetic)
⟦e₁ = e₂⟧σ       = (⟦e₁⟧σ == ⟦e₂⟧σ)             (equality — yields bool)
⟦!e⟧σ             = ¬⟦e⟧σ                         (negation)
⟦e₁ && e₂⟧σ      = ⟦e₁⟧σ ∧ ⟦e₂⟧σ               (conjunction)
⟦e₁ || e₂⟧σ      = ⟦e₁⟧σ ∨ ⟦e₂⟧σ               (disjunction)
⟦if c then e₁ else e₂⟧σ = ⟦e₁⟧σ  if ⟦c⟧σ = true
                          = ⟦e₂⟧σ  otherwise
⟦e.f⟧σ            = (⟦e⟧σ).f                     (field access)
⟦e[i]⟧σ           = (⟦e⟧σ)[⟦i⟧σ]                (index)
⟦f(e₁,…,eₙ)⟧σ    = ⟦body_f⟧σ[p₁↦⟦e₁⟧σ,…]      (function call)
⟦x => e⟧σ         = λv. ⟦e⟧σ[x↦v]               (lambda)
⟦{ f₁: e₁, … }⟧σ = record{ f₁=⟦e₁⟧σ, … }       (record literal)
⟦{ e₁, e₂, … }⟧σ = { ⟦e₁⟧σ, ⟦e₂⟧σ, … }        (set literal)
⟦[ e₁, e₂, … ]⟧σ = [ ⟦e₁⟧σ, ⟦e₂⟧σ, … ]        (list literal)
⟦{ ...r, f: e }⟧σ = ⟦r⟧σ with field f = ⟦e⟧σ   (record spread)
```

**Notes:**

- **`=` is equality in expressions, not assignment.** Inside a `require`, `ensure`, `invariant`, or temporal expression, `x = 5` means "is x equal to 5?" and produces a `bool`. The symbol `=` is only an assignment when it appears as the body of `init { x = … }` or as the right-hand side of a primed assignment `x' = …` inside an action. This dual role is a common source of confusion — when in doubt, `==` (which is unambiguously equality) can always be used instead.

- **`x'` (primed variable) is not evaluable as an expression.** The prime notation is a syntactic marker for "the next-state value" and may only appear on the *left-hand side* of an assignment statement inside `action` or `init`. Writing `y' = x' + 1` is invalid because `x'` on the right is not an expression.

- **All sub-expressions are evaluated in the same pre-state simultaneously.** In `balance + amount <= 1000000`, both `balance` and `amount` refer to their values before the transition fires. There is no sequencing within an expression.

- **Function calls are call-by-value.** Arguments are evaluated in `σ` before being substituted into the function body. `σ[p₁↦v₁]` denotes the environment `σ` extended with parameter binding `p₁ → v₁`.

- **Record spread `{ ...r, f: e }` is non-destructive.** It produces a *new* record identical to `r` except field `f` takes value `⟦e⟧σ`. The original `r` is unchanged.

- **`&&` and `||` are not short-circuiting at the semantic level.** Both sides must be well-typed. In practice the evaluator may short-circuit for performance, but the specification does not depend on evaluation order.

**Example — step-by-step evaluation:**

State: `{ balance → 500, amount → 200, limit → 1000000 }`

Expression: `balance + amount <= limit`

```
⟦balance + amount <= limit⟧σ
  = ⟦balance + amount⟧σ  <=  ⟦limit⟧σ
  = (⟦balance⟧σ + ⟦amount⟧σ)  <=  1000000
  = (500 + 200)  <=  1000000
  = 700 <= 1000000
  = true
```

### 6.3 Action Execution

An action `action A(p₁:T₁, …, pₙ:Tₙ) { S₁ … Sₖ }` applied to arguments `v₁,…,vₙ` in state `σ` produces a **next state** `σ'` as follows:

1. **Preconditions**: All `require` statements must evaluate to `true` in `σ` (with parameters bound). If any fails, the action is **disabled** in `σ` and produces no transition.

2. **Assignments**: Each `AssignStmt` `x' = e` computes `⟦e⟧σ` (in the *pre-state*, with parameters in scope) and records `σ'(x) = ⟦e⟧σ`. Variables not assigned retain their pre-state value: `σ'(y) = σ(y)` for all `y` not assigned.

3. **Postconditions**: All `ensure` statements must evaluate to `true` in `σ'`.

Formally:

```
Enabled(A, σ, v̄) ≡ ∀ require r in A: ⟦r⟧σ[p̄↦v̄] = true

Step(A, σ, v̄) = σ'  where
  σ'(x) = ⟦eₓ⟧σ[p̄↦v̄]   for each assignment x' = eₓ in A
  σ'(y) = σ(y)            for all other variables y
```

**Stuttering**: If `Step(A, σ, v̄) = σ` (the action produces the same state), the step is called a *stuttering step*. The verifier records these as warnings.

**Notes:**

- **All right-hand sides are evaluated in the pre-state.** Even if a body contains `x' = x + 1` and then `y' = x + 1`, both expressions read `x` from `σ`, not from a partially-updated `σ'`. Action bodies are *not* sequential imperative code — they describe a single simultaneous state transformation. The order of assignment statements in the body has no semantic effect.

- **Disabled actions simply don't fire.** When any `require` fails, the entire action is skipped for that `(args)` combination. This is not an error — it is the ordinary way to guard a transition. The verifier moves on to other enabled `(action, args)` pairs.

- **`ensure` is checked in the post-state `σ'`.** This lets you assert relationships between old and new values, such as `ensure counter' > counter`. Here `counter` refers to `σ(counter)` and `counter'` to `σ'(counter)`.

- **Frame condition: variables not mentioned are unchanged.** An action only specifies what changes. Everything else carries over implicitly. You do not write `y' = y` for unmodified variables.

- **Parameters are read-only locals.** Parameters cannot appear on the left-hand side of an assignment. They are bound before `require` checks and remain in scope for all right-hand side expressions in the body.

- **`when` guards are syntactic sugar** for a leading `require`. `action A when cond { … }` is exactly `action A { require cond; … }`.

**Example — tracing `deposit(amount=300)` from `{ balance → 800, txCount → 3 }`:**

```spectre
action deposit(amount: int) {
  require amount > 0
  require balance + amount <= 1000000
  balance'  = balance + amount
  txCount'  = txCount + 1
}
```

Step 1 — check preconditions in σ with `amount = 300`:
- `300 > 0` → `true` ✓
- `800 + 300 <= 1000000` → `true` ✓

Step 2 — evaluate all RHS expressions in σ (simultaneously):
- `balance + amount` = `800 + 300` = `1100`
- `txCount + 1` = `3 + 1` = `4`

Step 3 — build σ':
- `σ'(balance)` = `1100`
- `σ'(txCount)` = `4`
- `σ'(*)` = same as σ for everything else

If instead `amount = 0`: the first `require` (`0 > 0` = `false`) fails immediately — no transition occurs. `deposit(0)` is simply disabled in this state.

### 6.4 Transition System

A Spectre specification defines a **Kripke structure** `M = (S, S₀, R, L)`:

- `S` — set of all reachable states (explored by BFS/DFS).
- `S₀ ⊆ S` — initial states from `init`.
- `R ⊆ S × S` — transition relation: `(σ, σ') ∈ R` iff there exists an action `A` and arguments `v̄` such that `Enabled(A, σ, v̄)` and `Step(A, σ, v̄) = σ'`.
- `L : S → 2^{AP}` — labeling function mapping each state to the set of atomic propositions true in it.

**Notes:**

- The verifier builds `S` incrementally from `S₀` using BFS. States are fingerprinted (hashed) so each unique state is visited at most once. The frontier stops growing when no new states are discovered or a configured depth/state-count limit is reached.

- **Non-determinism is pervasive.** At any state, multiple `(action, args)` pairs may be enabled simultaneously. The verifier forks and explores *all* successors. This is what makes model checking exhaustive within the explored depth — it considers every possible interleaving and every possible argument value.

- **Parameter domains are bounded.** Because integer parameters range over an infinite domain, the verifier explores a configurable bounded range (e.g., `0..MAX_PARAM`). This is sound for finding bugs but not for proving universally-quantified correctness. Adding tight `require` guards on parameters (e.g., `require amount > 0 && amount <= 1000`) is the idiomatic way to constrain the search space without losing coverage of interesting cases.

- **State identity is by value, not by path.** If two different action sequences both lead to `{ balance → 100, txCount → 1 }`, that is one state in `S`, not two. The BFS deduplication is what prevents the state space from being infinite even for looping systems.

**Example — state space of a two-variable toggle:**

```spectre
var a: bool
var b: bool
init { a = false; b = false }
action flipA { a' = !a }
action flipB { b' = !b }
```

Full reachable `S`: `{ {F,F}, {T,F}, {F,T}, {T,T} }` — 4 states, 8 transitions. Both actions are always enabled so every state has exactly two successors.

### 6.5 Invariant Checking

An invariant `invariant I { φ }` holds iff:

```
∀ σ ∈ S: ⟦φ⟧σ = true
```

A violation is a reachable state `σ ∈ S` where `⟦φ⟧σ = false`. The verifier reports the shortest path (BFS witness) from an initial state to `σ`.

**Notes:**

- **Invariants are checked after every transition and also in `S₀`.** If the initial state itself violates an invariant, the trace has length zero — the error is flagged immediately before any action fires.

- **An invariant is different from a `require`.** A `require` inside action `A` prevents `A` from firing in a given state. An `invariant` monitors every reachable state regardless of which action produced it. A `require` in `withdraw` does not protect against a different action, say `fee`, that reduces `balance` without a guard. Only the `invariant` catches that.

- **The BFS witness is the shortest counterexample path.** Because BFS explores states in order of increasing depth, the first time a violating state is discovered it is guaranteed to be reachable via a minimal-length trace. A 3-step counterexample is much easier to debug than a 300-step one.

- **`invariant` vs `temporal always`.** Both `invariant I { φ }` and `temporal T { always φ }` say "φ holds in every reachable state." The difference is in how they are checked and reported: `invariant` is the primary safety mechanism with direct BFS counterexample generation; `temporal always` is useful when φ needs to be combined with other temporal operators (e.g., `always (p → eventually q)`).

**Key distinction — `require` vs `invariant`:**

```spectre
var balance: int
init { balance = 0 }

action withdraw(amount: int) {
  require balance >= amount   // guards THIS action only
  balance' = balance - amount
}

action fee {
  balance' = balance - 10    // no guard — can go negative!
}

invariant nonNegative {
  balance >= 0                // catches violations from ANY action
}
```

Without the `invariant`, the `fee` action would silently allow `balance` to go negative. The `require` in `withdraw` only protects that action.

### 6.6 Non-determinism

When an action has parameters, the verifier explores all values in the **parameter domain** (bounded by the configured depth/breadth limits). Multiple actions may be enabled in the same state, producing a branching computation tree.

**Notes:**

- **Parameter exploration is the main source of state-space blowup.** An action with two `int` parameters explored over range `[0, N]` produces up to `N²` branches at each state where the action is enabled. Tight `require` guards are the primary tool for managing this.

- **Non-determinism is a feature for finding bugs.** The verifier considers *every* valid `(action, args)` pair. A bug that only manifests for `amount = 0` or for a specific sequence of actions will be found because those paths are explored.

- **`oneOf` in init is independent of action non-determinism.** Each `oneOf` branch is a separate root in the exploration. The total number of states explored is at most `|init options| × |states reachable per option|`.

---

## 7. Temporal Logic Semantics

Spectre uses a fragment of **LTL (Linear Temporal Logic)** interpreted over infinite paths through the Kripke structure, plus **CTL**-style fairness conditions.

Temporal properties let you say things about *sequences of states over time*, not just individual snapshots. Where an `invariant` asks "is this true in every state?", a temporal property can ask "does this eventually become true?" or "whenever A happens, does B eventually follow?"

### 7.1 Paths

A **path** `π = σ₀ σ₁ σ₂ …` is an infinite sequence of states where each consecutive pair is connected by a transition: `(σᵢ, σᵢ₊₁) ∈ R` for all `i ≥ 0`, and `σ₀ ∈ S₀`.

`πⁱ` denotes the suffix path starting at position `i`.

**Notes:**

- Paths are *infinite* by definition. Finite executions that reach a deadlock (no enabled actions) are extended by stuttering — the final state is repeated indefinitely. This ensures the LTL semantics is well-defined even for terminating systems, but it also means a deadlocked system will *not* satisfy `eventually done` unless the deadlock state already satisfies `done`.

- The verifier does not enumerate all infinite paths literally. Instead it uses cycle detection on the finite state graph: a path is infinite iff it eventually enters a cycle. Liveness violations are found by searching for a cycle reachable from `S₀` that is also a counterexample to the property.

### 7.2 LTL Satisfaction

For a path `π` and temporal formula `φ`:

```
π ⊨ p                iff  ⟦p⟧π₀ = true           (atomic proposition)
π ⊨ ¬φ               iff  π ⊭ φ
π ⊨ φ ∧ ψ            iff  π ⊨ φ  and  π ⊨ ψ
π ⊨ φ ∨ ψ            iff  π ⊨ φ  or   π ⊨ ψ
π ⊨ always φ         iff  ∀ i ≥ 0: πⁱ ⊨ φ        (□φ)
π ⊨ eventually φ     iff  ∃ i ≥ 0: πⁱ ⊨ φ        (◇φ)
π ⊨ next φ           iff  π¹ ⊨ φ                  (Xφ)
π ⊨ φ until ψ        iff  ∃ i ≥ 0: πⁱ ⊨ ψ
                            and  ∀ j < i: πʲ ⊨ φ  (φ U ψ)
π ⊨ φ → ψ           iff  π ⊨ always (φ → eventually ψ)
                          (leads-to: □(φ → ◇ψ))
```

A temporal property `temporal P { φ }` holds iff:

```
∀ fair paths π from S₀: π ⊨ φ
```

**Notes on individual operators:**

- **`always φ` (□φ) — safety.** The formula `φ` must hold at every position of every path. This is the temporal version of an invariant. `temporal T { always (balance >= 0) }` and `invariant I { balance >= 0 }` are semantically equivalent safety properties; the `invariant` form is preferred for pure safety because it has more direct counterexample generation.

- **`eventually φ` (◇φ) — liveness.** There must exist some step on every path where `φ` becomes true. This cannot be verified by checking a single state — the verifier must ensure no infinite path avoids `φ` forever. A system that deadlocks before `φ` becomes true will fail `eventually φ` (because the infinite stuttering extension of the deadlock never reaches `φ`).

- **`next φ` (Xφ) — one-step lookahead.** `φ` must hold in the *immediately following* state. This is rarely needed in high-level specs but is useful for relating consecutive states (e.g., `next counter = counter + 1`).

- **`φ until ψ` (φ U ψ) — conditional progress.** `φ` must hold continuously until some point where `ψ` holds, and `ψ` *must* eventually hold. This is *strong until* — the `ψ` side is required to arrive. Example: `counter < 10 until counter = 10` says the counter stays below 10 until it reaches exactly 10 (and it must reach 10).

- **`φ → ψ` (leads-to, □(φ→◇ψ)) — response.** "Whenever `φ` becomes true, `ψ` will eventually become true." This is the most commonly used liveness pattern. It is *not* the propositional `→` (implication); the arrow here is the temporal leads-to operator, which is only valid inside a `temporal` block. Example: `requestSent → responseReceived` says every request is eventually answered.

  > **Gotcha:** `→` is syntactic sugar for `always (φ → eventually ψ)`, not just `φ → eventually ψ`. The outer `always` makes it apply to *every* point in time where `φ` becomes true, not just the first one.

- **Nesting operators.** `always eventually φ` means "at every point, there is some future point where φ holds" — i.e., φ recurs infinitely often. `eventually always φ` means "from some point on, φ holds forever" — i.e., φ is eventually stable. These are different properties and have different verification complexity.

**Example — tracing satisfaction of `eventually (counter = 3)`:**

Path: `{ counter=0 } → { counter=1 } → { counter=2 } → { counter=3 } → …`

- At position 0: `counter = 3`? No.
- At position 1: `counter = 3`? No.
- At position 2: `counter = 3`? No.
- At position 3: `counter = 3`? Yes. ∃ i=3 ≥ 0 where πⁱ ⊨ `counter = 3`. ✓

Property holds on this path.

**Example — `always eventually (counter = 0)` (counter keeps resetting):**

Every suffix of the path must eventually reach `counter = 0`. This fails on a path like `{ 0, 1, 2, 3, 4, … }` that only increments forever, because the suffix starting at position 1 never sees `counter = 0` again.

### 7.3 Fairness

Without fairness, a liveness property like `eventually done` can fail on a path where some action is always enabled but the scheduler just never picks it. Fairness conditions rule out such *unfair* paths.

**Weak fairness** `WF(A)` on action `A` requires that `A` does not remain continuously enabled without eventually executing:

```
WF(A) on path π  iff
  ¬(∃ i: ∀ j ≥ i: Enabled(A, πⱼ) ∧ A not taken at step j)
```

Equivalently: every suffix in which `A` is *continuously* enabled must contain a step where `A` fires.

**Strong fairness** `SF(A)` requires that if `A` is enabled infinitely often, it executes infinitely often:

```
SF(A) on path π  iff
  (∀ i: ∃ j ≥ i: Enabled(A, πⱼ)) → (∀ i: ∃ j ≥ i: A is taken at step j)
```

`WF(v)` / `SF(v)` for a variable name `v` applies the respective fairness to every action that contains `v' = …` in its body.

**Usage in temporal properties:**

- `temporal P { WF(A) }` — asserts that all paths are weakly fair for `A`. Standalone; rarely enough on its own.
- `temporal P { WF(A) → φ }` — asserts that `φ` holds on all weakly-fair paths for `A`. Unfair paths are filtered out before checking `φ`.

**Notes:**

- **The key intuition for weak vs. strong fairness:**
  - *Weak*: "if the action is *always* available from some point on, it will eventually run." Use WF when the action stays permanently enabled once conditions are met (e.g., a process waiting for a lock that is never released back).
  - *Strong*: "if the action becomes available *infinitely often*, it will eventually run." Use SF when the action's enablement may flicker — it is available, then not, then available again — and you still want to guarantee it eventually fires.

- **Why liveness needs fairness.** Consider an `increment` action with no `require` (always enabled). The path `{ 0, 0, 0, … }` — which just repeatedly fires a no-op `reset` action and never increments — satisfies all safety invariants but violates `eventually counter = 10`. Adding `temporal T { WF(increment) }` eliminates paths where `increment` is perpetually skipped while enabled, and `eventually counter = 10` then holds on all remaining (fair) paths.

- **`WF(A) → φ` filters paths, it does not change the system.** The transition graph `R` is unmodified. The verifier simply ignores paths that violate the fairness condition when checking `φ`. This means fairness hypotheses only strengthen liveness properties — they cannot cause a safety invariant to be discharged.

- **`WF(variable)` is shorthand.** `WF(balance)` applies weak fairness to every action that contains `balance' = …`. This is convenient when you want to say "the balance will eventually change" without listing all modifying actions by name.

**Example — why fairness is needed for progress:**

```spectre
var counter: int
init { counter = 0 }

action increment { counter' = counter + 1 }
action stay      { counter' = counter }     // stuttering no-op

temporal progress { eventually (counter = 5) }
```

Without fairness, `progress` fails: the path `stay, stay, stay, …` keeps `counter` at 0 forever. Adding:

```spectre
temporal fairIncrement { WF(increment) }
```

rules out that path (because `increment` is always enabled but never taken). Under weak fairness, `eventually counter = 5` holds on all remaining paths.

### 7.4 Temporal Operator Summary

| Spectre syntax | LTL equivalent | Description |
|---------------|----------------|-------------|
| `always φ` | `□φ` | `φ` holds at every step |
| `eventually φ` | `◇φ` | `φ` holds at some step |
| `next φ` | `Xφ` | `φ` holds at the next step |
| `φ until ψ` | `φ U ψ` | `φ` holds until `ψ` becomes true |
| `φ → ψ` | `□(φ → ◇ψ)` | whenever `φ`, eventually `ψ` |
| `WF(A)` | — | weak fairness for action `A` |
| `SF(A)` | — | strong fairness for action `A` |

---

## 8. Type System

### 8.1 Typing Judgments

`Γ ⊢ e : T` — expression `e` has type `T` under context `Γ`.

The context `Γ` maps variable names and parameter names to their declared types.

### 8.2 Typing Rules (selected)

**Variable:**
```
x : T ∈ Γ
──────────
Γ ⊢ x : T
```

**Primed identifier** (next-state, only in `action`/`init` bodies):
```
x : T ∈ Γ
────────────
Γ ⊢ x' : T
```

**Arithmetic:**
```
Γ ⊢ e₁ : int    Γ ⊢ e₂ : int
──────────────────────────────
Γ ⊢ e₁ + e₂ : int
```

**Comparison:**
```
Γ ⊢ e₁ : T    Γ ⊢ e₂ : T    T comparable
──────────────────────────────────────────
Γ ⊢ e₁ = e₂ : bool
```

**Record literal:**
```
Γ ⊢ e₁ : T₁  …  Γ ⊢ eₙ : Tₙ
───────────────────────────────────────────
Γ ⊢ { f₁: e₁, …, fₙ: eₙ } : { f₁: T₁, …, fₙ: Tₙ }
```

**Field access:**
```
Γ ⊢ e : { …, f: T, … }
───────────────────────
Γ ⊢ e.f : T
```

**Set literal:**
```
Γ ⊢ e₁ : T  …  Γ ⊢ eₙ : T
────────────────────────────
Γ ⊢ { e₁, …, eₙ } : Set<T>
```

**List literal:**
```
Γ ⊢ e₁ : T  …  Γ ⊢ eₙ : T
────────────────────────────
Γ ⊢ [ e₁, …, eₙ ] : List<T>
```

**Lambda:**
```
Γ[x↦T] ⊢ e : U
────────────────────────────
Γ ⊢ (x : T) => e : T → U
```

**Enum constructor:**
```
enum E { …, V, … }  ∈ Γ
────────────────────────
Γ ⊢ E.V : E
```

### 8.3 Action Well-Formedness

An action body is well-formed if:
1. All `require` expressions have type `bool`.
2. All `ensure` expressions have type `bool`.
3. For each `AssignStmt` `x' = e`: `x : T ∈ Γ` and `Γ ⊢ e : T`.
4. No `x'` appears in an expression position (only as the left-hand side of assignment).

### 8.4 Invariant Well-Formedness

`invariant I { φ }` is well-formed if `Γ_global ⊢ φ : bool`, where `Γ_global` contains only state variables and constants (no next-state variables, no action parameters).

---

## 9. Built-in Collection Operations

### 9.1 Set Operations

For `s : Set<T>`, `t : Set<T>`, `x : T`, `p : T → bool`, `f : T → U`:

| Method | Type | Semantics |
|--------|------|-----------|
| `s.union(t)` | `Set<T>` | `{x | x ∈ s ∨ x ∈ t}` |
| `s.intersection(t)` | `Set<T>` | `{x | x ∈ s ∧ x ∈ t}` |
| `s.contains(x)` | `bool` | `x ∈ s` |
| `s.size()` | `int` | `|s|` |
| `s.forall(p)` | `bool` | `∀ x ∈ s: p(x)` |
| `s.exists(p)` | `bool` | `∃ x ∈ s: p(x)` |
| `s.filter(p)` | `Set<T>` | `{x ∈ s | p(x)}` |
| `s.map(f)` | `Set<U>` | `{f(x) | x ∈ s}` |
| `s.toList()` | `List<T>` | arbitrary order |
| `Set.empty()` | `Set<T>` | `∅` (same as `{}`) |
| `Set.of(x)` | `Set<T>` | `{x}` |

### 9.2 List Operations

For `l : List<T>`, `x : T`, `p : T → bool`, `f : T → U`:

| Method | Type | Semantics |
|--------|------|-----------|
| `l.append(x)` | `List<T>` | `l ++ [x]` |
| `l.head()` | `T` | first element (undefined on empty) |
| `l.tail()` | `List<T>` | all but first element |
| `l.size()` | `int` | `len(l)` |
| `l.filter(p)` | `List<T>` | elements where `p` holds, order preserved |
| `l.map(f)` | `List<U>` | element-wise application of `f` |
| `l.reduce(z, f)` | `U` | left fold: `f(f(…f(z, l[0]), l[1])…, l[n-1])` |
| `l.forall(p)` | `bool` | `∀ x ∈ l: p(x)` |
| `l.exists(p)` | `bool` | `∃ x ∈ l: p(x)` |
| `l.toSet()` | `Set<T>` | deduplication |
| `List.empty()` | `List<T>` | `[]` |
| `List.of(x)` | `List<T>` | `[x]` |

### 9.3 Map Operations

For `m : Map<K, V>`, `k : K`, `v : V`:

| Method | Type | Semantics |
|--------|------|-----------|
| `m.put(k, v)` | `Map<K, V>` | `m` with `k` mapped to `v` |
| `m.get(k)` | `V` | value at `k` (undefined if absent) |
| `m.contains(k)` | `bool` | `k ∈ dom(m)` |
| `m.keys()` | `Set<K>` | `dom(m)` |
| `m.values()` | `List<V>` | range of `m` (arbitrary order) |
| `Map.empty()` | `Map<K, V>` | empty function |

---

## Appendix A: Grammar Summary (EBNF Quick Reference)

```ebnf
(* Top-level *)
File       ::= { TopDecl }
TopDecl    ::= OptDesc Decl
OptDesc    ::= [ "description" STRING ]
Decl       ::= VarDecl | ConstDecl | TypeAliasDecl | EnumDecl
             | ImportDecl | ModuleDecl | InitDecl | ActionDecl
             | FunDecl | InvariantDecl | TemporalDecl

(* Declarations *)
VarDecl    ::= "var" Ident ":" TypeExpr
ConstDecl  ::= "const" Ident ":" TypeExpr "=" Expr
EnumDecl   ::= "enum" Ident "{" Ident { "," Ident } "}"
TypeDecl   ::= "type" Ident "=" TypeExpr
ImportDecl ::= "import" Ident
ModuleDecl ::= "module" Ident [ "extends" Ident ] "{" { TopDecl } "}"
InitDecl   ::= "init" ( "{" { AssignStmt } "}" | OneOfBody | Expr )
ActionDecl ::= "action" Ident [ "(" ParamList ")" ] [ "when" Expr ] "{" { Stmt } "}"
FunDecl    ::= "fun" Ident "(" [ ParamList ] ")" [ ":" TypeExpr ] "{" { Stmt } "}"
InvDecl    ::= "invariant" Ident "{" Expr "}"
TempDecl   ::= "temporal" Ident "{" TemporalExpr "}"

(* Statements *)
Stmt       ::= Ident "'" "=" Expr        (* next-state assign *)
             | "require" Expr
             | "ensure" Expr
             | "let" Ident "=" Expr
             | "return" Expr
             | Expr

(* Types *)
TypeExpr   ::= "int" | "bool" | "str" | "float"
             | "Set" "<" TypeExpr ">"
             | "Map" "<" TypeExpr "," TypeExpr ">"
             | "List" "<" TypeExpr ">"
             | "Option" "<" TypeExpr ">"
             | "{" Ident ":" TypeExpr { "," Ident ":" TypeExpr } "}"
             | "(" TypeExpr { "," TypeExpr } ")"
             | Ident

(* Expressions — precedence shown in §4 *)
Expr       ::= Expr BinOp Expr
             | ( "!" | "-" ) Expr
             | Expr "." Ident
             | Expr "[" Expr "]"
             | Expr "(" ArgList ")"
             | Ident [ "'" ]
             | INT | FLOAT | STRING | BOOL
             | "(" Expr ")"
             | IfExpr | LambdaExpr | RecordLit | SetLit | ListLit

(* Temporal *)
TemporalExpr ::= "always" TemporalAtom
               | "eventually" TemporalAtom
               | "next" TemporalAtom
               | "WF" "(" Ident ")"   | "SF" "(" Ident ")"
               | TemporalAtom "until" TemporalAtom
               | TemporalAtom ( "->" | "→" ) TemporalAtom
               | TemporalAtom
TemporalAtom ::= "(" TemporalExpr ")" | Expr
```

---

## Appendix B: Reserved Words

All of the following identifiers are reserved and cannot be used as user-defined names:

```
action    always    const     description  else     ensure    enum
eventually extends  false     fun          if       import    init
invariant let       List      Map          module   next      oneOf
Option    private   public    require      return   SF        Set
super     temporal  then      true         type     until     var
when      with      WF
```

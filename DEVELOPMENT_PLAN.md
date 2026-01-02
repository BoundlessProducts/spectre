# Spectre Language Development Plan

## Overview
Build a complete implementation of the Spectre specification language using Go, including parser, type checker, semantic analyzer, and verification engine. Each phase is designed to be atomically testable - tests must pass before moving to the next step.

---

## Phase 1: Foundation & Lexer (Testable: Token Recognition)

**Step 1.1: Project Setup** : Initialize Go module, create project structure (cmd/spectre, pkg/, internal/lexer, internal/parser, etc.), add go.mod and basic README

**Step 1.2: Token Definitions** : Define Token struct with Type, Value, Position fields and all token types (keywords, operators, literals, identifiers)

**Step 1.3: Lexer Core** : Implement basic lexer that can tokenize simple identifiers, keywords, and whitespace

**Step 1.4: Literal Tokenization** : Add support for integer, boolean, string, and float literals

**Step 1.5: Operator Tokenization** : Tokenize all operators (+, -, *, /, =, !=, <, >, <=, >=, &&, ||, !, →, ', etc.)

**Step 1.6: Keyword Recognition** : Recognize all Spectre keywords (var, const, action, init, invariant, temporal, etc.)

**Step 1.7: Comment Handling** : Handle single-line (//) and multi-line comments

**Step 1.8: Error Handling** : Report lexer errors for invalid characters and unterminated strings/comments

**Step 1.9: Lexer Testing** : Create comprehensive test suite - all tests must pass before proceeding

**Test Criteria**: Can tokenize all example .spec files correctly, reports errors for invalid input

---

## Phase 2: Parser - Basic Constructs (Testable: AST Generation)

**Step 2.1: AST Node Definitions** : Define AST node interfaces and structs for types, expressions, and basic declarations

**Step 2.2: Expression Parser** : Parse arithmetic, comparison, logical expressions with proper precedence

**Step 2.3: Type Parser** : Parse primitive types, records, sets, maps, lists, enums, optionals

**Step 2.4: Variable Declaration Parser** : Parse var declarations with types and descriptions

**Step 2.5: Constant Parser** : Parse const declarations with values and descriptions

**Step 2.6: Description Parser** : Parse description fields attached to any element

**Step 2.7: Parser Testing (Basic)** : Test parsing of types, variables, constants - verify AST structure

**Test Criteria**: Can parse all type definitions and variable/constant declarations from examples

---

## Phase 3: Parser - Control Flow & Functions (Testable: Complex AST)

**Step 3.1: If-Else Parser** : Parse conditional expressions and if-else blocks

**Step 3.2: Let Binding Parser** : Parse let expressions for local variable bindings

**Step 3.3: Pure Function Parser** : Parse fun declarations with parameters, return types, and bodies

**Step 3.4: Function Call Parser** : Parse function calls in expressions

**Step 3.5: Collection Operations Parser** : Parse map, filter, reduce, union, intersection operations

**Step 3.6: Parser Testing (Functions)** : Test parsing of functions, conditionals, and complex expressions

**Test Criteria**: Can parse pure-functions.spec example file completely

---

## Phase 4: Parser - Actions & Initial States (Testable: State Machine AST)

**Step 4.1: Init Parser** : Parse init blocks and oneOf initial states

**Step 4.2: Action Parser** : Parse action declarations with parameters and bodies

**Step 4.3: Prime Notation Parser** : Parse next-state assignments (variable' = expression)

**Step 4.4: Require/Ensure Parser** : Parse preconditions (require) and postconditions (ensure)

**Step 4.5: Action Guard Parser** : Parse action guards (action name when condition)

**Step 4.6: Parser Testing (Actions)** : Test parsing of all actions from example files

**Test Criteria**: Can parse counter.spec, mutex.spec, and all action-related constructs

---

## Phase 5: Parser - Constraints & Temporal (Testable: Property AST)

**Step 5.1: Invariant Parser** : Parse invariant declarations with descriptions

**Step 5.2: Temporal Parser** : Parse temporal properties with always, eventually, until operators

**Step 5.3: Fairness Parser** : Parse WF() and SF() fairness operators

**Step 5.4: Leads-To Parser** : Parse → (leads to) operator in temporal properties

**Step 5.5: Parser Testing (Properties)** : Test parsing of all invariants and temporal properties

**Test Criteria**: Can parse all constraints and temporal properties from example files

---

## Phase 6: Parser - Modules (Testable: Module AST)

**Step 6.1: Module Parser** : Parse module definitions with public/private modifiers

**Step 6.2: Import Parser** : Parse import statements

**Step 6.3: Extends Parser** : Parse module extensions and super calls

**Step 6.4: Module Instance Parser** : Parse module instances with parameter substitution

**Step 6.5: Parser Testing (Modules)** : Test parsing of modules-example.spec and module constructs

**Test Criteria**: Can parse all module-related syntax correctly

---

## Phase 7: Parser Integration (Testable: Complete File Parsing)

**Step 7.1: File Parser** : Integrate all parsers to parse complete .spec files

**Step 7.2: Error Recovery** : Implement parser error recovery to continue parsing after errors

**Step 7.3: Error Reporting** : Report parse errors with line numbers and helpful messages

**Step 7.4: Integration Testing** : Parse all example .spec files - verify no parse errors

**Test Criteria**: All example files parse successfully, errors reported correctly for invalid syntax

---

## Phase 8: Type System - Core (Testable: Type Checking)

**Step 8.1: Type Representation** : Implement type system with all type kinds (primitive, record, set, map, list, enum, option)

**Step 8.2: Type Environment** : Build type environment for tracking variable and function types

**Step 8.3: Expression Type Checking** : Type-check expressions and infer types

**Step 8.4: Assignment Type Checking** : Verify assignments have compatible types

**Step 8.5: Type Checking Testing** : Test type checker with valid and invalid type scenarios

**Test Criteria**: Catches all type errors, accepts all valid types from examples

---

## Phase 9: Type System - Advanced (Testable: Complex Types)

**Step 9.1: Record Type Checking** : Type-check record field access and updates

**Step 9.2: Collection Type Checking** : Type-check set, map, list operations

**Step 9.3: Function Type Checking** : Type-check function calls with parameter and return types

**Step 9.4: Type Inference** : Infer types for let bindings and function returns

**Step 9.5: Type Testing (Advanced)** : Test complex type scenarios from examples

**Test Criteria**: Handles all type scenarios in user-management.spec and bank-account.spec

---

## Phase 10: Semantic Analysis (Testable: Semantic Validation)

**Step 10.1: Symbol Table** : Build symbol table with scoping for modules, constants, variables, functions

**Step 10.2: Name Resolution** : Resolve identifiers, qualified names (Module.member), and imports

**Step 10.3: Variable Validation** : Verify all variables are declared before use

**Step 10.4: Function Validation** : Verify function calls match declarations, check purity constraints

**Step 10.5: Semantic Testing** : Test name resolution and semantic rules

**Test Criteria**: Catches undefined variables, invalid references, and scope violations

---

## Phase 11: Module System Analysis (Testable: Module Resolution)

**Step 11.1: Module Resolution** : Resolve module imports and extensions

**Step 11.2: Visibility Checking** : Verify public/private access rules

**Step 11.3: Inheritance Analysis** : Analyze module extension and super calls

**Step 11.4: Module Testing** : Test module system with modules-example.spec

**Test Criteria**: Resolves all module references correctly, enforces visibility rules

---

## Phase 12: State Machine Model (Testable: State Representation)

**Step 12.1: State Variable Model** : Create representation of all state variables with types

**Step 12.2: Initial State Model** : Represent deterministic and oneOf initial states

**Step 12.3: Action Model** : Represent actions as state transitions with guards and updates

**Step 12.4: Constraint Model** : Represent invariants, preconditions, postconditions

**Step 12.5: Model Testing** : Verify state machine model captures all example specifications

**Test Criteria**: Can represent state machines from all example files correctly

---

## Phase 13: Pure Function Evaluation (Testable: Function Execution)

**Step 13.1: Function Environment** : Build environment for evaluating pure functions

**Step 13.2: Expression Evaluation** : Evaluate expressions in function context

**Step 13.3: Recursion Support** : Handle recursive function calls

**Step 13.4: Purity Verification** : Verify functions don't access state variables

**Step 13.5: Function Testing** : Test function evaluation with pure-functions.spec

**Test Criteria**: Evaluates all pure functions correctly, catches purity violations

---

## Phase 14: Verification - Invariants (Testable: Invariant Checking)

**Step 14.1: State Representation** : Create efficient state representation for verification

**Step 14.2: Expression Evaluation (State)** : Evaluate expressions in state context

**Step 14.3: Invariant Evaluation** : Evaluate invariants in given state

**Step 14.4: State Transition** : Apply action to state and produce next state

**Step 14.5: Invariant Testing** : Test invariant checking on example files

**Test Criteria**: Correctly identifies when invariants are violated, handles all invariant types

---

## Phase 15: Verification - State Exploration (Testable: State Space)

**Step 15.1: State Hashing** : Implement state hashing for efficient comparison

**Step 15.2: State Space Traversal** : Build BFS/DFS state space exploration

**Step 15.3: Cycle Detection** : Detect cycles in state space

**Step 15.4: Multiple Initial States** : Handle oneOf by exploring all initial states

**Step 15.5: Exploration Testing** : Test state exploration on small examples

**Test Criteria**: Explores all reachable states correctly, handles cycles and multiple initials

---

## Phase 16: Verification - Temporal Properties (Testable: Temporal Checking)

**Step 16.1: Execution Trace** : Build system to track execution traces

**Step 16.2: Temporal Operator Evaluation** : Implement always, eventually, until evaluation

**Step 16.3: Leads-To Evaluation** : Implement → (leads to) operator

**Step 16.4: Temporal Testing** : Test temporal property checking

**Test Criteria**: Correctly verifies temporal properties, identifies violations with traces

---

## Phase 17: Verification - Fairness (Testable: Fairness Checking)

**Step 17.1: Action Enabledness** : Determine when actions are enabled

**Step 17.2: Weak Fairness** : Implement WF() checking

**Step 17.3: Strong Fairness** : Implement SF() checking

**Step 17.4: Fairness Testing** : Test fairness conditions on mutex and fairness examples

**Test Criteria**: Correctly implements fairness semantics, verifies fairness properties

---

## Phase 18: Error Reporting (Testable: Error Messages)

**Step 18.1: Error Context** : Capture source locations and descriptions for errors

**Step 18.2: Error Formatting** : Format user-friendly error messages using descriptions

**Step 18.3: Trace Generation** : Generate execution traces showing path to error

**Step 18.4: Error Testing** : Verify error messages are helpful and include descriptions

**Test Criteria**: Error messages include descriptions, traces are readable and helpful

---

## Phase 19: CLI Tool (Testable: Command Execution)

**Step 19.1: Command Structure** : Implement CLI with parse, typecheck, verify commands

**Step 19.2: File Processing** : Handle single file and multi-file processing

**Step 19.3: Output Formatting** : Format results for console output

**Step 19.4: CLI Testing** : Test CLI with all example files

**Test Criteria**: CLI works correctly, produces expected output for all commands

---

## Phase 20: Integration & Polish (Testable: End-to-End)

**Step 20.1: End-to-End Testing** : Test complete flow on all example files

**Step 20.2: Performance Testing** : Profile and identify bottlenecks

**Step 20.3: Documentation** : Write user guide and API docs

**Step 20.4: Final Validation** : Verify all success criteria are met

**Test Criteria**: All example files verify correctly, performance is acceptable, docs are complete

---

## Testing Strategy

- **Unit Tests**: Each step has focused unit tests
- **Integration Tests**: Each phase has integration tests using example files
- **Test Files**: Use all .spec files in examples/ directory as test cases
- **Pass Criteria**: All tests in a phase must pass before moving to next step

## Success Criteria

1. ✅ Can parse all Spectre syntax correctly (Phase 7)
2. ✅ Type checking catches all type errors (Phase 9)
3. ✅ Semantic analysis validates all rules (Phase 11)
4. ✅ Verifies invariants correctly (Phase 14)
5. ✅ Verifies temporal properties correctly (Phase 16)
6. ✅ Generates helpful error messages with descriptions (Phase 18)
7. ✅ CLI tool works with all example files (Phase 19)
8. ✅ All example files verify successfully (Phase 20)

## Dependencies

- Go 1.21+ (for generics support)
- Standard Go testing framework
- Example .spec files in examples/ directory

---

## Execution Rules

1. **One step at a time**: Complete and test each step before moving to next
2. **Tests must pass**: All tests in current phase must pass before next step
3. **Atomic phases**: Each phase is independently testable
4. **Use examples**: Test with actual .spec files from examples/ directory
5. **Incremental**: Build on previous phases, don't skip ahead

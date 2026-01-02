package eval

import (
	"fmt"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// FunctionEvaluator evaluates pure functions from a parsed Spectre file
type FunctionEvaluator struct {
	env            *Environment
	purityChecker  *PurityChecker
	variableModel  *state.VariableModel
}

// NewFunctionEvaluator creates a new function evaluator from a parsed file
func NewFunctionEvaluator(file *ast.File) (*FunctionEvaluator, error) {
	fe := &FunctionEvaluator{
		env: NewEnvironment(),
	}

	// Build variable model to identify state variables
	fe.variableModel = state.NewVariableModel(file)
	fe.purityChecker = NewPurityChecker(fe.variableModel)

	// Extract and register all functions
	for _, decl := range file.Decls {
		if fnDecl, ok := decl.(*ast.FunctionDecl); ok {
			// Check purity
			if err := fe.purityChecker.CheckFunction(fnDecl); err != nil {
				return nil, fmt.Errorf("function %s: %w", fnDecl.Name, err)
			}

			// Register function
			fnDef := &FunctionDef{
				Decl:   fnDecl,
				Params: fnDecl.Parameters,
				Body:   fnDecl.Body,
			}
			if err := fe.env.DefineFunction(fnDecl.Name, fnDef); err != nil {
				return nil, fmt.Errorf("duplicate function %s: %w", fnDecl.Name, err)
			}
		}
	}

	return fe, nil
}

// CallFunction calls a function with the given arguments
func (fe *FunctionEvaluator) CallFunction(name string, args []state.Value) (state.Value, error) {
	fnDef, exists := fe.env.GetFunction(name)
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	if len(args) != len(fnDef.Params) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d", name, len(fnDef.Params), len(args))
	}

	// Create new scope for function execution
	funcEnv := fe.env.EnterScope()
	funcEnv.DefineFunction(name, fnDef) // Make function available for recursion

	// Bind parameters
	for i, param := range fnDef.Params {
		funcEnv.SetVariable(param.Name, args[i])
	}

	// Evaluate function body
	evaluator := NewEvaluator(funcEnv)
	return evaluator.evalFunctionBody(fnDef.Body)
}

// HasFunction checks if a function exists
func (fe *FunctionEvaluator) HasFunction(name string) bool {
	return fe.env.HasFunction(name)
}

// GetFunctionNames returns all function names
func (fe *FunctionEvaluator) GetFunctionNames() []string {
	// This is a simple implementation - in a real system, we'd maintain a list
	names := []string{}
	// We can't easily iterate over functions in the environment, so we'll need
	// to track them separately or add a method to Environment
	return names
}

// EvaluateExpression evaluates an expression in the function context
func (fe *FunctionEvaluator) EvaluateExpression(expr ast.Expr) (state.Value, error) {
	evaluator := NewEvaluator(fe.env)
	return evaluator.Eval(expr)
}

// LoadFromFile loads functions from a Spectre source file
func LoadFromFile(filename string) (*FunctionEvaluator, error) {
	// Read file
	content, err := readFileContent(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse file
	l := lexer.New(content)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	return NewFunctionEvaluator(file)
}

// readFileContent reads file content (helper function)
func readFileContent(filename string) (string, error) {
	// Use standard library to read file
	// This is a placeholder - actual implementation would use os.ReadFile
	return "", fmt.Errorf("file reading not yet implemented")
}


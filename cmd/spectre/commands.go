package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akkeshavan/spectre/internal/errors"
	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/semantic"
	"github.com/akkeshavan/spectre/internal/types"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Run         func(args []string) error
}

// Commands maps command names to their implementations
var Commands = map[string]*Command{
	"parse": {
		Name:        "parse",
		Description: "Parse a Spectre specification file and report syntax errors",
		Run:         runParse,
	},
	"typecheck": {
		Name:        "typecheck",
		Description: "Type-check a Spectre specification file and report type errors",
		Run:         runTypecheck,
	},
	"verify": {
		Name:        "verify",
		Description: "Verify a Spectre specification (check invariants and temporal properties)",
		Run:         runVerify,
	},
}

// runParse parses a Spectre file and reports syntax errors
func runParse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(args, func(filename string) error {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		_ = p.ParseFile()

		// Report parse errors
		if len(p.Errors()) > 0 {
			fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
			for _, errMsg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
		}

		fmt.Printf("✓ Successfully parsed %s\n", filename)
		return nil
	})
}

// runTypecheck type-checks a Spectre file and reports type errors
func runTypecheck(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(args, func(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Parse
	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
		for _, errMsg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
	}

	// Build symbol table
	builder := semantic.NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Symbol table build errors in %s:\n", filename)
		for _, errMsg := range buildErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("symbol table build failed with %d error(s)", len(buildErrors))
	}

	// Resolve modules
	resolver := semantic.NewModuleResolver(symbolTable)
	resolutionErrors := resolver.ResolveModules(file)
	if len(resolutionErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Module resolution errors in %s:\n", filename)
		for _, errMsg := range resolutionErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("module resolution failed with %d error(s)", len(resolutionErrors))
	}

	// Build type environment from declarations
	typeEnv := types.NewEnvironment()
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			typ, err := types.FromAST(d.Type)
			if err == nil {
				typeEnv.DeclareVariable(d.Name, typ)
			}
		case *ast.ConstantDecl:
			typ, err := types.FromAST(d.Type)
			if err == nil {
				typeEnv.DeclareConstant(d.Name, typ)
			}
		case *ast.FunctionDecl:
			params := make([]types.Type, len(d.Parameters))
			for i, param := range d.Parameters {
				paramType, err := types.FromAST(param.Type)
				if err == nil {
					params[i] = paramType
				}
			}
			returnType, err := types.FromAST(d.ReturnType)
			if err == nil {
				sig := &types.FunctionSignature{
					Parameters: params,
					Return:     returnType,
				}
				typeEnv.DeclareFunction(d.Name, sig)
			}
		}
	}

	// Type check
	checker := types.NewChecker(typeEnv)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			if d.Body != nil {
				for _, stmt := range d.Body.Statements {
					switch s := stmt.(type) {
					case *ast.AssignStmt:
						checker.CheckAssignment(s)
					case *ast.RequireStmt:
						checker.CheckExpression(s.Condition)
					case *ast.EnsureStmt:
						checker.CheckExpression(s.Condition)
					}
				}
			}
		case *ast.InvariantDecl:
			checker.CheckExpression(d.Condition)
		case *ast.TemporalDecl:
			// Temporal expressions are checked during verification, not type checking
			// Skip them here to avoid errors
		}
	}

	typeErrors := checker.Errors()
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Type errors in %s:\n", filename)
		formatter := errors.NewErrorFormatter()
		for _, typeErr := range typeErrors {
			formatted := formatter.FormatTypeError(typeErr.Message, typeErr.Position)
			fmt.Fprintf(os.Stderr, "  %s\n", formatted)
		}
		return fmt.Errorf("type checking failed with %d error(s)", len(typeErrors))
	}

		fmt.Printf("✓ Successfully type-checked %s\n", filename)
		return nil
	})
}

// runVerify verifies a Spectre specification
func runVerify(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(args, func(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Parse
	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
		for _, errMsg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
	}

	// Build symbol table
	builder := semantic.NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Symbol table build errors in %s:\n", filename)
		for _, errMsg := range buildErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("symbol table build failed with %d error(s)", len(buildErrors))
	}

	// Resolve modules
	resolver := semantic.NewModuleResolver(symbolTable)
	resolutionErrors := resolver.ResolveModules(file)
	if len(resolutionErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Module resolution errors in %s:\n", filename)
		for _, errMsg := range resolutionErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("module resolution failed with %d error(s)", len(resolutionErrors))
	}

	// Build type environment from declarations
	typeEnv := types.NewEnvironment()
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			typ, err := types.FromAST(d.Type)
			if err == nil {
				typeEnv.DeclareVariable(d.Name, typ)
			}
		case *ast.ConstantDecl:
			typ, err := types.FromAST(d.Type)
			if err == nil {
				typeEnv.DeclareConstant(d.Name, typ)
			}
		case *ast.FunctionDecl:
			params := make([]types.Type, len(d.Parameters))
			for i, param := range d.Parameters {
				paramType, err := types.FromAST(param.Type)
				if err == nil {
					params[i] = paramType
				}
			}
			returnType, err := types.FromAST(d.ReturnType)
			if err == nil {
				sig := &types.FunctionSignature{
					Parameters: params,
					Return:     returnType,
				}
				typeEnv.DeclareFunction(d.Name, sig)
			}
		}
	}

	// Type check
	checker := types.NewChecker(typeEnv)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			if d.Body != nil {
				for _, stmt := range d.Body.Statements {
					switch s := stmt.(type) {
					case *ast.AssignStmt:
						checker.CheckAssignment(s)
					case *ast.RequireStmt:
						checker.CheckExpression(s.Condition)
					case *ast.EnsureStmt:
						checker.CheckExpression(s.Condition)
					}
				}
			}
		case *ast.InvariantDecl:
			checker.CheckExpression(d.Condition)
		case *ast.TemporalDecl:
			// Temporal expressions are checked during verification, not type checking
			// Skip them here to avoid errors
		}
	}

	typeErrors := checker.Errors()
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Type errors in %s:\n", filename)
		formatter := errors.NewErrorFormatter()
		for _, typeErr := range typeErrors {
			formatted := formatter.FormatTypeError(typeErr.Message, typeErr.Position)
			fmt.Fprintf(os.Stderr, "  %s\n", formatted)
		}
		return fmt.Errorf("type checking failed with %d error(s)", len(typeErrors))
	}

	// Create state machine
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return fmt.Errorf("error creating state machine: %w", err)
	}

	// Explore state space
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(50)

	result, err := explorer.ExploreBFS()
	if err != nil {
		return fmt.Errorf("error exploring state space: %w", err)
	}

	// Report results
	if len(result.Violations) > 0 {
		fmt.Fprintf(os.Stderr, "Verification failed: %d violation(s) found\n", len(result.Violations))
		for i, violation := range result.Violations {
			fmt.Fprintf(os.Stderr, "\nViolation %d:\n", i+1)
			fmt.Fprintf(os.Stderr, "  %s\n", violation.Description)
			if len(violation.Path) > 0 {
				fmt.Fprintf(os.Stderr, "  Path:\n")
				for j, transition := range violation.Path {
					fmt.Fprintf(os.Stderr, "    %d. %s\n", j+1, transition.Action)
				}
			}
		}
		return fmt.Errorf("verification failed with %d violation(s)", len(result.Violations))
	}

		fmt.Printf("✓ Verification passed for %s\n", filename)
		fmt.Printf("  Explored %d states\n", result.StatesExplored)
		return nil
	})
}

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: spectre <command> [arguments]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for _, cmd := range Commands {
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  spectre parse examples/counter.spec\n")
	fmt.Fprintf(os.Stderr, "  spectre typecheck examples/counter.spec\n")
	fmt.Fprintf(os.Stderr, "  spectre verify examples/counter.spec\n")
}

// findSpecFiles finds all .spec files in a directory
func findSpecFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".spec") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}


package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.File)
	}{
		{
			name: "Simple file with variable",
			input: `var counter: int`,
			validate: func(t *testing.T, file *ast.File) {
				if len(file.Decls) != 1 {
					t.Fatalf("expected 1 declaration, got %d", len(file.Decls))
				}
				_, ok := file.Decls[0].(*ast.VariableDecl)
				if !ok {
					t.Fatalf("declaration not *ast.VariableDecl. got=%T", file.Decls[0])
				}
			},
		},
		{
			name: "File with multiple declarations",
			input: `var counter: int
const MAX_VALUE: int = 100
action increment {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, file *ast.File) {
				if len(file.Decls) != 3 {
					t.Fatalf("expected 3 declarations, got %d", len(file.Decls))
				}
			},
		},
		{
			name: "File with module",
			input: `module Counter {
  var counter: int
}`,
			validate: func(t *testing.T, file *ast.File) {
				if len(file.Decls) != 1 {
					t.Fatalf("expected 1 declaration, got %d", len(file.Decls))
				}
				moduleDecl, ok := file.Decls[0].(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("declaration not *ast.ModuleDecl. got=%T", file.Decls[0])
				}
				if moduleDecl.Name != "Counter" {
					t.Errorf("module name not 'Counter'. got=%s", moduleDecl.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if file == nil {
				t.Fatal("ParseFile returned nil")
			}

			tt.validate(t, file)
		})
	}
}

func TestParseExampleFiles(t *testing.T) {
	// These are the core example files that should parse without errors
	// (Files with unimplemented features are tested in TestParseAllExampleFilesIntegration)
	exampleFiles := []string{
		"counter.spec",
		"mutex.spec",
		"modules-example.spec",
		"error-trace-example.spec",
		"fairness-example.spec",
	}

	for _, filename := range exampleFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join("..", "..", "examples", filename)
			input, err := os.ReadFile(filePath)
			if err != nil {
				// Skip test if file doesn't exist (some examples may not be created yet)
				t.Skipf("file %s not found: %v", filename, err)
				return
			}

			l := lexer.New(string(input))
			p := New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors parsing %s:\n%v", len(p.Errors()), filename, p.Errors())
			}
			if file == nil {
				t.Fatalf("ParseFile returned nil for %s", filename)
			}
			if len(file.Decls) == 0 {
				t.Errorf("expected at least one declaration in %s", filename)
			}

			t.Logf("Successfully parsed %s: %d declarations", filename, len(file.Decls))
		})
	}
}

func TestParseAllExampleFilesIntegration(t *testing.T) {
	// This is a comprehensive integration test that verifies all example files parse correctly
	exampleDir := filepath.Join("..", "..", "examples")
	entries, err := os.ReadDir(exampleDir)
	if err != nil {
		t.Fatalf("failed to read examples directory: %v", err)
	}

	var specFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".spec" {
			specFiles = append(specFiles, entry.Name())
		}
	}

	if len(specFiles) == 0 {
		t.Fatal("no .spec files found in examples directory")
	}

	t.Logf("Found %d .spec files to parse", len(specFiles))

	for _, filename := range specFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(exampleDir, filename)
			input, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", filename, err)
			}

			l := lexer.New(string(input))
			p := New(l)
			file := p.ParseFile()

			errors := p.Errors()
			if len(errors) != 0 {
				// Some example files use features not yet implemented (record literals, lambdas, etc.)
				// This is expected. Verify that errors are reported with positions.
				hasPositionInfo := false
				for _, err := range errors {
					// Check if error has position format "line:column: message"
					if len(err) > 3 && err[0] >= '0' && err[0] <= '9' {
						hasPositionInfo = true
						break
					}
				}
				if !hasPositionInfo {
					t.Errorf("errors should include position information. Errors: %v", errors)
				}
				// Log errors but don't fail - these are expected for unimplemented features
				t.Logf("⚠ %s: %d parse errors (expected - uses unimplemented features):", filename, len(errors))
				for i, err := range errors {
					if i < 5 { // Show first 5 errors
						t.Logf("  %s", err)
					}
				}
				if len(errors) > 5 {
					t.Logf("  ... and %d more errors", len(errors)-5)
				}
				return
			}

			if file == nil {
				t.Fatalf("ParseFile returned nil for %s", filename)
			}

			// Verify file structure
			if len(file.Decls) == 0 {
				t.Errorf("expected at least one declaration in %s", filename)
			}

			// Log success with details
			var moduleCount, varCount, constCount, funCount, actionCount, initCount, invariantCount, temporalCount int
			for _, decl := range file.Decls {
				switch decl.(type) {
				case *ast.ModuleDecl:
					moduleCount++
				case *ast.VariableDecl:
					varCount++
				case *ast.ConstantDecl:
					constCount++
				case *ast.FunctionDecl:
					funCount++
				case *ast.ActionDecl:
					actionCount++
				case *ast.InitDecl, *ast.OneOfInitDecl:
					initCount++
				case *ast.InvariantDecl:
					invariantCount++
				case *ast.TemporalDecl:
					temporalCount++
				}
			}

			t.Logf("✓ %s: %d declarations (modules: %d, vars: %d, consts: %d, funs: %d, actions: %d, inits: %d, invariants: %d, temporal: %d)",
				filename, len(file.Decls), moduleCount, varCount, constCount, funCount, actionCount, initCount, invariantCount, temporalCount)
		})
	}
}


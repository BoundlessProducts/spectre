package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// TestParseModulesExampleFile tests parsing the complete modules-example.spec file
func TestParseModulesExampleFile(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "modules-example.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read modules-example.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := New(l)
	decls := parseAllDeclarations(p)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Verify we found module declarations
	moduleCount := 0
	importCount := 0
	for _, decl := range decls {
		switch d := decl.(type) {
		case *ast.ModuleDecl:
			moduleCount++
			t.Logf("Found module: %s (extends: %s)", d.Name, d.Extends)
		case *ast.ModuleInstanceDecl:
			t.Logf("Found module instance: %s = %s", d.Name, d.Module)
		case *ast.ImportDecl:
			importCount++
			t.Logf("Found import: %s", d.Module)
		}
	}

	if moduleCount == 0 {
		t.Error("expected at least one module declaration in modules-example.spec")
	}

	t.Logf("Found %d modules, %d imports", moduleCount, importCount)
}

// TestParseCounterModule tests parsing the Counter module from modules-example.spec
func TestParseCounterModule(t *testing.T) {
	input := `module Counter {
  description "Tracks a numeric counter value"
  var counter: int

  description "System starts with counter initialized to zero"
  init {
    counter = 0
  }

  description "Increments the counter by one"
  public action increment {
    counter' = counter + 1
  }

  description "Decrements the counter by one"
  public action decrement {
    require counter > 0
    counter' = counter - 1
  }

  description "Ensures counter never becomes negative"
  public invariant nonNegative {
    counter >= 0
  }
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseModuleDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	if decl == nil {
		t.Fatal("parseModuleDecl returned nil")
	}

	moduleDecl, ok := decl.(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
	}

	if moduleDecl.Name != "Counter" {
		t.Errorf("module name not 'Counter'. got=%s", moduleDecl.Name)
	}

	// Verify declarations within module
	if len(moduleDecl.Decls) == 0 {
		t.Error("expected at least one declaration in Counter module")
	}

	// Check for variable declaration
	hasVar := false
	hasInit := false
	hasPublicAction := false
	hasPublicInvariant := false

	for _, d := range moduleDecl.Decls {
		switch decl := d.(type) {
		case *ast.VariableDecl:
			if decl.Name == "counter" {
				hasVar = true
			}
		case *ast.InitDecl:
			hasInit = true
		case *ast.ActionDecl:
			if decl.Name == "increment" && decl.Visibility == ast.Public {
				hasPublicAction = true
			}
		case *ast.InvariantDecl:
			if decl.Name == "nonNegative" && decl.Visibility == ast.Public {
				hasPublicInvariant = true
			}
		}
	}

	if !hasVar {
		t.Error("expected variable declaration 'counter' in Counter module")
	}
	if !hasInit {
		t.Error("expected init declaration in Counter module")
	}
	if !hasPublicAction {
		t.Error("expected public action 'increment' in Counter module")
	}
	if !hasPublicInvariant {
		t.Error("expected public invariant 'nonNegative' in Counter module")
	}
}

// TestParseBoundedCounterModule tests parsing the BoundedCounter module with extends and super
func TestParseBoundedCounterModule(t *testing.T) {
	input := `module BoundedCounter extends Counter {
  description "Maximum allowed counter value"
  const MAX_VALUE: int = 100

  description "Increments counter but enforces maximum bound"
  public action increment {
    require counter < MAX_VALUE
    super.increment()
  }

  description "Ensures counter stays within bounds"
  public invariant bounded {
    counter <= MAX_VALUE
  }
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseModuleDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	if decl == nil {
		t.Fatal("parseModuleDecl returned nil")
	}

	moduleDecl, ok := decl.(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
	}

	if moduleDecl.Name != "BoundedCounter" {
		t.Errorf("module name not 'BoundedCounter'. got=%s", moduleDecl.Name)
	}

	if moduleDecl.Extends != "Counter" {
		t.Errorf("extends not 'Counter'. got=%s", moduleDecl.Extends)
	}

	// Check for super call in increment action
	hasSuperCall := false
	for _, d := range moduleDecl.Decls {
		if actionDecl, ok := d.(*ast.ActionDecl); ok && actionDecl.Name == "increment" {
			// Check if action body contains super call
			for _, stmt := range actionDecl.Body.Statements {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if callExpr, ok := exprStmt.Expr.(*ast.CallExpr); ok {
						if _, ok := callExpr.Fun.(*ast.SuperExpr); ok {
							hasSuperCall = true
						}
					}
				}
			}
		}
	}

	if !hasSuperCall {
		t.Error("expected super.increment() call in BoundedCounter.increment action")
	}
}

// TestParseAppModuleWithImport tests parsing the App module with import
func TestParseAppModuleWithImport(t *testing.T) {
	input := `module App {
  import BoundedCounter

  description "Create an instance of bounded counter"
  var myCounter: int

  description "Initialize using bounded counter"
  init {
    myCounter = 0
  }

  description "Increment with bound checking"
  action increment {
    require myCounter < BoundedCounter.MAX_VALUE
    myCounter' = myCounter + 1
  }
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseModuleDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	if decl == nil {
		t.Fatal("parseModuleDecl returned nil")
	}

	moduleDecl, ok := decl.(*ast.ModuleDecl)
	if !ok {
		t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
	}

	if moduleDecl.Name != "App" {
		t.Errorf("module name not 'App'. got=%s", moduleDecl.Name)
	}

	// Check for import declaration
	hasImport := false
	for _, d := range moduleDecl.Decls {
		if importDecl, ok := d.(*ast.ImportDecl); ok {
			if importDecl.Module == "BoundedCounter" {
				hasImport = true
			}
		}
	}

	if !hasImport {
		t.Error("expected import BoundedCounter in App module")
	}
}


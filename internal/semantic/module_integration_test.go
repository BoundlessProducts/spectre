package semantic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestModuleSystemOnExampleFile(t *testing.T) {
	// Test the module system with modules-example.spec
	examplePath := filepath.Join("..", "..", "examples", "modules-example.spec")
	
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("failed to read example file: %v", err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	// Log parse errors (may be expected for unimplemented features)
	if len(p.Errors()) > 0 {
		t.Logf("Parse errors in %s (may be expected): %v", examplePath, p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Logf("Build errors in %s: %v", examplePath, buildErrors)
	}

	// Resolve modules
	moduleResolver := NewModuleResolver(symbolTable)
	moduleErrors := moduleResolver.ResolveModules(file)
	if len(moduleErrors) > 0 {
		t.Logf("Module resolution errors in %s: %v", examplePath, moduleErrors)
		// Some errors may be expected (e.g., undefined identifiers due to unimplemented features)
	}

	// Check visibility
	visibilityChecker := NewVisibilityChecker(symbolTable, moduleResolver)
	visibilityErrors := visibilityChecker.CheckVisibility(file)
	if len(visibilityErrors) > 0 {
		t.Logf("Visibility errors in %s: %v", examplePath, visibilityErrors)
	}

	// Analyze inheritance
	inheritanceAnalyzer := NewInheritanceAnalyzer(symbolTable, moduleResolver)
	inheritanceErrors := inheritanceAnalyzer.AnalyzeInheritance(file)
	if len(inheritanceErrors) > 0 {
		t.Logf("Inheritance errors in %s: %v", examplePath, inheritanceErrors)
		// Some errors may be expected (e.g., undefined identifiers due to unimplemented features)
	}

	// Verify that modules were found
	modules := moduleResolver.GetModules()
	if len(modules) == 0 {
		t.Error("expected at least one module declaration in modules-example.spec")
	}

	// Verify Counter module exists
	counterModule, found := moduleResolver.GetModule("Counter")
	if !found {
		t.Error("expected Counter module not found")
	} else {
		if counterModule.Extends != "" {
			t.Errorf("expected Counter module to have no extension, got %s", counterModule.Extends)
		}
	}

	// Verify BoundedCounter module extends Counter
	boundedModule, found := moduleResolver.GetModule("BoundedCounter")
	if !found {
		t.Error("expected BoundedCounter module not found")
	} else {
		if boundedModule.Extends != "Counter" {
			t.Errorf("expected BoundedCounter to extend Counter, got %s", boundedModule.Extends)
		}
	}

	// Verify App module imports BoundedCounter
	// Note: imports are tracked by module name, but we need to check if App module has the import
	// For now, we just verify the import exists
	appModule, found := moduleResolver.GetModule("App")
	if !found {
		t.Error("expected App module not found")
	} else {
		// Check if App module has an import declaration
		hasImport := false
		for _, decl := range appModule.Decls {
			if importDecl, ok := decl.(*ast.ImportDecl); ok {
				if importDecl.Module == "BoundedCounter" {
					hasImport = true
					break
				}
			}
		}
		if !hasImport {
			t.Error("expected App module to import BoundedCounter")
		}
	}
}

func TestModuleSystemComplete(t *testing.T) {
	// Test complete module system analysis
	input := `
module Base {
    var x: int
    
    public action increment {
        x' = x + 1
    }
}

module Extended extends Base {
    var y: int
    
    public action increment {
        super.increment()
        y' = y + 1
    }
}

module App {
    import Extended
    
    var counter: int
    
    init {
        counter = 0
    }
    
    action increment {
        counter' = counter + 1
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Resolve modules
	moduleResolver := NewModuleResolver(symbolTable)
	moduleErrors := moduleResolver.ResolveModules(file)
	if len(moduleErrors) > 0 {
		t.Errorf("unexpected module resolution errors: %v", moduleErrors)
	}

	// Check visibility
	visibilityChecker := NewVisibilityChecker(symbolTable, moduleResolver)
	visibilityErrors := visibilityChecker.CheckVisibility(file)
	if len(visibilityErrors) > 0 {
		t.Errorf("unexpected visibility errors: %v", visibilityErrors)
	}

	// Analyze inheritance
	inheritanceAnalyzer := NewInheritanceAnalyzer(symbolTable, moduleResolver)
	inheritanceErrors := inheritanceAnalyzer.AnalyzeInheritance(file)
	if len(inheritanceErrors) > 0 {
		t.Errorf("unexpected inheritance errors: %v", inheritanceErrors)
	}

	// Verify module structure
	modules := moduleResolver.GetModules()
	if len(modules) != 3 {
		t.Errorf("expected 3 modules, got %d", len(modules))
	}

	// Verify extension chain
	extended, found := modules["Extended"]
	if !found {
		t.Error("Extended module not found")
	} else if extended.Extends != "Base" {
		t.Errorf("expected Extended to extend Base, got %s", extended.Extends)
	}
}


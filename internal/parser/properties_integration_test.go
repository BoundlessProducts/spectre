package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseAllDeclarations parses all declarations from the input until EOF
func parseAllDeclarations(p *Parser) []ast.Decl {
	var decls []ast.Decl
	
	for !p.curTokenIs(lexer.EOF) {
		// Skip comments and descriptions that aren't part of declarations
		if p.curTokenIs(lexer.DESCRIPTION) {
			p.parseDescription()
			continue
		}
		
		var decl ast.Decl
		switch p.curToken.Type {
		case lexer.VAR:
			decl = p.parseVariableDecl()
		case lexer.CONST:
			decl = p.parseConstantDecl()
		case lexer.FUN:
			decl = p.parseFunctionDecl()
		case lexer.ACTION:
			decl = p.parseActionDecl()
		case lexer.INIT:
			decl = p.parseInitDecl()
		case lexer.INVARIANT:
			decl = p.parseInvariantDecl()
		case lexer.TEMPORAL:
			decl = p.parseTemporalDecl()
		case lexer.MODULE:
			decl = p.parseModuleDecl()
		case lexer.IMPORT:
			decl = p.parseImportDecl()
		default:
			// Skip unknown tokens (comments, etc.)
			p.nextToken()
			continue
		}
		
		if decl != nil {
			decls = append(decls, decl)
		}
	}
	
	return decls
}

// TestParseInvariantsFromExampleFiles tests parsing invariants from example files
func TestParseInvariantsFromExampleFiles(t *testing.T) {
	exampleFiles := []string{
		"counter.spec",
		"mutex.spec",
		"fairness-example.spec",
	}

	for _, filename := range exampleFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join("..", "..", "examples", filename)
			input, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", filename, err)
			}

			l := lexer.New(string(input))
			p := New(l)
			decls := parseAllDeclarations(p)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			// Count invariants
			invariantCount := 0
			for _, decl := range decls {
				if _, ok := decl.(*ast.InvariantDecl); ok {
					invariantCount++
				}
			}

			// Verify we found at least one invariant
			if invariantCount == 0 {
				t.Errorf("expected at least one invariant in %s, found %d", filename, invariantCount)
			} else {
				t.Logf("Found %d invariants in %s", invariantCount, filename)
			}
		})
	}
}

// TestParseTemporalPropertiesFromExampleFiles tests parsing temporal properties from example files
func TestParseTemporalPropertiesFromExampleFiles(t *testing.T) {
	exampleFiles := []string{
		"counter.spec",
		"mutex.spec",
		"fairness-example.spec",
	}

	for _, filename := range exampleFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join("..", "..", "examples", filename)
			input, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", filename, err)
			}

			l := lexer.New(string(input))
			p := New(l)
			decls := parseAllDeclarations(p)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			// Count temporal properties
			temporalCount := 0
			for _, decl := range decls {
				if _, ok := decl.(*ast.TemporalDecl); ok {
					temporalCount++
				}
			}

			// Verify we found at least one temporal property
			if temporalCount == 0 {
				t.Errorf("expected at least one temporal property in %s, found %d", filename, temporalCount)
			} else {
				t.Logf("Found %d temporal properties in %s", temporalCount, filename)
			}
		})
	}
}

// TestParseCounterSpecProperties tests parsing all properties from counter.spec
func TestParseCounterSpecProperties(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "counter.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read counter.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := New(l)
	decls := parseAllDeclarations(p)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Verify invariants
	invariants := []string{"nonNegative", "bounded"}
	invariantMap := make(map[string]*ast.InvariantDecl)
	for _, decl := range decls {
		if inv, ok := decl.(*ast.InvariantDecl); ok {
			invariantMap[inv.Name] = inv
		}
	}

	for _, name := range invariants {
		if inv, ok := invariantMap[name]; !ok {
			t.Errorf("expected invariant %s not found", name)
		} else if inv.Condition == nil {
			t.Errorf("invariant %s has nil condition", name)
		}
	}

	// Verify temporal properties
	temporals := []string{"eventuallyReachesTen", "alwaysNonNegative", "progress"}
	temporalMap := make(map[string]*ast.TemporalDecl)
	for _, decl := range decls {
		if temp, ok := decl.(*ast.TemporalDecl); ok {
			temporalMap[temp.Name] = temp
		}
	}

	for _, name := range temporals {
		if temp, ok := temporalMap[name]; !ok {
			t.Errorf("expected temporal property %s not found", name)
		} else if temp.Expression == nil {
			t.Errorf("temporal property %s has nil expression", name)
		}
	}

	// Verify progress property has leads-to operator
	if progress, ok := temporalMap["progress"]; ok {
		alwaysExpr, ok := progress.Expression.(*ast.AlwaysExpr)
		if !ok {
			t.Fatalf("progress expression not *ast.AlwaysExpr. got=%T", progress.Expression)
		}
		leadsToExpr, ok := alwaysExpr.Expr.(*ast.LeadsToExpr)
		if !ok {
			t.Fatalf("progress inner expression not *ast.LeadsToExpr. got=%T", alwaysExpr.Expr)
		}
		if leadsToExpr.Left == nil || leadsToExpr.Right == nil {
			t.Error("progress leads-to expression has nil left or right")
		}
	}
}

// TestParseMutexSpecProperties tests parsing all properties from mutex.spec
func TestParseMutexSpecProperties(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "mutex.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read mutex.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := New(l)
	decls := parseAllDeclarations(p)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Verify invariants
	invariants := []string{"mutualExclusion", "lockConsistency"}
	invariantMap := make(map[string]*ast.InvariantDecl)
	for _, decl := range decls {
		if inv, ok := decl.(*ast.InvariantDecl); ok {
			invariantMap[inv.Name] = inv
		}
	}

	for _, name := range invariants {
		if inv, ok := invariantMap[name]; !ok {
			t.Errorf("expected invariant %s not found", name)
		} else if inv.Condition == nil {
			t.Errorf("invariant %s has nil condition", name)
		}
	}

	// Verify temporal properties with leads-to
	temporals := []string{"eventuallyProcess1Critical", "eventuallyProcess2Critical", "fairnessProcess1", "fairnessProcess2", "eventuallyRelease"}
	temporalMap := make(map[string]*ast.TemporalDecl)
	for _, decl := range decls {
		if temp, ok := decl.(*ast.TemporalDecl); ok {
			temporalMap[temp.Name] = temp
		}
	}

	for _, name := range temporals {
		if temp, ok := temporalMap[name]; !ok {
			t.Errorf("expected temporal property %s not found", name)
		} else if temp.Expression == nil {
			t.Errorf("temporal property %s has nil expression", name)
		}
	}
}

// TestParseFairnessExampleProperties tests parsing fairness properties from fairness-example.spec
func TestParseFairnessExampleProperties(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "fairness-example.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read fairness-example.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := New(l)
	decls := parseAllDeclarations(p)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Count WF and SF expressions
	wfCount := 0
	sfCount := 0
	for _, decl := range decls {
		if temp, ok := decl.(*ast.TemporalDecl); ok {
			if _, ok := temp.Expression.(*ast.WFExpr); ok {
				wfCount++
			}
			if _, ok := temp.Expression.(*ast.SFExpr); ok {
				sfCount++
			}
		}
	}

	if wfCount == 0 {
		t.Error("expected at least one WF expression in fairness-example.spec")
	}
	if sfCount == 0 {
		t.Error("expected at least one SF expression in fairness-example.spec")
	}

	t.Logf("Found %d WF expressions and %d SF expressions", wfCount, sfCount)
}


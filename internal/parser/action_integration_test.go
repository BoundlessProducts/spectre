package parser

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// TestParseActionsFromCounterSpec tests parsing actions from counter.spec examples
func TestParseActionsFromCounterSpec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "increment action",
			input: `description "Increments the counter by one"
action increment {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "increment" {
					t.Errorf("action name not 'increment'. got=%s", action.Name)
				}
				if action.Description == "" {
					t.Error("action should have a description")
				}
				if len(action.Body.Statements) == 0 {
					t.Error("action should have at least one statement")
				}
			},
		},
		{
			name: "decrement action with require",
			input: `description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "decrement" {
					t.Errorf("action name not 'decrement'. got=%s", action.Name)
				}
				if action.Description == "" {
					t.Error("action should have a description")
				}
				// decrement should have a require statement
				hasRequire := false
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.RequireStmt); ok {
						hasRequire = true
						break
					}
				}
				if !hasRequire {
					t.Error("decrement action should have a require statement")
				}
			},
		},
		{
			name: "reset action",
			input: `description "Resets the counter back to zero"
action reset {
  counter' = 0
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "reset" {
					t.Errorf("action name not 'reset'. got=%s", action.Name)
				}
				if action.Description == "" {
					t.Error("action should have a description")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}

// TestParseActionsFromMutexSpec tests parsing actions from mutex.spec examples
func TestParseActionsFromMutexSpec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "process1Request",
			input: `description "Process 1 requests and acquires the lock, entering critical section"
action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "process1Request" {
					t.Errorf("action name not 'process1Request'. got=%s", action.Name)
				}
				hasRequire := false
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.RequireStmt); ok {
						hasRequire = true
						break
					}
				}
				if !hasRequire {
					t.Error("process1Request should have a require statement")
				}
			},
		},
		{
			name: "process1Release",
			input: `description "Process 1 releases the lock and returns to idle state"
action process1Release {
  require process1 = ProcessState.Critical
  process1' = ProcessState.Idle
  lock' = false
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "process1Release" {
					t.Errorf("action name not 'process1Release'. got=%s", action.Name)
				}
				if len(action.Body.Statements) < 2 {
					t.Errorf("expected at least 2 statements, got %d", len(action.Body.Statements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}

// TestParseActionsFromOneOfExample tests parsing actions from oneof-example.spec examples
func TestParseActionsFromOneOfExample(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "setMode with parameters",
			input: `description "Changes the system mode to a new value"
action setMode(newMode: str) {
  require newMode = "start" || newMode = "resume" || newMode = "restart"
  counter' = counter
  mode' = newMode
  initialized' = initialized
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "setMode" {
					t.Errorf("action name not 'setMode'. got=%s", action.Name)
				}
				if len(action.Parameters) != 1 {
					t.Errorf("setMode should have 1 parameter, got %d", len(action.Parameters))
				}
				if action.Parameters[0].Name != "newMode" {
					t.Errorf("setMode parameter should be 'newMode', got %s", action.Parameters[0].Name)
				}
			},
		},
		{
			name: "increment with require",
			input: `description "Increments the counter by one, ensuring it stays within bounds"
action increment {
  require counter < 100
  counter' = counter + 1
  mode' = mode
  initialized' = true
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "increment" {
					t.Errorf("action name not 'increment'. got=%s", action.Name)
				}
				hasRequire := false
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.RequireStmt); ok {
						hasRequire = true
						break
					}
				}
				if !hasRequire {
					t.Error("increment should have a require statement")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}

// TestParseActionsFromUserManagement tests parsing actions from user-management.spec examples
func TestParseActionsFromUserManagement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "addUser with multiple parameters",
			input: `description "Adds a new user to the system"
action addUser(name: str, role: str) {
  users' = users
  nextId' = nextId + 1
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "addUser" {
					t.Errorf("action name not 'addUser'. got=%s", action.Name)
				}
				if len(action.Parameters) != 2 {
					t.Errorf("addUser should have 2 parameters, got %d", len(action.Parameters))
				}
				paramNames := make(map[string]bool)
				for _, param := range action.Parameters {
					paramNames[param.Name] = true
				}
				if !paramNames["name"] || !paramNames["role"] {
					t.Error("addUser should have 'name' and 'role' parameters")
				}
			},
		},
		{
			name: "removeUser with require",
			input: `description "Removes a user from the system"
action removeUser(id: int) {
  require id > 0
  users' = users
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "removeUser" {
					t.Errorf("action name not 'removeUser'. got=%s", action.Name)
				}
				hasRequire := false
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.RequireStmt); ok {
						hasRequire = true
						break
					}
				}
				if !hasRequire {
					t.Error("removeUser should have a require statement")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}

// TestParseActionsFromBankAccount tests parsing actions from bank-account.spec examples
func TestParseActionsFromBankAccount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "deposit with multiple require statements",
			input: `description "Deposits money into an account"
action deposit(accountId: int, amount: int) {
  require accountId > 0 && amount > 0
  require amount < 1000
  accounts' = accounts
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "deposit" {
					t.Errorf("action name not 'deposit'. got=%s", action.Name)
				}
				requireCount := 0
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.RequireStmt); ok {
						requireCount++
					}
				}
				if requireCount < 1 {
					t.Error("deposit should have at least one require statement")
				}
			},
		},
		{
			name: "transfer with multiple parameters",
			input: `description "Transfers money from one account to another"
action transfer(fromId: int, toId: int, amount: int) {
  require fromId > 0 && toId > 0 && amount > 0
  accounts' = accounts
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Name != "transfer" {
					t.Errorf("action name not 'transfer'. got=%s", action.Name)
				}
				if len(action.Parameters) != 3 {
					t.Errorf("transfer should have 3 parameters, got %d", len(action.Parameters))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}

// TestParseActionWithGuard tests parsing actions with guards from example files
func TestParseActionWithGuard(t *testing.T) {
	// Test action with guard syntax
	input := `description "Increments counter when below limit"
action increment when counter < 100 {
  counter' = counter + 1
}`

	l := lexer.New(input)
	p := New(l)

	decl := p.parseActionDecl()
	if decl == nil {
		t.Fatal("parseActionDecl returned nil")
	}

	actionDecl, ok := decl.(*ast.ActionDecl)
	if !ok {
		t.Fatalf("not *ast.ActionDecl. got=%T", decl)
	}

	if actionDecl.Guard == nil {
		t.Error("action should have a guard")
	}

	if actionDecl.Description == "" {
		t.Error("action should have a description")
	}

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
}

// TestParseActionWithMultipleStatements tests parsing actions with multiple statements
func TestParseActionWithMultipleStatements(t *testing.T) {
	input := `action complexAction {
  require counter > 0
  counter' = counter + 1
  mode' = "updated"
  ensure counter' > counter
}`

	l := lexer.New(input)
	p := New(l)

	decl := p.parseActionDecl()
	if decl == nil {
		t.Fatal("parseActionDecl returned nil")
	}

	actionDecl, ok := decl.(*ast.ActionDecl)
	if !ok {
		t.Fatalf("not *ast.ActionDecl. got=%T", decl)
	}

	if len(actionDecl.Body.Statements) < 4 {
		t.Errorf("expected at least 4 statements, got %d", len(actionDecl.Body.Statements))
	}

	// Check for require statement
	hasRequire := false
	hasEnsure := false
	for _, stmt := range actionDecl.Body.Statements {
		if _, ok := stmt.(*ast.RequireStmt); ok {
			hasRequire = true
		}
		if _, ok := stmt.(*ast.EnsureStmt); ok {
			hasEnsure = true
		}
	}

	if !hasRequire {
		t.Error("action should have a require statement")
	}
	if !hasEnsure {
		t.Error("action should have an ensure statement")
	}

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
}

// TestParseActionWithPrimedAssignments tests parsing actions with primed assignments
func TestParseActionWithPrimedAssignments(t *testing.T) {
	input := `action update {
  counter' = counter + 1
  status' = "active"
  flag' = true
}`

	l := lexer.New(input)
	p := New(l)

	decl := p.parseActionDecl()
	if decl == nil {
		t.Fatal("parseActionDecl returned nil")
	}

	actionDecl, ok := decl.(*ast.ActionDecl)
	if !ok {
		t.Fatalf("not *ast.ActionDecl. got=%T", decl)
	}

	// Count primed assignments
	primedCount := 0
	for _, stmt := range actionDecl.Body.Statements {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if ident, ok := assignStmt.Left.(*ast.Ident); ok && ident.Prime {
				primedCount++
			}
		}
	}

	if primedCount < 3 {
		t.Errorf("expected at least 3 primed assignments, got %d", primedCount)
	}

	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
}

// TestParseAllActionConstructs tests parsing various action constructs
func TestParseAllActionConstructs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, *ast.ActionDecl)
	}{
		{
			name: "action with guard",
			input: `action increment when counter < 100 {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				if action.Guard == nil {
					t.Error("action should have a guard")
				}
			},
		},
		{
			name: "action with ensure",
			input: `action increment {
  counter' = counter + 1
  ensure counter' > counter
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				hasEnsure := false
				for _, stmt := range action.Body.Statements {
					if _, ok := stmt.(*ast.EnsureStmt); ok {
						hasEnsure = true
						break
					}
				}
				if !hasEnsure {
					t.Error("action should have an ensure statement")
				}
			},
		},
		{
			name: "action with multiple primed assignments",
			input: `action update {
  counter' = counter + 1
  status' = "active"
  flag' = true
}`,
			validate: func(t *testing.T, action *ast.ActionDecl) {
				primedCount := 0
				for _, stmt := range action.Body.Statements {
					if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
						if ident, ok := assignStmt.Left.(*ast.Ident); ok && ident.Prime {
							primedCount++
						}
					}
				}
				if primedCount < 3 {
					t.Errorf("expected at least 3 primed assignments, got %d", primedCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseActionDecl returned nil")
			}

			actionDecl, ok := decl.(*ast.ActionDecl)
			if !ok {
				t.Fatalf("not *ast.ActionDecl. got=%T", decl)
			}

			tt.validate(t, actionDecl)
		})
	}
}


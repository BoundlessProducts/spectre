package state

import (
	"fmt"

	"github.com/akkeshavan/spectre/pkg/ast"
)

// InitialStateModel represents the initial state configuration
// It handles both deterministic (single init) and non-deterministic (oneOf) initial states
type InitialStateModel struct {
	Deterministic *InitStateConfig // Single deterministic initial state
	OneOf         []*InitStateConfig // Multiple possible initial states (oneOf)
}

// InitStateConfig represents a single initial state configuration
type InitStateConfig struct {
	InitDecl   *ast.InitDecl       // The init declaration (for deterministic)
	OneOfOption *ast.BlockStmt     // The oneOf option block (for oneOf)
	Position   ast.Position        // Position in source
	Description string             // Description from declaration
}

// NewInitialStateModel creates a new initial state model from an AST file
func NewInitialStateModel(file *ast.File) (*InitialStateModel, error) {
	model := &InitialStateModel{
		OneOf: []*InitStateConfig{},
	}

	// Find all init declarations
	var initDecls []*ast.InitDecl
	var oneOfDecls []*ast.OneOfInitDecl

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.InitDecl:
			initDecls = append(initDecls, d)
		case *ast.OneOfInitDecl:
			oneOfDecls = append(oneOfDecls, d)
		case *ast.ModuleDecl:
			// Check for init declarations within modules
			for _, innerDecl := range d.Decls {
				switch innerD := innerDecl.(type) {
				case *ast.InitDecl:
					initDecls = append(initDecls, innerD)
				case *ast.OneOfInitDecl:
					oneOfDecls = append(oneOfDecls, innerD)
				}
			}
		}
	}

	// Process deterministic init declarations
	if len(initDecls) > 0 {
		if len(initDecls) > 1 {
			return nil, fmt.Errorf("multiple init declarations found (use oneOf for multiple initial states)")
		}
		model.Deterministic = &InitStateConfig{
			InitDecl:    initDecls[0],
			Position:    initDecls[0].Position,
			Description: initDecls[0].Description,
		}
	}

	// Process oneOf declarations
	if len(oneOfDecls) > 0 {
		if len(oneOfDecls) > 1 {
			return nil, fmt.Errorf("multiple oneOf declarations found")
		}
		oneOfDecl := oneOfDecls[0]
		for _, option := range oneOfDecl.Options {
			model.OneOf = append(model.OneOf, &InitStateConfig{
				OneOfOption: option,
				Position:    oneOfDecl.Position,
				Description: oneOfDecl.Description,
			})
		}
	}

	// Validate that we have either deterministic or oneOf, but not both
	if model.Deterministic != nil && len(model.OneOf) > 0 {
		return nil, fmt.Errorf("cannot have both deterministic init and oneOf init")
	}

	if model.Deterministic == nil && len(model.OneOf) == 0 {
		return nil, fmt.Errorf("no initial state declaration found")
	}

	return model, nil
}

// IsDeterministic returns true if the initial state is deterministic
func (ism *InitialStateModel) IsDeterministic() bool {
	return ism.Deterministic != nil
}

// IsOneOf returns true if the initial state is non-deterministic (oneOf)
func (ism *InitialStateModel) IsOneOf() bool {
	return len(ism.OneOf) > 0
}

// GetInitialStates returns all possible initial states
// For deterministic: returns a single state config
// For oneOf: returns all option configs
func (ism *InitialStateModel) GetInitialStates() []*InitStateConfig {
	if ism.IsDeterministic() {
		return []*InitStateConfig{ism.Deterministic}
	}
	return ism.OneOf
}

// Count returns the number of possible initial states
func (ism *InitialStateModel) Count() int {
	if ism.IsDeterministic() {
		return 1
	}
	return len(ism.OneOf)
}

// GetDeterministicInit returns the deterministic init declaration
func (ism *InitialStateModel) GetDeterministicInit() *ast.InitDecl {
	if ism.Deterministic == nil {
		return nil
	}
	return ism.Deterministic.InitDecl
}

// GetOneOfOptions returns all oneOf option blocks
func (ism *InitialStateModel) GetOneOfOptions() []*ast.BlockStmt {
	if !ism.IsOneOf() {
		return nil
	}
	options := make([]*ast.BlockStmt, len(ism.OneOf))
	for i, config := range ism.OneOf {
		options[i] = config.OneOfOption
	}
	return options
}


package exec

import (
	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// ActionWithArgs represents an action with its argument values
type ActionWithArgs struct {
	ActionName string
	Args       []state.Value
}

// generateArgumentCombinations generates argument value combinations for a parameterized action
// This is a heuristic approach - for ints, it tries a reasonable range; for bools, true/false; for enums, all values
func generateArgumentCombinations(actionInfo *state.ActionInfo, env *eval.Environment, file *ast.File) [][]state.Value {
	if len(actionInfo.Parameters) == 0 {
		return [][]state.Value{{}} // Single combination with no arguments
	}

	// Generate possible values for each parameter
	paramValues := make([][]state.Value, len(actionInfo.Parameters))
	
	for i, param := range actionInfo.Parameters {
		values := generateValuesForParameter(param, env, file)
		paramValues[i] = values
	}

	// Generate all combinations using cartesian product
	return cartesianProduct(paramValues)
}

// generateValuesForParameter generates possible values for a single parameter
func generateValuesForParameter(param ast.Parameter, env *eval.Environment, file *ast.File) []state.Value {
	// Try to resolve the parameter type
	// For now, use a simple heuristic based on the type name
	switch t := param.Type.(type) {
	case *ast.PrimitiveType:
		switch t.Name {
		case "int":
			// Try to get NUM_FLOORS or use default range 0-20
			maxValue := int64(20) // Default
			if numFloors, exists := env.GetConstant("NUM_FLOORS"); exists {
				if pv, ok := numFloors.(*state.PrimitiveValue); ok && pv.IntValue != nil {
					maxValue = *pv.IntValue
				}
			}
			// Generate values 0 to maxValue-1
			values := make([]state.Value, maxValue)
			for i := int64(0); i < maxValue; i++ {
				values[i] = state.NewIntValue(i)
			}
			return values
		case "bool":
			return []state.Value{
				state.NewBoolValue(true),
				state.NewBoolValue(false),
			}
		case "str":
			// For strings, we can't generate all possible values
			// Return empty - this parameter type won't be explored
			return []state.Value{}
		case "float":
			// For floats, we can't generate all possible values
			// Return empty - this parameter type won't be explored
			return []state.Value{}
		}
	case *ast.EnumType:
		// Generate all enum values
		values := make([]state.Value, len(t.Values))
		for i, valName := range t.Values {
			values[i] = state.NewEnumValue(t.Name, valName)
		}
		return values
	case *ast.NamedType:
		// Try to resolve the named type
		// For now, if it's a known enum, generate its values
		// Otherwise, return empty
		enumDef, exists := env.GetEnumType(t.Name)
		if exists {
			values := make([]state.Value, len(enumDef.Values))
			for i, valName := range enumDef.Values {
				values[i] = state.NewEnumValue(t.Name, valName)
			}
			return values
		}
		// For other named types (like records, sets, lists), we can't generate values
		return []state.Value{}
	}

	// Unknown type - return empty
	return []state.Value{}
}

// cartesianProduct generates all combinations from multiple slices
func cartesianProduct(slices [][]state.Value) [][]state.Value {
	if len(slices) == 0 {
		return [][]state.Value{{}}
	}

	if len(slices) == 1 {
		result := make([][]state.Value, len(slices[0]))
		for i, val := range slices[0] {
			result[i] = []state.Value{val}
		}
		return result
	}

	// Recursively compute cartesian product
	rest := cartesianProduct(slices[1:])
	result := [][]state.Value{}
	
	for _, val := range slices[0] {
		for _, combo := range rest {
			newCombo := make([]state.Value, len(combo)+1)
			newCombo[0] = val
			copy(newCombo[1:], combo)
			result = append(result, newCombo)
		}
	}

	return result
}


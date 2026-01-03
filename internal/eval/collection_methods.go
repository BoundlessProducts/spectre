package eval

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// evalFilter evaluates the filter method: collection.filter(predicate)
// predicate is a lambda: element => bool
func (e *Evaluator) evalFilter(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("filter expects 1 argument (predicate), got %d", len(args))
	}

	// Get the lambda predicate
	lambda, ok := args[0].(*LambdaValue)
	if !ok {
		return nil, fmt.Errorf("filter predicate must be a lambda expression")
	}

	if len(lambda.Params) != 1 {
		return nil, fmt.Errorf("filter predicate lambda must have exactly 1 parameter")
	}

	// Handle Set
	if setVal, ok := obj.(*state.SetValue); ok {
		result := state.NewSetValue()
		for _, elem := range setVal.Values {
			// Call lambda with element
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling filter predicate: %w", err)
			}
			
			// Check if predicate returned true
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil && *boolVal.BoolValue {
				result.Add(elem)
			}
		}
		return result, nil
	}

	// Handle List
	if listVal, ok := obj.(*state.ListValue); ok {
		result := state.NewListValue()
		for _, elem := range listVal.Elements {
			// Call lambda with element
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling filter predicate: %w", err)
			}
			
			// Check if predicate returned true
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil && *boolVal.BoolValue {
				result.Append(elem)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("filter can only be called on Set or List, got %s", obj.Type())
}

// evalMap evaluates the map method: collection.map(fn)
// fn is a lambda: element => newElement
func (e *Evaluator) evalMap(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("map expects 1 argument (function), got %d", len(args))
	}

	// Get the lambda function
	lambda, ok := args[0].(*LambdaValue)
	if !ok {
		return nil, fmt.Errorf("map function must be a lambda expression")
	}

	if len(lambda.Params) != 1 {
		return nil, fmt.Errorf("map function lambda must have exactly 1 parameter")
	}

	// Handle Set
	if setVal, ok := obj.(*state.SetValue); ok {
		result := state.NewSetValue()
		for _, elem := range setVal.Values {
			// Call lambda with element
			mappedElem, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling map function: %w", err)
			}
			result.Add(mappedElem)
		}
		return result, nil
	}

	// Handle List
	if listVal, ok := obj.(*state.ListValue); ok {
		result := state.NewListValue()
		for _, elem := range listVal.Elements {
			// Call lambda with element
			mappedElem, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling map function: %w", err)
			}
			result.Append(mappedElem)
		}
		return result, nil
	}

	return nil, fmt.Errorf("map can only be called on Set or List, got %s", obj.Type())
}

// evalReduce evaluates the reduce method: collection.reduce(initial, fn)
// fn is a lambda: (accumulator, element) => accumulator
func (e *Evaluator) evalReduce(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("reduce expects 2 arguments (initial, function), got %d", len(args))
	}

	accumulator := args[0] // initial value

	// Get the lambda function
	lambda, ok := args[1].(*LambdaValue)
	if !ok {
		return nil, fmt.Errorf("reduce function must be a lambda expression")
	}

	if len(lambda.Params) != 2 {
		return nil, fmt.Errorf("reduce function lambda must have exactly 2 parameters")
	}

	// Handle Set
	if setVal, ok := obj.(*state.SetValue); ok {
		for _, elem := range setVal.Values {
			// Call lambda with accumulator and element
			newAccumulator, err := lambda.Call([]state.Value{accumulator, elem})
			if err != nil {
				return nil, fmt.Errorf("error calling reduce function: %w", err)
			}
			accumulator = newAccumulator
		}
		return accumulator, nil
	}

	// Handle List
	if listVal, ok := obj.(*state.ListValue); ok {
		for _, elem := range listVal.Elements {
			// Call lambda with accumulator and element
			newAccumulator, err := lambda.Call([]state.Value{accumulator, elem})
			if err != nil {
				return nil, fmt.Errorf("error calling reduce function: %w", err)
			}
			accumulator = newAccumulator
		}
		return accumulator, nil
	}

	return nil, fmt.Errorf("reduce can only be called on Set or List, got %s", obj.Type())
}

// evalForall evaluates the forall method: collection.forall(predicate)
// predicate is a lambda: element => bool
func (e *Evaluator) evalForall(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("forall expects 1 argument (predicate), got %d", len(args))
	}

	// Get the lambda predicate
	lambda, ok := args[0].(*LambdaValue)
	if !ok {
		return nil, fmt.Errorf("forall predicate must be a lambda expression")
	}

	if len(lambda.Params) != 1 {
		return nil, fmt.Errorf("forall predicate lambda must have exactly 1 parameter")
	}

	// Handle Set
	if setVal, ok := obj.(*state.SetValue); ok {
		for _, elem := range setVal.Values {
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling forall predicate: %w", err)
			}
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
				if !*boolVal.BoolValue {
					return state.NewBoolValue(false), nil
				}
			} else {
				return nil, fmt.Errorf("forall predicate must return bool")
			}
		}
		return state.NewBoolValue(true), nil
	}

	// Handle List
	if listVal, ok := obj.(*state.ListValue); ok {
		for _, elem := range listVal.Elements {
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling forall predicate: %w", err)
			}
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
				if !*boolVal.BoolValue {
					return state.NewBoolValue(false), nil
				}
			} else {
				return nil, fmt.Errorf("forall predicate must return bool")
			}
		}
		return state.NewBoolValue(true), nil
	}

	return nil, fmt.Errorf("forall can only be called on Set or List, got %s", obj.Type())
}

// evalExists evaluates the exists method: collection.exists(predicate)
// predicate is a lambda: element => bool
func (e *Evaluator) evalExists(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists expects 1 argument (predicate), got %d", len(args))
	}

	// Get the lambda predicate
	lambda, ok := args[0].(*LambdaValue)
	if !ok {
		return nil, fmt.Errorf("exists predicate must be a lambda expression")
	}

	if len(lambda.Params) != 1 {
		return nil, fmt.Errorf("exists predicate lambda must have exactly 1 parameter")
	}

	// Handle Set
	if setVal, ok := obj.(*state.SetValue); ok {
		for _, elem := range setVal.Values {
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling exists predicate: %w", err)
			}
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
				if *boolVal.BoolValue {
					return state.NewBoolValue(true), nil
				}
			} else {
				return nil, fmt.Errorf("exists predicate must return bool")
			}
		}
		return state.NewBoolValue(false), nil
	}

	// Handle List
	if listVal, ok := obj.(*state.ListValue); ok {
		for _, elem := range listVal.Elements {
			predResult, err := lambda.Call([]state.Value{elem})
			if err != nil {
				return nil, fmt.Errorf("error calling exists predicate: %w", err)
			}
			if boolVal, ok := predResult.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
				if *boolVal.BoolValue {
					return state.NewBoolValue(true), nil
				}
			} else {
				return nil, fmt.Errorf("exists predicate must return bool")
			}
		}
		return state.NewBoolValue(false), nil
	}

	return nil, fmt.Errorf("exists can only be called on Set or List, got %s", obj.Type())
}

// evalSize evaluates the size method: collection.size()
func (e *Evaluator) evalSize(obj state.Value) (state.Value, error) {
	if setVal, ok := obj.(*state.SetValue); ok {
		return state.NewIntValue(setVal.Size()), nil
	}
	if listVal, ok := obj.(*state.ListValue); ok {
		return state.NewIntValue(listVal.Size()), nil
	}
	if mapVal, ok := obj.(*state.MapValue); ok {
		return state.NewIntValue(mapVal.Size()), nil
	}
	return nil, fmt.Errorf("size can only be called on Set, List, or Map, got %s", obj.Type())
}

// evalContains evaluates the contains method: collection.contains(element)
func (e *Evaluator) evalContains(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("contains expects 1 argument (element), got %d", len(args))
	}
	elem := args[0]
	
	if setVal, ok := obj.(*state.SetValue); ok {
		return state.NewBoolValue(setVal.Contains(elem)), nil
	}
	if mapVal, ok := obj.(*state.MapValue); ok {
		return state.NewBoolValue(mapVal.Contains(elem)), nil
	}
	return nil, fmt.Errorf("contains can only be called on Set or Map, got %s", obj.Type())
}

// evalUnion evaluates the union method: set.union(otherSet)
func (e *Evaluator) evalUnion(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("union expects 1 argument (other set), got %d", len(args))
	}
	
	setVal, ok := obj.(*state.SetValue)
	if !ok {
		return nil, fmt.Errorf("union can only be called on Set, got %s", obj.Type())
	}
	
	otherSet, ok := args[0].(*state.SetValue)
	if !ok {
		return nil, fmt.Errorf("union argument must be a Set, got %s", args[0].Type())
	}
	
	result := state.NewSetValue()
	// Add all elements from first set
	for _, elem := range setVal.Values {
		result.Add(elem)
	}
	// Add all elements from second set
	for _, elem := range otherSet.Values {
		result.Add(elem)
	}
	return result, nil
}

// evalIntersection evaluates the intersection method: set.intersection(otherSet)
func (e *Evaluator) evalIntersection(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("intersection expects 1 argument (other set), got %d", len(args))
	}
	
	setVal, ok := obj.(*state.SetValue)
	if !ok {
		return nil, fmt.Errorf("intersection can only be called on Set, got %s", obj.Type())
	}
	
	otherSet, ok := args[0].(*state.SetValue)
	if !ok {
		return nil, fmt.Errorf("intersection argument must be a Set, got %s", args[0].Type())
	}
	
	result := state.NewSetValue()
	// Add elements that exist in both sets
	for _, elem := range setVal.Values {
		if otherSet.Contains(elem) {
			result.Add(elem)
		}
	}
	return result, nil
}

// evalHead evaluates the head method: list.head()
func (e *Evaluator) evalHead(obj state.Value) (state.Value, error) {
	if listVal, ok := obj.(*state.ListValue); ok {
		return listVal.Head()
	}
	return nil, fmt.Errorf("head can only be called on List, got %s", obj.Type())
}

// evalTail evaluates the tail method: list.tail()
func (e *Evaluator) evalTail(obj state.Value) (state.Value, error) {
	if listVal, ok := obj.(*state.ListValue); ok {
		return listVal.Tail(), nil
	}
	return nil, fmt.Errorf("tail can only be called on List, got %s", obj.Type())
}

// evalToList evaluates the toList method: collection.toList()
func (e *Evaluator) evalToList(obj state.Value) (state.Value, error) {
	if setVal, ok := obj.(*state.SetValue); ok {
		result := state.NewListValue()
		for _, elem := range setVal.Values {
			result.Append(elem)
		}
		return result, nil
	}
	if listVal, ok := obj.(*state.ListValue); ok {
		// Already a list, return as-is
		return listVal, nil
	}
	return nil, fmt.Errorf("toList can only be called on Set or List, got %s", obj.Type())
}

// evalToSet evaluates the toSet method: collection.toSet()
func (e *Evaluator) evalToSet(obj state.Value) (state.Value, error) {
	if listVal, ok := obj.(*state.ListValue); ok {
		result := state.NewSetValue()
		for _, elem := range listVal.Elements {
			result.Add(elem)
		}
		return result, nil
	}
	if setVal, ok := obj.(*state.SetValue); ok {
		// Already a set, return as-is
		return setVal, nil
	}
	return nil, fmt.Errorf("toSet can only be called on List or Set, got %s", obj.Type())
}

// evalAppend evaluates the append method: list.append(element)
func (e *Evaluator) evalAppend(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("append expects 1 argument (element), got %d", len(args))
	}
	
	if listVal, ok := obj.(*state.ListValue); ok {
		result := state.NewListValue()
		// Copy existing elements
		for _, elem := range listVal.Elements {
			result.Append(elem)
		}
		// Append new element
		result.Append(args[0])
		return result, nil
	}
	
	return nil, fmt.Errorf("append can only be called on List, got %s", obj.Type())
}

// evalIndexExpr evaluates an index expression: container[index] or map[key]
// This is called from evaluator.go
func (e *Evaluator) evalIndexExpr(expr *ast.IndexExpr) (state.Value, error) {
	// Evaluate the container (map, list, etc.)
	container, err := e.Eval(expr.X)
	if err != nil {
		return nil, fmt.Errorf("error evaluating container in index expression: %w", err)
	}

	// Evaluate the index/key
	index, err := e.Eval(expr.Index)
	if err != nil {
		return nil, fmt.Errorf("error evaluating index in index expression: %w", err)
	}

	// Handle map indexing: map[key]
	if mapVal, ok := container.(*state.MapValue); ok {
		value, exists := mapVal.Get(index)
		if !exists {
			return nil, fmt.Errorf("key not found in map")
		}
		return value, nil
	}

	// Handle list indexing: list[index]
	if listVal, ok := container.(*state.ListValue); ok {
		// Index must be an integer
		indexPrim, ok := index.(*state.PrimitiveValue)
		if !ok || indexPrim.TypeName != "int" {
			return nil, fmt.Errorf("list index must be int, got %s", index.Type())
		}
		if indexPrim.IntValue == nil {
			return nil, fmt.Errorf("invalid list index")
		}
		idx := int(*indexPrim.IntValue)
		if idx < 0 || idx >= len(listVal.Elements) {
			return nil, fmt.Errorf("list index out of bounds: %d (length: %d)", idx, len(listVal.Elements))
		}
		return listVal.Elements[idx], nil
	}

	return nil, fmt.Errorf("cannot index type %s", container.Type())
}

// evalSetLiteral evaluates a set literal: { value1, value2, ... }
func (e *Evaluator) evalSetLiteral(expr *ast.SetLiteral) (state.Value, error) {
	result := state.NewSetValue()
	
	for _, elemExpr := range expr.Elements {
		elem, err := e.Eval(elemExpr)
		if err != nil {
			return nil, fmt.Errorf("error evaluating set element: %w", err)
		}
		result.Add(elem)
	}
	
	return result, nil
}

// evalListLiteral evaluates a list literal: [ value1, value2, ... ]
func (e *Evaluator) evalListLiteral(expr *ast.ListLiteral) (state.Value, error) {
	result := state.NewListValue()
	
	for _, elemExpr := range expr.Elements {
		elem, err := e.Eval(elemExpr)
		if err != nil {
			return nil, fmt.Errorf("error evaluating list element: %w", err)
		}
		result.Append(elem)
	}
	
	return result, nil
}

// evalPut evaluates the put method: map.put(key, value)
// Returns a new map with the key-value pair added/updated
func (e *Evaluator) evalPut(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("put expects 2 arguments (key, value), got %d", len(args))
	}
	
	mapVal, ok := obj.(*state.MapValue)
	if !ok {
		return nil, fmt.Errorf("put can only be called on Map, got %s", obj.Type())
	}
	
	key := args[0]
	value := args[1]
	
	// Create a new map with all existing entries plus the new/updated entry
	result := state.NewMapValue()
	// Copy all existing entries
	for k, v := range mapVal.Entries {
		result.Entries[k] = v
	}
	// Add or update the new entry
	result.Put(key, value)
	return result, nil
}

// evalGet evaluates the get method: map.get(key)
// Returns the value for the key, or nil if key doesn't exist
func (e *Evaluator) evalGet(obj state.Value, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("get expects 1 argument (key), got %d", len(args))
	}
	
	mapVal, ok := obj.(*state.MapValue)
	if !ok {
		return nil, fmt.Errorf("get can only be called on Map, got %s", obj.Type())
	}
	
	key := args[0]
	value, exists := mapVal.Get(key)
	if !exists {
		// Return a zero value or handle missing key
		// For now, return nil which will cause an error
		// In the future, this could return Option<Value>
		return nil, fmt.Errorf("key not found in map")
	}
	return value, nil
}


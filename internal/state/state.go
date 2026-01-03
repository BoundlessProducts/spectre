package state

import (
	"fmt"

	"github.com/akkeshavan/spectre/pkg/ast"
)

// State represents a state in the state machine
// It holds the values of all state variables
type State struct {
	Variables map[string]Value // Variable name -> value
	Position  ast.Position     // Position in source (for error reporting)
}

// Value represents a value in the state machine
// It can be a primitive value or a compound value
type Value interface {
	Type() string
	String() string
}

// PrimitiveValue represents a primitive value
type PrimitiveValue struct {
	TypeName string
	IntValue    *int64
	FloatValue  *float64
	StringValue *string
	BoolValue   *bool
}

func (v *PrimitiveValue) Type() string {
	return v.TypeName
}

func (v *PrimitiveValue) String() string {
	switch v.TypeName {
	case "int":
		if v.IntValue != nil {
			return fmt.Sprintf("%d", *v.IntValue)
		}
		return "0"
	case "float":
		if v.FloatValue != nil {
			return fmt.Sprintf("%f", *v.FloatValue)
		}
		return "0.0"
	case "str":
		if v.StringValue != nil {
			return fmt.Sprintf("%q", *v.StringValue)
		}
		return `""`
	case "bool":
		if v.BoolValue != nil {
			return fmt.Sprintf("%t", *v.BoolValue)
		}
		return "false"
	default:
		return "unknown"
	}
}

// EnumValue represents an enum value
type EnumValue struct {
	EnumName  string // Name of the enum type (e.g., "ProcessState")
	ValueName string // Name of the enum value (e.g., "Idle")
}

func (v *EnumValue) Type() string {
	return fmt.Sprintf("enum(%s)", v.EnumName)
}

func (v *EnumValue) String() string {
	return fmt.Sprintf("%s.%s", v.EnumName, v.ValueName)
}

// NewEnumValue creates a new enum value
func NewEnumValue(enumName, valueName string) *EnumValue {
	return &EnumValue{
		EnumName:  enumName,
		ValueName: valueName,
	}
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Variables: make(map[string]Value),
		Position:  ast.Position{Line: 0, Column: 0, Offset: 0},
	}
}

// Copy creates a deep copy of the state
func (s *State) Copy() *State {
	copy := NewState()
	copy.Position = s.Position
	for name, value := range s.Variables {
		copy.Variables[name] = value // For now, shallow copy (will need deep copy for compound types)
	}
	return copy
}

// SetVariable sets a variable value in the state
func (s *State) SetVariable(name string, value Value) {
	s.Variables[name] = value
}

// GetVariable gets a variable value from the state
func (s *State) GetVariable(name string) (Value, bool) {
	value, exists := s.Variables[name]
	return value, exists
}

// Equals checks if two states are equal
func (s *State) Equals(other *State) bool {
	if len(s.Variables) != len(other.Variables) {
		return false
	}

	for name, value := range s.Variables {
		otherValue, exists := other.Variables[name]
		if !exists {
			return false
		}
		if !valuesEqual(value, otherValue) {
			return false
		}
	}

	return true
}

// valuesEqual checks if two values are equal
func valuesEqual(v1, v2 Value) bool {
	if v1.Type() != v2.Type() {
		return false
	}

	// For primitive values, compare directly
	if p1, ok := v1.(*PrimitiveValue); ok {
		if p2, ok := v2.(*PrimitiveValue); ok {
			return primitiveValuesEqual(p1, p2)
		}
	}

	// For enum values, compare enum name and value name
	if e1, ok := v1.(*EnumValue); ok {
		if e2, ok := v2.(*EnumValue); ok {
			return e1.EnumName == e2.EnumName && e1.ValueName == e2.ValueName
		}
	}

	// For compound types, will need more sophisticated comparison
	// For now, just compare string representation
	return v1.String() == v2.String()
}

// primitiveValuesEqual checks if two primitive values are equal
func primitiveValuesEqual(v1, v2 *PrimitiveValue) bool {
	if v1.TypeName != v2.TypeName {
		return false
	}

	switch v1.TypeName {
	case "int":
		val1 := int64(0)
		val2 := int64(0)
		if v1.IntValue != nil {
			val1 = *v1.IntValue
		}
		if v2.IntValue != nil {
			val2 = *v2.IntValue
		}
		return val1 == val2
	case "float":
		val1 := float64(0)
		val2 := float64(0)
		if v1.FloatValue != nil {
			val1 = *v1.FloatValue
		}
		if v2.FloatValue != nil {
			val2 = *v2.FloatValue
		}
		return val1 == val2
	case "str":
		val1 := ""
		val2 := ""
		if v1.StringValue != nil {
			val1 = *v1.StringValue
		}
		if v2.StringValue != nil {
			val2 = *v2.StringValue
		}
		return val1 == val2
	case "bool":
		val1 := false
		val2 := false
		if v1.BoolValue != nil {
			val1 = *v1.BoolValue
		}
		if v2.BoolValue != nil {
			val2 = *v2.BoolValue
		}
		return val1 == val2
	default:
		return false
	}
}

// NewIntValue creates a new integer value
func NewIntValue(val int64) *PrimitiveValue {
	return &PrimitiveValue{
		TypeName: "int",
		IntValue:  &val,
	}
}

// NewFloatValue creates a new float value
func NewFloatValue(val float64) *PrimitiveValue {
	return &PrimitiveValue{
		TypeName:   "float",
		FloatValue: &val,
	}
}

// NewStringValue creates a new string value
func NewStringValue(val string) *PrimitiveValue {
	return &PrimitiveValue{
		TypeName:   "str",
		StringValue: &val,
	}
}

// NewBoolValue creates a new boolean value
func NewBoolValue(val bool) *PrimitiveValue {
	return &PrimitiveValue{
		TypeName: "bool",
		BoolValue: &val,
	}
}

// SetValue represents a Set value
type SetValue struct {
	Elements map[string]Value // Use string key for set membership (hash-based)
	Values   []Value          // Keep order for iteration
}

func (v *SetValue) Type() string {
	return "set"
}

func (v *SetValue) String() string {
	if len(v.Values) == 0 {
		return "Set.empty()"
	}
	result := "Set{"
	for i, elem := range v.Values {
		if i > 0 {
			result += ", "
		}
		result += elem.String()
	}
	result += "}"
	return result
}

// NewSetValue creates a new empty set value
func NewSetValue() *SetValue {
	return &SetValue{
		Elements: make(map[string]Value),
		Values:   []Value{},
	}
}

// Add adds an element to the set
func (v *SetValue) Add(elem Value) {
	key := elem.String()
	if _, exists := v.Elements[key]; !exists {
		v.Elements[key] = elem
		v.Values = append(v.Values, elem)
	}
}

// Contains checks if an element is in the set
func (v *SetValue) Contains(elem Value) bool {
	key := elem.String()
	_, exists := v.Elements[key]
	return exists
}

// Size returns the size of the set
func (v *SetValue) Size() int64 {
	return int64(len(v.Values))
}

// ListValue represents a List value
type ListValue struct {
	Elements []Value
}

func (v *ListValue) Type() string {
	return "list"
}

func (v *ListValue) String() string {
	if len(v.Elements) == 0 {
		return "List.empty()"
	}
	result := "List["
	for i, elem := range v.Elements {
		if i > 0 {
			result += ", "
		}
		result += elem.String()
	}
	result += "]"
	return result
}

// NewListValue creates a new empty list value
func NewListValue() *ListValue {
	return &ListValue{
		Elements: []Value{},
	}
}

// Append adds an element to the end of the list
func (v *ListValue) Append(elem Value) {
	v.Elements = append(v.Elements, elem)
}

// Head returns the first element
func (v *ListValue) Head() (Value, error) {
	if len(v.Elements) == 0 {
		return nil, fmt.Errorf("cannot get head of empty list")
	}
	return v.Elements[0], nil
}

// Tail returns the list without the first element
func (v *ListValue) Tail() *ListValue {
	if len(v.Elements) == 0 {
		return NewListValue()
	}
	return &ListValue{
		Elements: v.Elements[1:],
	}
}

// Size returns the size of the list
func (v *ListValue) Size() int64 {
	return int64(len(v.Elements))
}

// MapValue represents a Map value
type MapValue struct {
	Entries map[string]Value // Key -> Value mapping
}

func (v *MapValue) Type() string {
	return "map"
}

func (v *MapValue) String() string {
	if len(v.Entries) == 0 {
		return "Map.empty()"
	}
	result := "Map{"
	first := true
	for k, val := range v.Entries {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s: %s", k, val.String())
		first = false
	}
	result += "}"
	return result
}

// NewMapValue creates a new empty map value
func NewMapValue() *MapValue {
	return &MapValue{
		Entries: make(map[string]Value),
	}
}

// Put adds or updates a key-value pair
func (v *MapValue) Put(key, val Value) {
	v.Entries[key.String()] = val
}

// Get retrieves a value by key
func (v *MapValue) Get(key Value) (Value, bool) {
	val, exists := v.Entries[key.String()]
	return val, exists
}

// Contains checks if a key exists in the map
func (v *MapValue) Contains(key Value) bool {
	_, exists := v.Entries[key.String()]
	return exists
}

// Size returns the size of the map
func (v *MapValue) Size() int64 {
	return int64(len(v.Entries))
}


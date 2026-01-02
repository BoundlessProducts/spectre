package types

import (
	"fmt"

	"github.com/akkeshavan/spectre/pkg/ast"
)

// Type represents a type in the Spectre type system
// This is separate from ast.Type which represents type syntax in the AST
type Type interface {
	// String returns a string representation of the type
	String() string
	// Equals checks if two types are equal
	Equals(other Type) bool
}

// Primitive represents a primitive type (int, bool, str, float)
type Primitive struct {
	Kind PrimitiveKind
}

type PrimitiveKind int

const (
	Int PrimitiveKind = iota
	Bool
	Str
	Float
)

func (p *Primitive) String() string {
	switch p.Kind {
	case Int:
		return "int"
	case Bool:
		return "bool"
	case Str:
		return "str"
	case Float:
		return "float"
	default:
		return "unknown"
	}
}

func (p *Primitive) Equals(other Type) bool {
	if otherPrim, ok := other.(*Primitive); ok {
		return p.Kind == otherPrim.Kind
	}
	return false
}

// Set represents a Set type (Set<T>)
type Set struct {
	Element Type
}

func (s *Set) String() string {
	return fmt.Sprintf("Set<%s>", s.Element.String())
}

func (s *Set) Equals(other Type) bool {
	if otherSet, ok := other.(*Set); ok {
		return s.Element.Equals(otherSet.Element)
	}
	return false
}

// Map represents a Map type (Map<K, V>)
type Map struct {
	Key   Type
	Value Type
}

func (m *Map) String() string {
	return fmt.Sprintf("Map<%s, %s>", m.Key.String(), m.Value.String())
}

func (m *Map) Equals(other Type) bool {
	if otherMap, ok := other.(*Map); ok {
		return m.Key.Equals(otherMap.Key) && m.Value.Equals(otherMap.Value)
	}
	return false
}

// List represents a List type (List<T>)
type List struct {
	Element Type
}

func (l *List) String() string {
	return fmt.Sprintf("List<%s>", l.Element.String())
}

func (l *List) Equals(other Type) bool {
	if otherList, ok := other.(*List); ok {
		return l.Element.Equals(otherList.Element)
	}
	return false
}

// Option represents an Optional type (Option<T>)
type Option struct {
	Element Type
}

func (o *Option) String() string {
	return fmt.Sprintf("Option<%s>", o.Element.String())
}

func (o *Option) Equals(other Type) bool {
	if otherOpt, ok := other.(*Option); ok {
		return o.Element.Equals(otherOpt.Element)
	}
	return false
}

// Record represents a record/struct type
type Record struct {
	Fields map[string]Type // Field name -> type
}

func (r *Record) String() string {
	if len(r.Fields) == 0 {
		return "{}"
	}
	result := "{"
	first := true
	for name, typ := range r.Fields {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s: %s", name, typ.String())
		first = false
	}
	result += "}"
	return result
}

func (r *Record) Equals(other Type) bool {
	if otherRec, ok := other.(*Record); ok {
		if len(r.Fields) != len(otherRec.Fields) {
			return false
		}
		for name, typ := range r.Fields {
			otherTyp, exists := otherRec.Fields[name]
			if !exists || !typ.Equals(otherTyp) {
				return false
			}
		}
		return true
	}
	return false
}

// Enum represents an enum type
type Enum struct {
	Name   string
	Values []string
}

func (e *Enum) String() string {
	return fmt.Sprintf("enum %s", e.Name)
}

func (e *Enum) Equals(other Type) bool {
	if otherEnum, ok := other.(*Enum); ok {
		if e.Name != otherEnum.Name {
			return false
		}
		if len(e.Values) != len(otherEnum.Values) {
			return false
		}
		for i, val := range e.Values {
			if val != otherEnum.Values[i] {
				return false
			}
		}
		return true
	}
	return false
}

// Named represents a named type (type alias)
type Named struct {
	Name string
	Base Type // The underlying type
}

func (n *Named) String() string {
	return n.Name
}

func (n *Named) Equals(other Type) bool {
	if otherNamed, ok := other.(*Named); ok {
		// Named types are equal if they have the same name
		// or if their base types are equal
		return n.Name == otherNamed.Name || n.Base.Equals(otherNamed.Base)
	}
	return false
}

// FromAST converts an AST type to a runtime Type
func FromAST(astType ast.Type) (Type, error) {
	switch t := astType.(type) {
	case *ast.PrimitiveType:
		return FromPrimitiveName(t.Name)
	case *ast.SetType:
		elem, err := FromAST(t.Element)
		if err != nil {
			return nil, err
		}
		return &Set{Element: elem}, nil
	case *ast.MapType:
		key, err := FromAST(t.Key)
		if err != nil {
			return nil, err
		}
		val, err := FromAST(t.Value)
		if err != nil {
			return nil, err
		}
		return &Map{Key: key, Value: val}, nil
	case *ast.ListType:
		elem, err := FromAST(t.Element)
		if err != nil {
			return nil, err
		}
		return &List{Element: elem}, nil
	case *ast.OptionType:
		elem, err := FromAST(t.Element)
		if err != nil {
			return nil, err
		}
		return &Option{Element: elem}, nil
	case *ast.RecordType:
		fields := make(map[string]Type)
		for _, field := range t.Fields {
			fieldType, err := FromAST(field.Type)
			if err != nil {
				return nil, err
			}
			fields[field.Name] = fieldType
		}
		return &Record{Fields: fields}, nil
	case *ast.EnumType:
		return &Enum{
			Name:   t.Name,
			Values: t.Values,
		}, nil
	case *ast.NamedType:
		// For named types, we'll need to resolve them later via the type environment
		// For now, return an error indicating resolution is needed
		return nil, fmt.Errorf("named type %s requires type environment for resolution", t.Name)
	default:
		return nil, fmt.Errorf("unknown AST type: %T", astType)
	}
}

// FromPrimitiveName creates a Primitive type from a string name
func FromPrimitiveName(name string) (*Primitive, error) {
	switch name {
	case "int":
		return &Primitive{Kind: Int}, nil
	case "bool":
		return &Primitive{Kind: Bool}, nil
	case "str":
		return &Primitive{Kind: Str}, nil
	case "float":
		return &Primitive{Kind: Float}, nil
	default:
		return nil, fmt.Errorf("unknown primitive type: %s", name)
	}
}

// IsAssignable checks if a value of type 'from' can be assigned to a variable of type 'to'
func IsAssignable(from, to Type) bool {
	// Exact match
	if from.Equals(to) {
		return true
	}

	// Option types: Option<T> can be assigned from T (wrapping) or Option<T>
	if optTo, ok := to.(*Option); ok {
		if from.Equals(optTo.Element) {
			return true // T can be assigned to Option<T>
		}
		if optFrom, ok := from.(*Option); ok {
			return IsAssignable(optFrom.Element, optTo.Element) // Option<T1> to Option<T2> if T1 to T2
		}
	}

	// Numeric conversions: int can be assigned to float
	if toPrim, ok := to.(*Primitive); ok && toPrim.Kind == Float {
		if fromPrim, ok := from.(*Primitive); ok && fromPrim.Kind == Int {
			return true
		}
	}

	return false
}


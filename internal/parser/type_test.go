package parser

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestPrimitiveTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int", "int"},
		{"bool", "bool"},
		{"str", "str"},
		{"float", "float"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		typ := p.parseType()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
		}

		prim, ok := typ.(*ast.PrimitiveType)
		if !ok {
			t.Fatalf("type not *ast.PrimitiveType. got=%T", typ)
		}

		if prim.Name != tt.expected {
			t.Errorf("prim.Name not %s. got=%s", tt.expected, prim.Name)
		}
	}
}

func TestSetType(t *testing.T) {
	input := "Set<int>"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	setType, ok := typ.(*ast.SetType)
	if !ok {
		t.Fatalf("type not *ast.SetType. got=%T", typ)
	}

	elementType, ok := setType.Element.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("element type not *ast.PrimitiveType. got=%T", setType.Element)
	}

	if elementType.Name != "int" {
		t.Errorf("element type name not 'int'. got=%s", elementType.Name)
	}
}

func TestMapType(t *testing.T) {
	input := "Map<int, str>"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	mapType, ok := typ.(*ast.MapType)
	if !ok {
		t.Fatalf("type not *ast.MapType. got=%T", typ)
	}

	keyType, ok := mapType.Key.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("key type not *ast.PrimitiveType. got=%T", mapType.Key)
	}

	if keyType.Name != "int" {
		t.Errorf("key type name not 'int'. got=%s", keyType.Name)
	}

	valueType, ok := mapType.Value.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("value type not *ast.PrimitiveType. got=%T", mapType.Value)
	}

	if valueType.Name != "str" {
		t.Errorf("value type name not 'str'. got=%s", valueType.Name)
	}
}

func TestListType(t *testing.T) {
	input := "List<int>"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	listType, ok := typ.(*ast.ListType)
	if !ok {
		t.Fatalf("type not *ast.ListType. got=%T", typ)
	}

	elementType, ok := listType.Element.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("element type not *ast.PrimitiveType. got=%T", listType.Element)
	}

	if elementType.Name != "int" {
		t.Errorf("element type name not 'int'. got=%s", elementType.Name)
	}
}

func TestOptionType(t *testing.T) {
	input := "Option<int>"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	optionType, ok := typ.(*ast.OptionType)
	if !ok {
		t.Fatalf("type not *ast.OptionType. got=%T", typ)
	}

	elementType, ok := optionType.Element.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("element type not *ast.PrimitiveType. got=%T", optionType.Element)
	}

	if elementType.Name != "int" {
		t.Errorf("element type name not 'int'. got=%s", elementType.Name)
	}
}

func TestEnumType(t *testing.T) {
	input := "enum Status { Pending, Active, Completed }"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	enumType, ok := typ.(*ast.EnumType)
	if !ok {
		t.Fatalf("type not *ast.EnumType. got=%T", typ)
	}

	if enumType.Name != "Status" {
		t.Errorf("enum name not 'Status'. got=%s", enumType.Name)
	}

	if len(enumType.Values) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(enumType.Values))
	}

	expectedValues := []string{"Pending", "Active", "Completed"}
	for i, expected := range expectedValues {
		if enumType.Values[i] != expected {
			t.Errorf("enum value[%d] not %s. got=%s", i, expected, enumType.Values[i])
		}
	}
}

func TestRecordType(t *testing.T) {
	input := "{ name: str, age: int }"

	l := lexer.New(input)
	p := New(l)
	typ := p.parseType()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	recordType, ok := typ.(*ast.RecordType)
	if !ok {
		t.Fatalf("type not *ast.RecordType. got=%T", typ)
	}

	if len(recordType.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordType.Fields))
	}

	// Check first field
	if recordType.Fields[0].Name != "name" {
		t.Errorf("first field name not 'name'. got=%s", recordType.Fields[0].Name)
	}
	nameType, ok := recordType.Fields[0].Type.(*ast.PrimitiveType)
	if !ok || nameType.Name != "str" {
		t.Errorf("first field type not 'str'")
	}

	// Check second field
	if recordType.Fields[1].Name != "age" {
		t.Errorf("second field name not 'age'. got=%s", recordType.Fields[1].Name)
	}
	ageType, ok := recordType.Fields[1].Type.(*ast.PrimitiveType)
	if !ok || ageType.Name != "int" {
		t.Errorf("second field type not 'int'")
	}
}

func TestNestedTypes(t *testing.T) {
	tests := []struct {
		input string
		test  func(ast.Type) bool
	}{
		{"Set<List<int>>", func(typ ast.Type) bool {
			setType, ok := typ.(*ast.SetType)
			if !ok {
				return false
			}
			_, ok = setType.Element.(*ast.ListType)
			return ok
		}},
		{"Map<str, List<int>>", func(typ ast.Type) bool {
			mapType, ok := typ.(*ast.MapType)
			if !ok {
				return false
			}
			_, ok = mapType.Value.(*ast.ListType)
			return ok
		}},
		{"Option<Set<int>>", func(typ ast.Type) bool {
			optionType, ok := typ.(*ast.OptionType)
			if !ok {
				return false
			}
			_, ok = optionType.Element.(*ast.SetType)
			return ok
		}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		typ := p.parseType()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors for %s: %v", len(p.Errors()), tt.input, p.Errors())
		}

		if !tt.test(typ) {
			t.Errorf("nested type test failed for: %s", tt.input)
		}
	}
}


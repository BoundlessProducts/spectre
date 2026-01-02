package types

import (
	"testing"

	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name     string
		kind     PrimitiveKind
		expected string
	}{
		{"int", Int, "int"},
		{"bool", Bool, "bool"},
		{"str", Str, "str"},
		{"float", Float, "float"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prim := &Primitive{Kind: tt.kind}
			if prim.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, prim.String())
			}
		})
	}
}

func TestPrimitiveEquality(t *testing.T) {
	int1 := &Primitive{Kind: Int}
	int2 := &Primitive{Kind: Int}
	bool1 := &Primitive{Kind: Bool}

	if !int1.Equals(int2) {
		t.Error("int1 should equal int2")
	}
	if int1.Equals(bool1) {
		t.Error("int1 should not equal bool1")
	}
}

func TestSetType(t *testing.T) {
	elem := &Primitive{Kind: Int}
	set := &Set{Element: elem}

	if set.String() != "Set<int>" {
		t.Errorf("expected Set<int>, got %s", set.String())
	}

	set2 := &Set{Element: elem}
	if !set.Equals(set2) {
		t.Error("set should equal set2")
	}

	set3 := &Set{Element: &Primitive{Kind: Bool}}
	if set.Equals(set3) {
		t.Error("set should not equal set3")
	}
}

func TestMapType(t *testing.T) {
	key := &Primitive{Kind: Str}
	val := &Primitive{Kind: Int}
	m := &Map{Key: key, Value: val}

	if m.String() != "Map<str, int>" {
		t.Errorf("expected Map<str, int>, got %s", m.String())
	}

	m2 := &Map{Key: key, Value: val}
	if !m.Equals(m2) {
		t.Error("m should equal m2")
	}

	m3 := &Map{Key: key, Value: &Primitive{Kind: Bool}}
	if m.Equals(m3) {
		t.Error("m should not equal m3")
	}
}

func TestListType(t *testing.T) {
	elem := &Primitive{Kind: Int}
	list := &List{Element: elem}

	if list.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", list.String())
	}

	list2 := &List{Element: elem}
	if !list.Equals(list2) {
		t.Error("list should equal list2")
	}
}

func TestOptionType(t *testing.T) {
	elem := &Primitive{Kind: Int}
	opt := &Option{Element: elem}

	if opt.String() != "Option<int>" {
		t.Errorf("expected Option<int>, got %s", opt.String())
	}

	opt2 := &Option{Element: elem}
	if !opt.Equals(opt2) {
		t.Error("opt should equal opt2")
	}
}

func TestRecordType(t *testing.T) {
	fields := map[string]Type{
		"x": &Primitive{Kind: Int},
		"y": &Primitive{Kind: Int},
	}
	rec := &Record{Fields: fields}

	str := rec.String()
	if str != "{x: int, y: int}" && str != "{y: int, x: int}" {
		t.Errorf("unexpected record string: %s", str)
	}

	rec2 := &Record{Fields: fields}
	if !rec.Equals(rec2) {
		t.Error("rec should equal rec2")
	}

	rec3 := &Record{Fields: map[string]Type{
		"x": &Primitive{Kind: Int},
		"y": &Primitive{Kind: Bool},
	}}
	if rec.Equals(rec3) {
		t.Error("rec should not equal rec3")
	}
}

func TestEnumType(t *testing.T) {
	enum := &Enum{
		Name:   "Status",
		Values: []string{"Active", "Inactive"},
	}

	if enum.String() != "enum Status" {
		t.Errorf("expected enum Status, got %s", enum.String())
	}

	enum2 := &Enum{
		Name:   "Status",
		Values: []string{"Active", "Inactive"},
	}
	if !enum.Equals(enum2) {
		t.Error("enum should equal enum2")
	}

	enum3 := &Enum{
		Name:   "Status",
		Values: []string{"Active"},
	}
	if enum.Equals(enum3) {
		t.Error("enum should not equal enum3")
	}
}

func TestFromASTPrimitive(t *testing.T) {
	tests := []struct {
		name     string
		astType  ast.Type
		expected string
	}{
		{"int", &ast.PrimitiveType{Name: "int"}, "int"},
		{"bool", &ast.PrimitiveType{Name: "bool"}, "bool"},
		{"str", &ast.PrimitiveType{Name: "str"}, "str"},
		{"float", &ast.PrimitiveType{Name: "float"}, "float"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := FromAST(tt.astType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if typ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, typ.String())
			}
		})
	}
}

func TestFromASTSet(t *testing.T) {
	astSet := &ast.SetType{
		Element: &ast.PrimitiveType{Name: "int"},
	}

	typ, err := FromAST(astSet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if typ.String() != "Set<int>" {
		t.Errorf("expected Set<int>, got %s", typ.String())
	}
}

func TestFromASTMap(t *testing.T) {
	astMap := &ast.MapType{
		Key:   &ast.PrimitiveType{Name: "str"},
		Value: &ast.PrimitiveType{Name: "int"},
	}

	typ, err := FromAST(astMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if typ.String() != "Map<str, int>" {
		t.Errorf("expected Map<str, int>, got %s", typ.String())
	}
}

func TestFromASTList(t *testing.T) {
	astList := &ast.ListType{
		Element: &ast.PrimitiveType{Name: "int"},
	}

	typ, err := FromAST(astList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}
}

func TestFromASTOption(t *testing.T) {
	astOpt := &ast.OptionType{
		Element: &ast.PrimitiveType{Name: "int"},
	}

	typ, err := FromAST(astOpt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if typ.String() != "Option<int>" {
		t.Errorf("expected Option<int>, got %s", typ.String())
	}
}

func TestFromASTRecord(t *testing.T) {
	astRec := &ast.RecordType{
		Fields: []ast.Field{
			{Name: "x", Type: &ast.PrimitiveType{Name: "int"}},
			{Name: "y", Type: &ast.PrimitiveType{Name: "int"}},
		},
	}

	typ, err := FromAST(astRec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec, ok := typ.(*Record)
	if !ok {
		t.Fatalf("expected *Record, got %T", typ)
	}

	if len(rec.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(rec.Fields))
	}

	if rec.Fields["x"].String() != "int" {
		t.Errorf("field x should be int")
	}
}

func TestFromASTEnum(t *testing.T) {
	astEnum := &ast.EnumType{
		Name:   "Status",
		Values: []string{"Active", "Inactive"},
	}

	typ, err := FromAST(astEnum)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	enum, ok := typ.(*Enum)
	if !ok {
		t.Fatalf("expected *Enum, got %T", typ)
	}

	if enum.Name != "Status" {
		t.Errorf("expected Status, got %s", enum.Name)
	}
}

func TestIsAssignable(t *testing.T) {
	intType := &Primitive{Kind: Int}
	boolType := &Primitive{Kind: Bool}
	floatType := &Primitive{Kind: Float}
	optInt := &Option{Element: intType}

	tests := []struct {
		name     string
		from     Type
		to       Type
		expected bool
	}{
		{"int to int", intType, intType, true},
		{"int to bool", intType, boolType, false},
		{"int to float", intType, floatType, true},
		{"int to Option<int>", intType, optInt, true},
		{"Option<int> to Option<int>", optInt, optInt, true},
		{"int to Option<bool>", intType, &Option{Element: boolType}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAssignable(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsAssignable(%s, %s) = %v, expected %v",
					tt.from.String(), tt.to.String(), result, tt.expected)
			}
		})
	}
}

func TestFromPrimitiveName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"int", "int", "int", false},
		{"bool", "bool", "bool", false},
		{"str", "str", "str", false},
		{"float", "float", "float", false},
		{"unknown", "unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := FromPrimitiveName(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if typ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, typ.String())
			}
		})
	}
}


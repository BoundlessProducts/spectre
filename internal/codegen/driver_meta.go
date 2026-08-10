package codegen

import (
	"encoding/json"
	"os"
	"strings"
)

// DriverMeta is the sidecar file written by `spectre mine` alongside the .spec.
// It records the exact field and method names from the Rust source so that
// `generate-driver` can emit a fully-connected adapter without any manual edits
// for primitive-typed structs.
//
// The sidecar is a snapshot of the source at mine time.  If the spec is edited
// after mining (variables renamed, synthetic actions added), the affected entries
// will be absent from the sidecar and generate-driver will emit annotated stubs
// for those entries only.
type DriverMeta struct {
	StructName string       `json:"struct"`
	SourceFile string       `json:"source"`
	Fields     []MetaField  `json:"fields"`
	Actions    []MetaAction `json:"actions"`
}

// MetaField maps one spec variable to the corresponding Rust struct field.
type MetaField struct {
	SpecVar     string `json:"spec_var"`     // name as it appears in the .spec
	RustField   string `json:"rust_field"`   // name as it appears in the Rust source
	RustType    string `json:"rust_type"`    // e.g. "i64", "bool", "String"
	SpectreType string `json:"spectre_type"` // e.g. "int", "bool", "str"
}

// MetaAction maps one spec action to the corresponding Rust method.
type MetaAction struct {
	SpecAction string      `json:"spec_action"` // name in .spec
	RustMethod string      `json:"rust_method"` // name in Rust source
	Params     []MetaParam `json:"params"`
}

// MetaParam is one parameter of an action/method.
type MetaParam struct {
	Name        string `json:"name"`
	RustType    string `json:"rust_type"`
	SpectreType string `json:"spectre_type"`
}

// MetaPath returns the expected sidecar path for a given spec file path.
// e.g. "bank-account.spec" → "bank-account.driver.json"
func MetaPath(specPath string) string {
	base := strings.TrimSuffix(specPath, ".spec")
	return base + ".driver.json"
}

// LoadDriverMeta reads the sidecar file.  Returns (nil, nil) if the file does
// not exist — callers treat absence as "no sidecar, use stub generation".
func LoadDriverMeta(path string) (*DriverMeta, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var meta DriverMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// spectreTypeToRust returns the canonical Rust primitive type for a Spectre type.
func spectreTypeToRust(spectreType string) string {
	switch spectreType {
	case "int":
		return "i64"
	case "bool":
		return "bool"
	case "str":
		return "String"
	case "float":
		return "f64"
	default:
		return spectreType // enum or named type — keep as-is
	}
}

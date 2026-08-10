package mine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ── JSON schema mirroring Rust MinedSpec output ───────────────────────────────

type asmSpec struct {
	SpecName   string      `json:"spec_name"`
	StructName string      `json:"struct_name"`
	SourceFile string      `json:"source_file"`
	Enums      []asmEnum   `json:"enums"`
	Fields     []asmField  `json:"fields"`
	Methods    []asmMethod `json:"methods"`
}

type asmEnum struct {
	Name     string   `json:"name"`
	Variants []string `json:"variants"`
}

type asmField struct {
	Name        string `json:"name"`
	RustType    string `json:"rust_type"`
	SpectreType string `json:"spectre_type"`
	Default     string `json:"default"`
}

type asmMethod struct {
	Name     string      `json:"name"`
	MutSelf  bool        `json:"mut_self"`
	Params   []asmParam  `json:"params"`
	Requires []string    `json:"requires"`
	Assigns  []asmAssign `json:"assigns"`
	BodyRaw  string      `json:"body_raw"`
}

type asmParam struct {
	Name        string `json:"name"`
	SpectreType string `json:"spectre_type"`
}

type asmAssign struct {
	Field   string `json:"field"`
	Op      string `json:"op"`
	RHSRust string `json:"rhs_rust"`
}

// ── Public entry point ────────────────────────────────────────────────────────

// MineFromRustFile attempts AST-based mining via the spectre-mine-rs helper
// binary, then falls back to the regex miner if the helper is unavailable or
// the file cannot be parsed.
func MineFromRustFile(filename, specName string) (*MinedSpec, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filename, err)
	}

	if spec, ok := mineWithSubprocess(filename, specName); ok {
		return spec, nil
	}

	// Regex fallback — always available, no external dependency.
	fmt.Fprintf(os.Stderr,
		"Note: spectre-mine-rs not found; using regex miner (run 'make install' for AST-based extraction).\n")
	return MineFromRust(string(content), specName, filepath.Base(filename)), nil
}

// ── Subprocess logic ──────────────────────────────────────────────────────────

func mineWithSubprocess(filename, specName string) (*MinedSpec, bool) {
	helperPath, found := findMineHelper()
	if !found {
		return nil, false
	}

	args := []string{filename}
	if specName != "" {
		args = append(args, "--spec-name", specName)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(helperPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "spectre-mine-rs: %s\n", stderr.String())
		}
		// Exit code 2 means the Rust file failed to parse — fall through to regex.
		return nil, false
	}

	var raw asmSpec
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: spectre-mine-rs produced invalid JSON: %v\n", err)
		return nil, false
	}

	return convertASMSpec(&raw), true
}

// findMineHelper looks for spectre-mine-rs next to the running binary, then in PATH.
func findMineHelper() (string, bool) {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "spectre-mine-rs")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}
	if path, err := exec.LookPath("spectre-mine-rs"); err == nil {
		return path, true
	}
	return "", false
}

// ── JSON → MinedSpec conversion ───────────────────────────────────────────────

func convertASMSpec(r *asmSpec) *MinedSpec {
	spec := &MinedSpec{
		SpecName:   r.SpecName,
		StructName: r.StructName,
		SourceFile: r.SourceFile,
	}

	for _, e := range r.Enums {
		spec.Enums = append(spec.Enums, RustEnum{Name: e.Name, Variants: e.Variants})
	}

	for _, f := range r.Fields {
		spec.Fields = append(spec.Fields, RustField{
			Name:        f.Name,
			RustType:    f.RustType,
			SpectreType: f.SpectreType,
			Default:     f.Default,
		})
	}

	for _, m := range r.Methods {
		method := RustMethod{
			Name:    m.Name,
			MutSelf: m.MutSelf,
			BodyRaw: m.BodyRaw,
		}
		for _, p := range m.Params {
			method.Params = append(method.Params, RustParam{Name: p.Name, SpectreType: p.SpectreType})
		}
		method.Requires = m.Requires
		for _, a := range m.Assigns {
			method.Assigns = append(method.Assigns, RustAssign{
				Field:   a.Field,
				Op:      a.Op,
				RHSRust: a.RHSRust,
			})
		}
		spec.Methods = append(spec.Methods, method)
	}

	return spec
}

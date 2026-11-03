package z3

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/backend"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Z3Backend runs bounded model checking via the Z3 binary as a subprocess.
type Z3Backend struct{}

func NewZ3Backend() *Z3Backend { return &Z3Backend{} }

func (b *Z3Backend) Name() backend.Kind { return backend.KindZ3 }

func (b *Z3Backend) Verify(file *ast.File, opts backend.Options) (*backend.Result, error) {
	depth := opts.BMCDepth
	if depth <= 0 {
		depth = 20
	}
	z3bin := opts.Z3Bin
	if z3bin == "" {
		z3bin = "z3"
	}
	if opts.Params == nil {
		opts.Params = make(map[string]string)
	}

	if _, err := exec.LookPath(z3bin); err != nil {
		return nil, fmt.Errorf("z3 not found — install Z3 or set --z3-bin (looked for %q)", z3bin)
	}

	smtSrc, err := Encode(file, opts.Params, depth)
	if err != nil {
		return nil, fmt.Errorf("z3: encode: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "spectre_z3_*")
	if err != nil {
		return nil, fmt.Errorf("z3: mktemp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	smtFile := filepath.Join(tmpDir, "spectre_bmc.smt2")
	if err := os.WriteFile(smtFile, []byte(smtSrc), 0600); err != nil {
		return nil, fmt.Errorf("z3: write smt2: %w", err)
	}

	cmd := exec.Command(z3bin, smtFile)
	out, runErr := cmd.CombinedOutput()
	output := string(out)

	// Z3 exits non-zero on "unsat" in some configurations; check content.
	if runErr != nil && !strings.Contains(output, "sat") {
		return nil, fmt.Errorf("z3 execution failed: %w\n%s", runErr, output)
	}

	note := fmt.Sprintf("bounded model checking to depth %d", depth)

	if strings.Contains(output, "unsat") {
		return &backend.Result{
			Safe:        true,
			Exhaustive:  false,
			BackendName: string(backend.KindZ3),
			Note:        note,
		}, nil
	}

	if strings.Contains(output, "sat") {
		v, modelNote := parseModel(output, file)
		fullNote := note
		if modelNote != "" {
			fullNote += "; " + modelNote
		}
		var violations []*explore.Violation
		if v != nil {
			violations = []*explore.Violation{v}
		}
		return &backend.Result{
			Safe:        false,
			Exhaustive:  false,
			Violations:  violations,
			BackendName: string(backend.KindZ3),
			Note:        fullNote,
		}, nil
	}

	return &backend.Result{
		Safe:        true,
		Exhaustive:  false,
		BackendName: string(backend.KindZ3),
		Note:        note + "; no sat/unsat in z3 output",
	}, nil
}

// reDefFun matches: (define-fun varname_N () Sort value)
var reDefFun = regexp.MustCompile(`\(define-fun\s+(\w+)_(\d+)\s+\(\)\s+\w+\s+([^)]+)\)`)

// parseModel extracts variable values from Z3 model output.
func parseModel(output string, file *ast.File) (*explore.Violation, string) {
	matches := reDefFun.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return &explore.Violation{Invariant: firstInvariantName(file)}, ""
	}

	maxStep := -1
	for _, m := range matches {
		if n, err := strconv.Atoi(m[2]); err == nil && n > maxStep {
			maxStep = n
		}
	}

	vals := map[string]string{}
	for _, m := range matches {
		if n, err := strconv.Atoi(m[2]); err == nil && n == maxStep {
			vals[m[1]] = strings.TrimSpace(m[3])
		}
	}

	parts := make([]string, 0, len(vals))
	for k, v := range vals {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	modelNote := fmt.Sprintf("counterexample at step %d: %s", maxStep, strings.Join(parts, ", "))

	return &explore.Violation{
		Invariant:   firstInvariantName(file),
		Description: modelNote,
	}, modelNote
}

func firstInvariantName(file *ast.File) string {
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok {
			return inv.Name
		}
	}
	return "invariant"
}

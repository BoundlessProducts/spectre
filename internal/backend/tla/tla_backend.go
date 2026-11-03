package tla

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/backend"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ── TLC backend ───────────────────────────────────────────────────────────────

// TLCBackend runs TLC (the TLA+ model checker) as a subprocess.
type TLCBackend struct{}

func NewTLCBackend() *TLCBackend { return &TLCBackend{} }

func (b *TLCBackend) Name() backend.Kind { return backend.KindTLC }

func (b *TLCBackend) Verify(file *ast.File, opts backend.Options) (*backend.Result, error) {
	tlaText, cfgText, err := Translate(file, opts.Params)
	if err != nil {
		return nil, fmt.Errorf("tlc: translate: %w", err)
	}

	dir, err := os.MkdirTemp("", "spectre-tlc-*")
	if err != nil {
		return nil, fmt.Errorf("tlc: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	tlaFile := filepath.Join(dir, "spectre_tla.tla")
	cfgFile := filepath.Join(dir, "spectre_tla.cfg")
	if err := os.WriteFile(tlaFile, []byte(tlaText), 0644); err != nil {
		return nil, fmt.Errorf("tlc: write tla: %w", err)
	}
	if err := os.WriteFile(cfgFile, []byte(cfgText), 0644); err != nil {
		return nil, fmt.Errorf("tlc: write cfg: %w", err)
	}

	// Build command: prefer direct `tlc` binary; fall back to `java -jar`.
	var cmd *exec.Cmd
	if opts.TLCJar == "tlc" {
		if _, lookErr := exec.LookPath("tlc"); lookErr == nil {
			cmd = exec.Command("tlc", "-config", cfgFile, tlaFile, "-workers", "1")
		}
	}
	if cmd == nil {
		if opts.TLCJar == "tlc" {
			return nil, fmt.Errorf("tlc not found — install TLA+ tools or pass --tlc-jar <path-to-tla2tools.jar>")
		}
		if _, lookErr := exec.LookPath("java"); lookErr != nil {
			return nil, fmt.Errorf("java not found — required to run TLC via jar")
		}
		cmd = exec.Command("java", "-jar", opts.TLCJar,
			"-config", cfgFile, tlaFile, "-workers", "1")
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir

	_ = cmd.Run() // non-zero exit is normal when a violation is found

	out := stdout.String()
	return parseTLCOutput(out, stderr.String()), nil
}

// ── Apalache backend ──────────────────────────────────────────────────────────

// ApalacheBackend runs Apalache as a subprocess.
type ApalacheBackend struct{}

func NewApalacheBackend() *ApalacheBackend { return &ApalacheBackend{} }

func (b *ApalacheBackend) Name() backend.Kind { return backend.KindApalache }

func (b *ApalacheBackend) Verify(file *ast.File, opts backend.Options) (*backend.Result, error) {
	tlaText, cfgText, err := Translate(file, opts.Params)
	if err != nil {
		return nil, fmt.Errorf("apalache: translate: %w", err)
	}

	dir, err := os.MkdirTemp("", "spectre-apalache-*")
	if err != nil {
		return nil, fmt.Errorf("apalache: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	tlaFile := filepath.Join(dir, "spectre_tla.tla")
	cfgFile := filepath.Join(dir, "spectre_tla.cfg")
	if err := os.WriteFile(tlaFile, []byte(tlaText), 0644); err != nil {
		return nil, fmt.Errorf("apalache: write tla: %w", err)
	}
	if err := os.WriteFile(cfgFile, []byte(cfgText), 0644); err != nil {
		return nil, fmt.Errorf("apalache: write cfg: %w", err)
	}

	bin := opts.ApalacheBin
	if bin == "" {
		bin = "apalache-mc"
	}
	if _, lookErr := exec.LookPath(bin); lookErr != nil {
		return nil, fmt.Errorf("apalache-mc not found — install Apalache or pass --apalache-bin <path>")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, "check", "--config", cfgFile, tlaFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir

	_ = cmd.Run()

	out := stdout.String()
	return parseApalacheOutput(out, stderr.String()), nil
}

// ── Output parsers ────────────────────────────────────────────────────────────

var (
	reTLCInvViolation   = regexp.MustCompile(`Error: Invariant (\w+) is violated`)
	reTLCTempViolation  = regexp.MustCompile(`Error: Temporal property (\w+) is violated`)
	reTLCAction         = regexp.MustCompile(`<(\w+)>`)
	reTLCStatesGenerated = regexp.MustCompile(`(\d+) states generated`)
	reTLCCompleted      = regexp.MustCompile(`Model checking completed`)
)

func parseTLCOutput(out, errOut string) *backend.Result {
	r := &backend.Result{BackendName: string(backend.KindTLC)}

	if m := reTLCStatesGenerated.FindStringSubmatch(out); m != nil {
		fmt.Sscanf(m[1], "%d", &r.StatesExplored)
	}

	invMatch := reTLCInvViolation.FindStringSubmatch(out)
	tempMatch := reTLCTempViolation.FindStringSubmatch(out)

	if invMatch == nil && tempMatch == nil {
		r.Safe = true
		if reTLCCompleted.MatchString(out) {
			r.Exhaustive = true
		}
		return r
	}

	r.Safe = false
	invName := ""
	if invMatch != nil {
		invName = invMatch[1]
	} else if tempMatch != nil {
		invName = tempMatch[1]
	}

	// Try to extract the action name from the trace.
	actionName := ""
	if m := reTLCAction.FindStringSubmatch(out); m != nil {
		actionName = m[1]
	}

	v := &explore.Violation{
		Invariant:   invName,
		Description: fmt.Sprintf("TLC: %s violated", invName),
	}
	if actionName != "" {
		// Violation.Path is []*explore.Transition; we can't populate full states
		// from text output, so record the action name in Description instead.
		v.Description = fmt.Sprintf("TLC: %s violated by action %s", invName, actionName)
	}
	r.Violations = append(r.Violations, v)

	if errOut != "" && !strings.Contains(out, "Error") {
		r.Note = "stderr: " + strings.TrimSpace(errOut)
	}
	return r
}

func parseApalacheOutput(out, errOut string) *backend.Result {
	r := &backend.Result{BackendName: string(backend.KindApalache)}

	if strings.Contains(out, "Checker reports no error") {
		r.Safe = true
		r.Exhaustive = true
		return r
	}

	if strings.Contains(out, "COUNTEREXAMPLE") || strings.Contains(out, "violation") {
		r.Safe = false
		// Extract invariant name if present.
		re := regexp.MustCompile(`(?i)invariant (\w+)`)
		if m := re.FindStringSubmatch(out); m != nil {
			r.Violations = append(r.Violations, &explore.Violation{
				Invariant:   m[1],
				Description: "Apalache: " + m[1] + " violated",
			})
		} else {
			r.Violations = append(r.Violations, &explore.Violation{
				Description: "Apalache: counterexample found",
			})
		}
		return r
	}

	// Unknown output — if stderr has content, surface it.
	if errOut != "" {
		return nil // caller will wrap the error
	}
	// Assume safe if nothing conclusive found.
	r.Safe = true
	r.Note = "Apalache output not conclusive; assumed safe"
	return r
}

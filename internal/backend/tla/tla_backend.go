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

// tlaModuleFile is the filename used for all TLA+ temp files.
// It must match the MODULE name inside the file (SANY requirement).
const tlaModuleFile = "SpectreSpec.tla"

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

	tlaFile := filepath.Join(dir, tlaModuleFile)
	cfgFile := filepath.Join(dir, "SpectreSpec.cfg")
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
	tlaText, _, err := Translate(file, opts.Params)
	if err != nil {
		return nil, fmt.Errorf("apalache: translate: %w", err)
	}

	dir, err := os.MkdirTemp("", "spectre-apalache-*")
	if err != nil {
		return nil, fmt.Errorf("apalache: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Filename must match MODULE name for SANY.
	tlaFile := filepath.Join(dir, tlaModuleFile)
	if err := os.WriteFile(tlaFile, []byte(tlaText), 0644); err != nil {
		return nil, fmt.Errorf("apalache: write tla: %w", err)
	}

	bin := opts.ApalacheBin
	if bin == "" {
		bin = "apalache-mc"
	}
	if _, lookErr := exec.LookPath(bin); lookErr != nil {
		return nil, fmt.Errorf("apalache-mc not found — install Apalache or pass --apalache-bin <path>")
	}

	// Collect invariant names from the AST for --inv= flags.
	var invNames []string
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok {
			invNames = append(invNames, inv.Name)
		}
	}

	// Apalache v0.62+ requires = syntax; --inv takes a comma-separated list.
	args := []string{"check", "--init=Init", "--next=Next"}
	if len(invNames) > 0 {
		args = append(args, "--inv="+strings.Join(invNames, ","))
	}
	args = append(args, tlaFile)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir
	cmd.Env = apalacheEnv()

	_ = cmd.Run()

	out := stdout.String()
	result := parseApalacheOutput(out, stderr.String())
	if result == nil {
		return nil, fmt.Errorf("apalache: %s", strings.TrimSpace(stderr.String()))
	}
	return result, nil
}

// apalacheEnv builds an environment that sets JAVA_HOME and JVM_GC_ARGS
// for Apalache subprocesses, while preserving all other env vars.
func apalacheEnv() []string {
	env := os.Environ()
	hasJavaHome, hasGCArgs := false, false
	for _, e := range env {
		if strings.HasPrefix(e, "JAVA_HOME=") {
			hasJavaHome = true
		}
		if strings.HasPrefix(e, "JVM_GC_ARGS=") {
			hasGCArgs = true
		}
	}
	if !hasJavaHome {
		for _, candidate := range []string{
			"/opt/homebrew/opt/openjdk",
			"/opt/homebrew/opt/openjdk@21",
			"/usr/local/opt/openjdk",
			"/usr/local/opt/openjdk@21",
		} {
			if _, statErr := os.Stat(candidate); statErr == nil {
				env = append(env, "JAVA_HOME="+candidate)
				break
			}
		}
	}
	if !hasGCArgs {
		env = append(env, "JVM_GC_ARGS=-XX:+UseG1GC")
	}
	return env
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

	// Apalache v0.62+ outcome strings.
	if strings.Contains(out, "The outcome is: NoError") ||
		strings.Contains(out, "Checker reports no error") {
		r.Safe = true
		// Bounded check — not exhaustive unless the state space is fully covered.
		r.Exhaustive = false
		return r
	}

	if strings.Contains(out, "The outcome is: Error") ||
		strings.Contains(out, "Checker has found an error") {
		r.Safe = false

		// Map invariant index → name from "Set an invariant to X" lines.
		reInvSet := regexp.MustCompile(`Set an invariant to (\w+)`)
		invMatches := reInvSet.FindAllStringSubmatch(out, -1)

		reViolated := regexp.MustCompile(`state invariant (\d+) violated`)
		seen := map[string]bool{}
		for _, m := range reViolated.FindAllStringSubmatch(out, -1) {
			var idx int
			fmt.Sscanf(m[1], "%d", &idx)
			name := fmt.Sprintf("invariant_%d", idx)
			if idx < len(invMatches) {
				name = invMatches[idx][1]
			}
			if !seen[name] {
				seen[name] = true
				r.Violations = append(r.Violations, &explore.Violation{
					Invariant:   name,
					Description: "Apalache: " + name + " violated",
				})
			}
		}
		if len(r.Violations) == 0 {
			r.Violations = append(r.Violations, &explore.Violation{
				Description: "Apalache: counterexample found",
			})
		}
		return r
	}

	// Unknown output — surface stderr if present.
	if errOut != "" {
		return nil
	}
	r.Safe = true
	r.Note = "Apalache output not conclusive; assumed safe"
	return r
}

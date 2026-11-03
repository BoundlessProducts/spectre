package main

// commands_tla.go — Item 13: TLA+ export with translation limits and source-mapped diagnostics.

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/tla"
)

func init() {
	Commands["export-tla"] = &Command{
		Name:        "export-tla",
		Description: "Translate a Spectre spec to TLA+, reporting translation limits and source-mapped diagnostics",
		Run:         runExportTLA,
	}
}

func runExportTLA(args []string) error {
	fs := flag.NewFlagSet("export-tla", flag.ContinueOnError)
	output := fs.String("output", "", "Write TLA+ to this file (default: stdout)")
	showDiags := fs.Bool("show-diagnostics", true, "Print translation diagnostics to stderr")
	failOnUnsupported := fs.Bool("fail-on-unsupported", false, "Exit non-zero if any UNSUPPORTED diagnostics are emitted")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		return fmt.Errorf("usage: spectre export-tla [--output <out.tla>] [--show-diagnostics] [--fail-on-unsupported] <spec.spec>")
	}
	filename := fs.Arg(0)

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s:\n  %s", filename, strings.Join(p.Errors(), "\n  "))
	}

	result := tla.Translate(file, filename)

	// Emit the TLA+ output.
	if *output == "" {
		fmt.Print(result.TLA)
	} else {
		if err := os.WriteFile(*output, []byte(result.TLA), 0644); err != nil {
			return fmt.Errorf("error writing %s: %w", *output, err)
		}
		fmt.Fprintf(os.Stderr, "Wrote TLA+ to %s\n", *output)
	}

	// Report diagnostics.
	if *showDiags && len(result.Diagnostics) > 0 {
		nUnsupported, nPartial := 0, 0
		for _, d := range result.Diagnostics {
			switch d.Kind {
			case tla.LimitUnsupported:
				nUnsupported++
			case tla.LimitPartial:
				nPartial++
			}
			fmt.Fprintf(os.Stderr, "[%s] %s\n", d.Kind, d.Message)
			if d.Position.Line > 0 {
				fmt.Fprintf(os.Stderr, "  → %s:%d:%d\n", filename, d.Position.Line, d.Position.Column)
			}
		}
		fmt.Fprintf(os.Stderr, "\nTranslation summary: %d UNSUPPORTED, %d PARTIAL\n", nUnsupported, nPartial)
		if *failOnUnsupported && nUnsupported > 0 {
			return fmt.Errorf("translation produced %d unsupported construct(s)", nUnsupported)
		}
	} else if len(result.Diagnostics) == 0 && *showDiags {
		fmt.Fprintln(os.Stderr, "Translation complete: no limits encountered.")
	}

	return nil
}

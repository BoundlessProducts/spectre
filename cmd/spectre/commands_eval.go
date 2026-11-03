package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/mine"
)

func init() {
	Commands["eval-rust"] = &Command{
		Name:        "eval-rust",
		Description: "Evaluate Rust source files for extraction coverage and print a Markdown report",
		Run:         runEvalRust,
	}
}

// runEvalRust evaluates one or more Rust source files (or all *.rs files under
// a directory given via --dir) and prints a Markdown extraction report.
func runEvalRust(args []string) error {
	dir := ""
	var files []string

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			}
		default:
			files = append(files, args[i])
		}
		i++
	}

	if dir != "" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("cannot read directory %q: %w", dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".rs") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no Rust files specified (use --dir <dir> or pass .rs files explicitly)")
	}

	summary, err := mine.EvaluateFiles(files)
	if err != nil {
		return err
	}

	fmt.Print(mine.FormatReport(summary))
	return nil
}

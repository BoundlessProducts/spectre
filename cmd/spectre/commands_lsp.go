package main

// commands_lsp.go — Item 17: editor diagnostics, completion, hover, and go-to-definition.

import (
	"fmt"
	"os"

	"github.com/BoundlessProducts/spectre/internal/lsp"
)

func init() {
	Commands["lsp"] = &Command{
		Name:        "lsp",
		Description: "Start an LSP server over stdin/stdout (diagnostics, hover, completion, definition)",
		Run:         runLSP,
	}
}

func runLSP(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Println("spectre lsp")
		fmt.Println("  Starts a Language Server Protocol server over stdin/stdout.")
		fmt.Println("  Supported capabilities:")
		fmt.Println("    textDocument/publishDiagnostics — parse errors pushed on open/change")
		fmt.Println("    textDocument/hover              — type info for identifiers")
		fmt.Println("    textDocument/completion         — declared names (vars, actions, fns, enums)")
		fmt.Println("    textDocument/definition         — go-to-declaration position")
		fmt.Println()
		fmt.Println("  Configure your editor to run: spectre lsp")
		return nil
	}
	srv := lsp.NewServer(os.Stdin, os.Stdout)
	return srv.Run()
}

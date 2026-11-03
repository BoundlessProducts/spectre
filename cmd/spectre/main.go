package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Version is set at build time via -ldflags "-X main.Version=<tag>".
// The fallback "dev" is used for local builds without ldflags.
var Version = "dev"

func main() {
	// Check for version flag.
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("spectre version %s\n", Version)
		os.Exit(ExitSuccess)
	}

	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(ExitUsage)
	}

	commandName := flag.Arg(0)
	args := flag.Args()[1:]

	cmd, exists := Commands[commandName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", commandName)
		printUsage()
		os.Exit(ExitUsage)
	}

	if err := cmd.Run(args); err != nil {
		errStr := err.Error()
		// Verification violations are reported inline; exit with ExitViolations.
		if strings.Contains(errStr, "verification failed") ||
			strings.Contains(errStr, "test(s) failed") {
			os.Exit(ExitViolations)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitError)
	}
	os.Exit(ExitSuccess)
}


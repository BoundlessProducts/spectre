package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const Version = "0.1.0"

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("spectre version %s\n", Version)
		os.Exit(0)
	}

	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	commandName := flag.Arg(0)
	args := flag.Args()[1:]

	cmd, exists := Commands[commandName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", commandName)
		printUsage()
		os.Exit(1)
	}

	if err := cmd.Run(args); err != nil {
		// Check if it's a verification failure - if so, already reported, just exit
		errStr := err.Error()
		if strings.Contains(errStr, "verification failed") {
			// Verification failures are already reported with clean output
			os.Exit(1)
		}
		// For actual errors, show the error message
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}


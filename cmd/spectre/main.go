package main

import (
	"flag"
	"fmt"
	"os"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}


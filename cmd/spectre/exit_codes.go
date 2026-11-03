package main

// exit_codes.go — Versioned, consistent CLI exit codes.
//
// Exit code policy (also documented in the help text):
//   0  — Success: command completed, no violations found.
//   1  — Violations: verification found invariant or temporal violations.
//   2  — Error: command failed due to an I/O, parse, or internal error.
//   3  — Usage: invalid arguments or unknown command.

const (
	ExitSuccess    = 0
	ExitViolations = 1
	ExitError      = 2
	ExitUsage      = 3
)

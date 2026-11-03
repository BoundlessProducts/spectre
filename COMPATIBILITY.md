# Compatibility Policy

## Versioning

Spectre follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

| Component | When it changes |
|-----------|----------------|
| `MAJOR`   | Breaking changes to the spec language, CLI interface, or public Go API |
| `MINOR`   | New features that are backwards-compatible |
| `PATCH`   | Bug fixes and performance improvements |

Pre-1.0 releases (`0.x.y`) may include breaking changes in minor versions.

## Spec Language Stability

Spec language constructs marked **stable** will not change semantics across
minor versions. Constructs marked **experimental** may change in any release.

Currently **stable**: `var`, `param`, `init`, `action`, `invariant`, `enum`,
`require`, primed assignments.

Currently **experimental**: `temporal` properties, `liveness` keywords.

## CLI Interface Stability

All top-level commands (`verify`, `mine`, `generate-driver`, `eval-rust`,
`validate-monitor`, `benchmark`) are stable in their flag names and exit codes.
New flags may be added in minor releases.

Exit codes are defined in `exit_codes.go` and will not change in minor releases:

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | Verification violation(s) found |
| 2    | Usage error |
| 3    | Internal / runtime error |

## Go Package API

The `internal/` packages are not part of the public API and may change in any
release. If you depend on them, pin to a specific tag.

## Binary Releases

Pre-built binaries are published on the GitHub Releases page for each tag.
Supported platforms: `linux/amd64`, `linux/arm64`, `darwin/amd64`,
`darwin/arm64`, `windows/amd64`.

## Minimum Go Version

The `go.mod` file specifies the minimum Go version required to build Spectre.
We target the two most recent stable Go releases.

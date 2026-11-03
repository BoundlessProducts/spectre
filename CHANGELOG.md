# Changelog

All notable changes to Spectre are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.3.3] — 2026-10-01
### Changed
- Zenodo DOI updated to `10.5281/zenodo.21897571`.

## [0.3.2] — 2026-09-15
### Added
- `spectre eval-rust` command: evaluate Rust source files for extraction coverage and print a Markdown report.
- `internal/mine`: `EvaluateFile`, `EvaluateFiles`, `FormatReport` for programmatic evaluation.
- Four Rust testdata files: `counter.rs`, `bank_account.rs`, `state_machine.rs`, `resource_pool.rs`.
- `spectre validate-monitor` command: replay violation traces and confirm invariant violations.
- `spectre benchmark` command: BFS timing table over spec files in `benchmarks/`.
- `spectre generate-driver --ci`: emit `{stem}.cargo.toml` and `{stem}.ci.yml` alongside the driver.
- `spectre verify --suggest-repairs`: label and validate repair candidates for invariant violations.
- Extraction omission tracking (`mine.ScanOmissions`, `--show-omissions` flag).
- `internal/repair`: `SuggestRepairs`, `ValidateRepair`, `PatchSpec`.
- `internal/monitor`: `ValidateTrace`, `ValidateAll`, `Summary`.
- Five BFS benchmark spec files under `benchmarks/`.

### Changed
- Version is now injected at build time via `-ldflags "-X main.Version=<tag>"`.

## [0.3.1] — 2026-08-01
### Added
- Property-directed model-based testing.
- Incremental model checking with state-hash caching.
- Semantic sync with 30/30 Quint mutation coverage.

## [0.1.0] — 2025-01-01
### Added
- Initial public release: BFS-based safety verification, Quint-like spec language, Rust driver generation.

.PHONY: build install uninstall test clean release

# Build the binary
build:
	go build -o spectre ./cmd/spectre

# Install locally (for development)
install: build
	go install ./cmd/spectre

# Uninstall
uninstall:
	go clean -i ./cmd/spectre

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f spectre
	rm -f coverage.out coverage.html

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/spectre-linux-amd64 ./cmd/spectre
	GOOS=linux GOARCH=arm64 go build -o dist/spectre-linux-arm64 ./cmd/spectre
	GOOS=darwin GOARCH=amd64 go build -o dist/spectre-darwin-amd64 ./cmd/spectre
	GOOS=darwin GOARCH=arm64 go build -o dist/spectre-darwin-arm64 ./cmd/spectre
	GOOS=windows GOARCH=amd64 go build -o dist/spectre-windows-amd64.exe ./cmd/spectre
	GOOS=windows GOARCH=arm64 go build -o dist/spectre-windows-arm64.exe ./cmd/spectre

# Release using GoReleaser
release:
	goreleaser release --snapshot

release-production:
	goreleaser release

# Install Homebrew formula locally (for testing)
install-homebrew-local:
	brew install --build-from-source Formula/spectre.rb


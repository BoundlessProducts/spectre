.PHONY: build build-go build-rs install uninstall test clean release docker-build docker-test artifact-zip

GOBIN          := $(shell go env GOPATH)/bin
MINE_RS_DIR    := rust/spectre-mine-rs
MINE_RS_BIN    := $(MINE_RS_DIR)/target/release/spectre-mine-rs

# Build both binaries into the repo root
build: build-go build-rs

build-go:
	go build -o spectre ./cmd/spectre

build-rs: $(MINE_RS_BIN)

$(MINE_RS_BIN): $(MINE_RS_DIR)/src/main.rs $(MINE_RS_DIR)/Cargo.toml
	cargo build --manifest-path $(MINE_RS_DIR)/Cargo.toml --release

# Install both binaries to $GOPATH/bin (single command for reviewers)
install: build
	go install ./cmd/spectre
	install -m755 $(MINE_RS_BIN) $(GOBIN)/spectre-mine-rs
	@echo "Installed spectre and spectre-mine-rs to $(GOBIN)"

# Uninstall
uninstall:
	go clean -i ./cmd/spectre
	rm -f $(GOBIN)/spectre-mine-rs

# Run tests
test:
	go test ./...
	cargo test --manifest-path $(MINE_RS_DIR)/Cargo.toml

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f spectre
	rm -f coverage.out coverage.html
	cargo clean --manifest-path $(MINE_RS_DIR)/Cargo.toml

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

# Docker artifact (for VMCAI submission)
DOCKER_IMAGE := spectre-vmcai

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-test: docker-build
	@echo "=== spectre --help ==="
	docker run --rm $(DOCKER_IMAGE) --help
	@echo "\n=== full benchmark suite (all Table 2 & 3 entries) ==="
	docker run --rm --entrypoint sh $(DOCKER_IMAGE) /artifact/reproduce.sh

# Build the Zenodo upload zip (excludes git history and LaTeX build output)
ARTIFACT_NAME := spectre-vmcai-2027
artifact-zip:
	@echo "Building artifact zip for Zenodo upload..."
	git archive --format=zip --prefix=$(ARTIFACT_NAME)/ HEAD \
	    --add-file=v2-paper-vmcai-27/main.pdf \
	    -o $(ARTIFACT_NAME).zip
	@echo "Created $(ARTIFACT_NAME).zip"
	@ls -lh $(ARTIFACT_NAME).zip


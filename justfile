set shell := ["bash", "-euo", "pipefail", "-c"]
set dotenv-load := true

[private]
default:
    @just --list --unsorted

# ------------------------------------------------------------------ #
# Development                                                          #
# ------------------------------------------------------------------ #

# Run the server
[group("development")]
run:
    go run ./cmd/mithras serve | hl -P

# Format all code
[group("development")]
fmt:
    go tool goimports -w .
    gofmt -w .

# Check formatting without modifying files
[group("development")]
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Code is not formatted. Run 'just fmt'" && exit 1)

# Run linter
[group("development")]
lint:
    golangci-lint run ./...

# Run all checks (fmt + lint)
[group("development")]
check: fmt-check lint

# Run tests
[group("development")]
test *args:
    go test ./... {{ args }}

# Run tests with race detector
[group("development")]
test-race *args:
    go test -race ./... {{ args }}

# Generate all code (OpenAPI, SQL, etc)
[group("development")]
gen:
    go generate ./...

# ------------------------------------------------------------------ #
# Security                                                             #
# ------------------------------------------------------------------ #

# Run vulnerability check
[group("security")]
vuln:
    go tool govulncheck ./...

# Run static application security testing
[group("security")]
sast:
    go tool gosec ./...

# Run all security checks
[group("security")]
security: vuln sast

# ------------------------------------------------------------------ #
# Build                                                                #
# ------------------------------------------------------------------ #

# Debug build
[group("build")]
build:
    go build -o bin/ ./...

# Install globally
[group("build")]
install:
    go install ./...

# ------------------------------------------------------------------ #
# Release                                                              #
# ------------------------------------------------------------------ #

# Tag and push — run after bump PR is merged
[group("release")]
release version:
    git tag -a "v{{ version }}" -m "Release v{{ version }}"
    git push origin "v{{ version }}"
    @echo "✓ Tagged and pushed v{{ version }}"

# ------------------------------------------------------------------ #
# Dependency management                                                #
# ------------------------------------------------------------------ #

# Tidy go.mod
[group("dependency management")]
mod-tidy:
    go mod tidy

# Verify dependencies
[group("dependency management")]
mod-verify:
    go mod verify

# Download all dependencies
[group("dependency management")]
mod-download:
    go mod download

# ------------------------------------------------------------------ #
# CI                                                                   #
# ------------------------------------------------------------------ #

# Run the full CI pipeline locally
[group("ci")]
ci: check test-race security
    @echo "✓ CI pipeline passed"

# ------------------------------------------------------------------ #
# Housekeeping                                                         #
# ------------------------------------------------------------------ #

# Remove build artifacts
[group("misc")]
clean:
    go clean ./...
    rm -rf bin/ coverage.out result

# Remove go cache
[group("misc")]
clean-cache:
    go clean -cache -modcache

# Remove everything
[group("misc")]
clean-all: clean clean-cache

# Print tool versions
[group("misc")]
versions:
    @echo "go:            $(go version)"
    @echo "gopls:         $(go tool gopls version | head -1)"
    @echo "golangci-lint: $(golangci-lint version --short)"
    @echo "gosec:         $(go tool gosec --version 2>&1)"
    @echo "govulncheck:   $(go tool govulncheck -version)"
    @echo "just:          $(just --version)"

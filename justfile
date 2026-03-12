set shell := ["bash", "-euo", "pipefail", "-c"]
set dotenv-load := true

[private]
default:
    @just --list --unsorted

# ------------------------------------------------------------------ #
# Development                                                          #
# ------------------------------------------------------------------ #

# Format all code
[group("development")]
fmt:
    goimports -w ./..
    gofmt -w ./..

# Check formatting without modifying files
[group("development")]
fmt-check:
    test -z "$(gofmt -l ./..)"

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

# ------------------------------------------------------------------ #
# Security                                                             #
# ------------------------------------------------------------------ #

# Run vulnerability check
[group("security")]
vuln:
    govulncheck ./...

# Run static application security testing
[group("security")]
sast:
    gosec ./...

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

# Release build (optimized)
[group("build")]
build-release:
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ ./...

# Build the main binary (mithras)
[group("build")]
build-mithras:
    go build -o bin/mithras ./cmd/mithras

# Install globally
[group("build")]
install:
    go install ./...

# ------------------------------------------------------------------ #
# Release                                                              #
# ------------------------------------------------------------------ #

# Verify go.mod module version matches a given tag
[group("release")]
verify-version tag:
    #!/usr/bin/env bash
    TAG_VERSION="{{ tag }}"
    TAG_VERSION="${TAG_VERSION#v}"
    echo "✓ Version check passed for tag {{ tag }}"

# Tag and push a release (e.g. just release 0.2.0)
[group("release")]
release version:
    git tag -a "v{{ version }}" -m "Release v{{ version }}"
    git push origin "v{{ version }}"
    @echo "✓ Tagged and pushed v{{ version }}"

# Build release and create tag
[group("release")]
publish version:
    @just build-release
    @just release {{ version }}
    @echo "✓ Release {{ version }} published"

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

# Show outdated dependencies
[group("dependency management")]
outdated:
    go list -u -m all | grep '\['

# Download all dependencies
[group("dependency management")]
mod-download:
    go mod download

# ------------------------------------------------------------------ #
# CI                                                                   #
# ------------------------------------------------------------------ #

# Run the full CI pipeline locally
ci: check test vuln sast
    @echo "✓ CI pipeline passed"

# ------------------------------------------------------------------ #
# Housekeeping                                                         #
# ------------------------------------------------------------------ #

# Remove build artifacts
[group("misc")]
clean:
    go clean ./...
    rm -rf bin/ coverage.out

# Remove go cache
[group("misc")]
clean-cache:
    go clean -cache -modcache

# Remove everything
[group("misc")]
clean-all: clean clean-cache

# Print tool versions (useful for debugging CI vs local discrepancies)
versions:
    @echo "go:             $(go version)"
    @echo "golangci-lint:  $(golangci-lint version)"
    @echo "sqlc:           $(sqlc version)"
    @echo "gosec:          $(gosec --version)"
    @echo "govulncheck:    $(govulncheck -version)"
    @echo "just:           $(just --version)"

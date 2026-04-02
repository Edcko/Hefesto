# Hefesto Makefile
# Cross-platform build and release automation

BINARY=hefesto
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
CMD_DIR=cmd/hefesto
DIST_DIR=dist
PWD=$(shell pwd)

# Platform targets: GOOS/GOARCH
TARGETS=darwin-arm64 darwin-amd64 linux-arm64 linux-amd64

# Build flags
LDFLAGS=-s -w -X main.version=$(VERSION)
GOFLAGS=-trimpath

.PHONY: all build clean sha256sum release test fmt lint help

all: build

## build: Build binaries for all platforms
build:
	@echo "Building $(BINARY) $(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for target in $(TARGETS); do \
		GOOS=$${target%-*} GOARCH=$${target#*-} go build -C $(CMD_DIR) $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$(PWD)/$(DIST_DIR)/$(BINARY)-$$target" .; \
		echo "Built $(BINARY)-$$target"; \
	done
	@echo "✅ Build complete!"

## sha256sum: Generate SHA256 checksums for all binaries
sha256sum: build
	@echo "Generating checksums..."
	@rm -f checksums.txt
	@for target in $(TARGETS); do \
		binary="$(DIST_DIR)/$(BINARY)-$$target"; \
		if command -v shasum >/dev/null 2>&1; then \
			shasum -a 256 "$$binary" >> checksums.txt; \
		elif command -v sha256sum >/dev/null 2>&1; then \
			sha256sum "$$binary" >> checksums.txt; \
		fi; \
	done
	@echo "✅ Checksums generated in checksums.txt"

## release: Create GitHub release with binaries and checksums
release: sha256sum
	@echo "Creating GitHub release $(VERSION)..."
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "❌ Error: gh CLI not installed. Install from: https://cli.github.com/"; \
		exit 1; \
	fi
	@if gh release view $(VERSION) >/dev/null 2>&1; then \
		echo "⚠️  Release $(VERSION) already exists. Deleting and recreating..."; \
		gh release delete $(VERSION) --yes; \
	fi
	@echo "## Hefesto $(VERSION)" > /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "TUI installer for OpenCode configuration." >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "### Installation" >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "\`\`\`bash" >> /tmp/release-notes.md
	@echo "brew install edcko/tap/hefesto" >> /tmp/release-notes.md
	@echo "\`\`\`" >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "### Binaries" >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "| Platform | Architecture | Binary |" >> /tmp/release-notes.md
	@echo "|----------|-------------|--------|" >> /tmp/release-notes.md
	@echo "| macOS | ARM64 (M1/M2/M3) | \`hefesto-darwin-arm64\` |" >> /tmp/release-notes.md
	@echo "| macOS | AMD64 (Intel) | \`hefesto-darwin-amd64\` |" >> /tmp/release-notes.md
	@echo "| Linux | ARM64 | \`hefesto-linux-arm64\` |" >> /tmp/release-notes.md
	@echo "| Linux | AMD64 | \`hefesto-linux-amd64\` |" >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "### Checksums" >> /tmp/release-notes.md
	@echo "" >> /tmp/release-notes.md
	@echo "See \`checksums.txt\` for SHA256 checksums." >> /tmp/release-notes.md
	gh release create $(VERSION) \
		--title "Hefesto $(VERSION)" \
		--notes-file /tmp/release-notes.md \
		$(DIST_DIR)/$(BINARY)-darwin-arm64 \
		$(DIST_DIR)/$(BINARY)-darwin-amd64 \
		$(DIST_DIR)/$(BINARY)-linux-arm64 \
		$(DIST_DIR)/$(BINARY)-linux-amd64 \
		checksums.txt
	@rm /tmp/release-notes.md
	@echo "✅ Release $(VERSION) created successfully!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@rm -f checksums.txt
	@echo "✅ Clean complete!"

## test: Run tests
test:
	@echo "Running tests..."
	go test -C $(CMD_DIR) -v ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt -C $(CMD_DIR) ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(CMD_DIR) && golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Skipping..."; \
	fi

## help: Show this help message
help:
	@echo "Hefesto Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

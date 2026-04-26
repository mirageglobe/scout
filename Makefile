.PHONY: help build run test fmt clean lint bump-patch bump-minor bump-major release release-dry

# Default target
.DEFAULT_GOAL := help

# derive version from latest git tag; fallback to "dev" if no tag exists
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "dev")

# compute next patch version: v1.2.3 -> v1.2.4; falls back to v0.1.0 if no tag exists
NEXT_VERSION := $(shell \
	tag=$$(git describe --tags --abbrev=0 2>/dev/null); \
	if [ -z "$$tag" ]; then echo "v0.1.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		minor=$$(echo $$tag | sed 's/^v//' | cut -d. -f2); \
		patch=$$(echo $$tag | sed 's/^v//' | cut -d. -f3); \
		echo "v$$major.$$minor.$$((patch + 1))"; \
	fi)

# compute next minor version: v1.2.3 -> v1.3.0
NEXT_MINOR_VERSION := $(shell \
	tag=$$(git describe --tags --abbrev=0 2>/dev/null); \
	if [ -z "$$tag" ]; then echo "v0.1.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		minor=$$(echo $$tag | sed 's/^v//' | cut -d. -f2); \
		echo "v$$major.$$((minor + 1)).0"; \
	fi)

# compute next major version: v1.2.3 -> v2.0.0
NEXT_MAJOR_VERSION := $(shell \
	tag=$$(git describe --tags --abbrev=0 2>/dev/null); \
	if [ -z "$$tag" ]; then echo "v1.0.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		echo "v$$((major + 1)).0.0"; \
	fi)

help: ## Display this help menu
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the scout binary
	go build -ldflags "-X github.com/mirageglobe/scout/internal/ui.Version=$(VERSION)" -o scout cmd/scout/main.go

run: build ## Build and run scout locally
	./scout

test: ## Run Go tests
	go test -v ./...

fmt: ## Format the Go source code
	go fmt ./...

lint: ## Run go vet (basic linting)
	go vet ./...

demo: build ## Generate a VHS demo GIF
	vhs < demo.tape

bump-patch: ## [main only] Tag the next patch version (e.g. v0.1.2 -> v0.1.3) on the current commit
	@echo "current: v$(VERSION)  ->  next: $(NEXT_VERSION)"
	@read -p "tag $(NEXT_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_VERSION) && echo "tagged $(NEXT_VERSION)" || echo "aborted"

bump-minor: ## [main only] Tag the next minor version (e.g. v0.1.3 -> v0.2.0) on the current commit
	@echo "current: v$(VERSION)  ->  next: $(NEXT_MINOR_VERSION)"
	@read -p "tag $(NEXT_MINOR_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_MINOR_VERSION) && echo "tagged $(NEXT_MINOR_VERSION)" || echo "aborted"

bump-major: ## [main only] Tag the next major version (e.g. v0.2.0 -> v1.0.0) on the current commit
	@echo "current: v$(VERSION)  ->  next: $(NEXT_MAJOR_VERSION)"
	@read -p "tag $(NEXT_MAJOR_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_MAJOR_VERSION) && echo "tagged $(NEXT_MAJOR_VERSION)" || echo "aborted"

push-tags: ## [main only] Push local tags to origin (run after bump-*, before release)
	git push origin --tags

release: ## [main only] Build and publish via goreleaser (requires GITHUB_TOKEN + HOMEBREW_TAP_GITHUB_TOKEN)
	goreleaser release --clean

release-reset: ## [main only] Delete existing GitHub release for current tag (use before re-running a failed release)
	gh release delete v$(VERSION) --yes 2>/dev/null || true

release-dry: ## Dry-run goreleaser locally without publishing (safe on any branch)
	goreleaser release --snapshot --clean

clean: ## Remove the compiled binary and demo assets
	rm -f scout demo.gif

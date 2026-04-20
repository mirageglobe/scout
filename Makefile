.PHONY: help build run test fmt clean lint release release-dry

# Default target
.DEFAULT_GOAL := help

help: ## Display this help menu
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the scout binary
	go build -o scout cmd/scout/main.go

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

release: ## Tag and release via goreleaser (requires GITHUB_TOKEN + HOMEBREW_TAP_GITHUB_TOKEN)
	goreleaser release --clean

release-dry: ## Dry-run goreleaser release (no publish, no tag)
	goreleaser release --snapshot --clean

clean: ## Remove the compiled binary and demo assets
	rm -f scout demo.gif

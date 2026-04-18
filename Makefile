.PHONY: help build run test fmt clean lint

# Default target
.DEFAULT_GOAL := help

help: ## Display this help menu
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the scout binary
	go build -o scout .

run: build ## Build and run scout locally
	./scout

test: ## Run Go tests
	go test -v ./...

fmt: ## Format the Go source code
	go fmt ./...

lint: ## Run go vet (basic linting)
	go vet ./...

clean: ## Remove the compiled binary
	rm -f scout

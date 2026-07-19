# ================================================================ info ===== #

# project : scout
# author  : Jimmy MG Lim

# ======================================================= configuration ===== #

.DEFAULT_GOAL := help

.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:

.PHONY: help build run test fmt clean lint version bump-patch bump-minor bump-major push-tags release release-reset release-dry release-check demo site-install site-build site-preview site-dev

# ============================================================== targets ===== #

# derive version from the highest semver tag; fallback to "dev" if no tag exists.
# use `git tag --sort` not `git describe`: describe only sees tags reachable from
# HEAD, so a tag left orphaned by a rebase (e.g. v0.8.0) would be skipped and the
# bump math would anchor to a stale, lower version.
VERSION := $(shell v=$$(git tag --list 'v*' --sort=-v:refname | head -n1 | sed 's/^v//'); echo "$${v:-dev}")

# compute next patch version: v1.2.3 -> v1.2.4; falls back to v0.1.0 if no tag exists
NEXT_VERSION := $(shell \
	tag=$$(git tag --list 'v*' --sort=-v:refname | head -n1); \
	if [ -z "$$tag" ]; then echo "v0.1.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		minor=$$(echo $$tag | sed 's/^v//' | cut -d. -f2); \
		patch=$$(echo $$tag | sed 's/^v//' | cut -d. -f3); \
		echo "v$$major.$$minor.$$((patch + 1))"; \
	fi)

# compute next minor version: v1.2.3 -> v1.3.0
NEXT_MINOR_VERSION := $(shell \
	tag=$$(git tag --list 'v*' --sort=-v:refname | head -n1); \
	if [ -z "$$tag" ]; then echo "v0.1.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		minor=$$(echo $$tag | sed 's/^v//' | cut -d. -f2); \
		echo "v$$major.$$((minor + 1)).0"; \
	fi)

# compute next major version: v1.2.3 -> v2.0.0
NEXT_MAJOR_VERSION := $(shell \
	tag=$$(git tag --list 'v*' --sort=-v:refname | head -n1); \
	if [ -z "$$tag" ]; then echo "v1.0.0"; \
	else \
		major=$$(echo $$tag | sed 's/^v//' | cut -d. -f1); \
		echo "v$$((major + 1)).0.0"; \
	fi)

# ----------------------------------------------------------------- meta ----- #

help: ## show this menu
	@printf "\n  \033[33mscout\033[0m\n"
	@printf "\n  usage: make <target>\n\n"
	@awk '/^##@/ { printf "\n  \033[1m%s\033[0m\n", substr($$0, 5) } /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-15s\033[0m %s\n", substr($$1, 1, length($$1)-1), substr($$0, index($$0, "##")+3) }' $(MAKEFILE_LIST)
	@printf "\n"

version: ## show tool versions
	@printf "\n  [ versions ]\n\n"
	@printf "  make : $$(make --version | head -1)\n"
	@printf "  git  : $$(git --version)\n"
	@printf "  go   : $$(go version)\n"
	@printf "\n"

##@ build

build: ## compile the scout binary
	go build -ldflags "-X github.com/mirageglobe/scout/internal/ui.Version=$(VERSION)" -o scout cmd/scout/main.go

run: build ## build and run scout locally
	./scout

demo: ## record demo.gif from a throwaway HOME so the header renders ~/scout (no local-path leak)
	tmp=$$(mktemp -d)
	trap 'rm -rf "$$tmp"' EXIT
	git clone -q . "$$tmp/scout"
	cd "$$tmp/scout"
	go build -o scout ./cmd/scout
	printf '\n' >> README.md   # dirty a tracked file so the M badge shows
	: > SCRATCH.txt             # untracked file so the ? badge shows
	HOME="$$tmp" vhs demo.tape
	cp demo.gif "$(CURDIR)/demo.gif"
	echo "[ ok ] demo.gif recorded from a throwaway HOME (header renders ~/scout)"

##@ verify

test: ## run Go tests
	go test -v ./...

fmt: ## format Go source code
	go fmt ./...

lint: ## run go vet
	go vet ./...

##@ release

bump-patch: ## tag next patch version (e.g. v0.1.2 -> v0.1.3)
	@echo "current: v$(VERSION)  ->  next: $(NEXT_VERSION)"
	@read -p "tag $(NEXT_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_VERSION) && echo "tagged $(NEXT_VERSION)" || echo "aborted"

bump-minor: ## tag next minor version (e.g. v0.1.3 -> v0.2.0)
	@echo "current: v$(VERSION)  ->  next: $(NEXT_MINOR_VERSION)"
	@read -p "tag $(NEXT_MINOR_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_MINOR_VERSION) && echo "tagged $(NEXT_MINOR_VERSION)" || echo "aborted"

bump-major: ## tag next major version (e.g. v0.2.0 -> v1.0.0)
	@echo "current: v$(VERSION)  ->  next: $(NEXT_MAJOR_VERSION)"
	@read -p "tag $(NEXT_MAJOR_VERSION)? [y/N] " ans && [ "$$ans" = "y" ] && \
		git tag $(NEXT_MAJOR_VERSION) && echo "tagged $(NEXT_MAJOR_VERSION)" || echo "aborted"

push-tags: ## push local tags to origin
	git push origin --tags

release: ## publish via goreleaser (requires GITHUB_TOKEN)
	goreleaser release --clean

release-reset: ## delete GitHub release for current tag
	gh release delete v$(VERSION) --yes 2>/dev/null || true

release-dry: ## dry-run goreleaser without publishing
	goreleaser release --snapshot --clean

release-check: ## verify the next patch/minor/major tags are unused on origin (run before bump)
	@git fetch --tags --quiet origin 2>/dev/null || true; for t in $(NEXT_VERSION) $(NEXT_MINOR_VERSION) $(NEXT_MAJOR_VERSION); do if git ls-remote --tags --exit-code origin "refs/tags/$$t" >/dev/null 2>&1; then printf "  [ fail ] %s already exists on origin; bump math would re-publish it\n" "$$t"; exit 1; fi; done; printf "  [ ok ] next tags free on origin: %s / %s / %s\n" "$(NEXT_VERSION)" "$(NEXT_MINOR_VERSION)" "$(NEXT_MAJOR_VERSION)"

##@ site

site-install: ## install site dependencies
	cd site && npm install

site-dev: ## start astro dev server with hot reload
	mkdir -p site/public && cp demo.gif site/public/demo.gif
	cd site && npm run dev

site-build: ## build the astro site
	mkdir -p site/public && cp demo.gif site/public/demo.gif
	cd site && npm run build

site-preview: ## serve the built site (no hot reload; use site-dev for development)
	mkdir -p site/public && cp demo.gif site/public/demo.gif
	cd site && npm run build && npm run preview

##@ clean

clean: ## remove the compiled binary and demo assets
	rm -f scout demo.gif

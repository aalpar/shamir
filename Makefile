# Version from VERSION file
BUILD_VERSION ?= $(shell cat ./VERSION | tr -d '[:space:]')

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	"$(GOLANGCI_LINT)" run --fix

##@ CI/CD

.PHONY: ci
ci: lint build test ## Full CI validation (lint, build, test).

##@ Release

.PHONY: tag
tag: ## Create annotated git tag from VERSION file.
	git tag -a "$(BUILD_VERSION)" -m "Release $(BUILD_VERSION)"
	@echo "Tagged $(BUILD_VERSION). Push with: git push origin $(BUILD_VERSION)"

.PHONY: bump-major
bump-major: ## Bump major version in VERSION file.
	./tools/sh/bump-version.sh major

.PHONY: bump-minor
bump-minor: ## Bump minor version in VERSION file.
	./tools/sh/bump-version.sh minor

.PHONY: bump-patch
bump-patch: ## Bump patch version in VERSION file.
	./tools/sh/bump-version.sh patch

.PHONY: release-check
release-check: ## Validate .goreleaser.yml syntax.
	goreleaser check

.PHONY: release-snapshot
release-snapshot: ## Local GoReleaser snapshot build (no publish).
	goreleaser release --snapshot --clean

.PHONY: release
release: ## Full GoReleaser release (publishes to GitHub).
	goreleaser release --clean

##@ Build

.PHONY: build
build: fmt vet ## Build CLI binary.
	go build -o bin/shamir ./cmd/shamir/

.PHONY: clean
clean: ## Remove build artifacts.
	chmod -R u+w $(LOCALBIN) 2>/dev/null; rm -rf $(LOCALBIN)
	rm -rf dist/
	rm -f cover.out

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.7.2

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef

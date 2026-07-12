#-----------------------------------------------------------------------------------------------------------------------
# Variables (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
#-----------------------------------------------------------------------------------------------------------------------
BINARIES_DIR = $(CURDIR)/bin
COVERAGE_DIR = $(CURDIR)/coverage

#-----------------------------------------------------------------------------------------------------------------------
# Help (default goal — `make` with no args prints the target catalogue)
#-----------------------------------------------------------------------------------------------------------------------
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message and exit
	@awk 'BEGIN {FS = ":.*?## "; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?## / { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

#-----------------------------------------------------------------------------------------------------------------------
# Tooling (SHA-pinned helpers installed into ./bin on first use)
#-----------------------------------------------------------------------------------------------------------------------
$(BINARIES_DIR)/golangci-lint:
	@echo "==> Installing golangci-lint within ${BINARIES_DIR}"
	@GOBIN=$(BINARIES_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@c0d3ddc9cf3faa61a4e378e879ece580256d76e5 # v2.12.2

$(BINARIES_DIR)/commitlint:
	@echo "==> Installing commitlint within ${BINARIES_DIR}"
	@GOBIN=$(BINARIES_DIR) go install github.com/conventionalcommit/commitlint@e9a606ce7074ac884ea091765be1651be18356d4 # v0.10.1

$(BINARIES_DIR)/govulncheck:
	@echo "==> Installing govulncheck within ${BINARIES_DIR}"
	@GOBIN=$(BINARIES_DIR) go install golang.org/x/vuln/cmd/govulncheck@0782b76014f15f24e22a438f30f308df42899ba1 # v1.3.0

#-----------------------------------------------------------------------------------------------------------------------
# Test (https://pkg.go.dev/testing — unit + coverage + testcontainers integration)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: test
test: ## Run unit tests with the race detector
	@go test -race -count=1 ./...

.PHONY: test-cover
test-cover: ## Run unit tests with coverage -> coverage/unit.out
	@mkdir -p $(COVERAGE_DIR)
	@echo "==> Running unit tests with coverage"
	@go test -race -count=1 -coverprofile=$(COVERAGE_DIR)/unit.out -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1

.PHONY: coverage
coverage: test-cover ## Run tests with coverage and print the total
	@printf "  total: %s\n" "$$(go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1 | awk '{print $$3}')"

.PHONY: integration
integration: ## Run the testcontainers + godog integration suite (requires Docker)
	@echo "==> Running integration suite (test/integration)"
	@cd test/integration && go test -count=1 -v ./...

#-----------------------------------------------------------------------------------------------------------------------
# Lint & security (golangci-lint, commitlint, govulncheck)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: lint
lint: $(BINARIES_DIR)/golangci-lint ## Run golangci-lint over every module (with --fix)
	@echo "==> Running golangci-lint"
	@$(BINARIES_DIR)/golangci-lint run --fix -c .golangci.yaml ./...
	@echo "==> Running golangci-lint (dsl)"
	@cd dsl && $(BINARIES_DIR)/golangci-lint run --fix -c $(CURDIR)/.golangci.yaml ./...
	@echo "==> Running golangci-lint (test/integration)"
	@cd test/integration && $(BINARIES_DIR)/golangci-lint run --fix -c $(CURDIR)/.golangci.yaml ./...

.PHONY: lint-commits
lint-commits: $(BINARIES_DIR)/commitlint ## Lint the current commit message against commitlint.yaml
	@$(BINARIES_DIR)/commitlint lint

.PHONY: vuln
vuln: $(BINARIES_DIR)/govulncheck ## Scan the root, dsl and integration module graphs for known Go vulnerabilities
	@echo "==> Scanning module graph for known Go vulnerabilities"
	@$(BINARIES_DIR)/govulncheck ./...
	@echo "==> Scanning dsl module graph for known Go vulnerabilities"
	@cd dsl && $(BINARIES_DIR)/govulncheck ./...
	@echo "==> Scanning test/integration module graph for known Go vulnerabilities"
	@cd test/integration && $(BINARIES_DIR)/govulncheck ./...

#-----------------------------------------------------------------------------------------------------------------------
# Housekeeping (formatting + module hygiene + git hooks)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: fmt
fmt: ## Format all Go sources with gofmt -s
	@gofmt -s -w .

.PHONY: tidy
tidy: ## Tidy the module graph for the root, dsl and integration modules
	@go mod tidy
	@cd dsl && go mod tidy
	@cd test/integration && go mod tidy

.PHONY: pre-commit
pre-commit: ## Install local pre-commit, commit-msg and pre-push hooks
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "'pre-commit' is not installed. Install with one of:"; \
		echo "  pipx install pre-commit       # recommended on PEP-668 distros (Arch, Debian 12+)"; \
		echo "  brew install pre-commit       # macOS"; \
		echo "  pip install --user pre-commit # any other Python environment"; \
		exit 1; \
	fi
	@pre-commit install --hook-type pre-commit --hook-type commit-msg --hook-type pre-push
	@echo "==> pre-commit hooks installed"

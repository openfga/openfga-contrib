#-----------------------------------------------------------------------------------------------------------------------
# Variables (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
#-----------------------------------------------------------------------------------------------------------------------
.DEFAULT_GOAL := help

BUILD_DIR ?= $(CURDIR)/dist
GO_BIN ?= $(shell go env GOPATH)/bin
GO_PACKAGES := $(shell go list ./... | grep -vE "vendor")

# Colors for the printf
RESET = $(shell tput sgr0)
COLOR_WHITE = $(shell tput setaf 7)
COLOR_BLUE = $(shell tput setaf 4)
TEXT_ENABLE_STANDOUT = $(shell tput smso)
TEXT_DISABLE_STANDOUT = $(shell tput rmso)

#-----------------------------------------------------------------------------------------------------------------------
# Rules (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: help clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Clean project files
	${call print, "Cleaning solution"}
	@rm -rf "${BUILD_DIR}"
	@go clean -x -r -i

#-----------------------------------------------------------------------------------------------------------------------
# Dependencies
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: deps

deps: ## Download dependencies
	${call print, "Downloading dependencies"}
	@go mod vendor && go mod tidy

$(GO_BIN)/golangci-lint:
	${call print, "Installing golangci-lint within ${GO_BIN}"}
	@go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest

#-----------------------------------------------------------------------------------------------------------------------
# Building & Installing
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: build install

build: ## Build the OpenFGA plugins. Build directory can be overridden using BUILD_DIR="desired/path", default is ".dist/". Usage `BUILD_DIR="." make build`
	${call print, "Building the OpenFGA plugins"}
	@go build -buildmode=plugin -v -o "$(BUILD_DIR)/plugins/storage/mssql/mssql.so" "$(CURDIR)/plugins/storage/mssql"
	@go build -buildmode=plugin -v -o "$(BUILD_DIR)/plugins/storage/sqlite/sqlite.so" "$(CURDIR)/plugins/storage/sqlite"
	@go build -buildmode=plugin -v -o "$(BUILD_DIR)/plugins/middleware/middleware.so" "$(CURDIR)/plugins/middleware"

#-----------------------------------------------------------------------------------------------------------------------
# Checks
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: lint

lint: $(GO_BIN)/golangci-lint ## Lint Go source files
	${call print, "Linting Go source files"}
	@golangci-lint run -v --fix -c .golangci.yaml ./...

#-----------------------------------------------------------------------------------------------------------------------
# Tests
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: test

test: ## Run all tests. To run a specific test, pass the FILTER var. Usage `make test FILTER="TestCheckLogs"`
	${call print, "Running tests"}
	@go test -race \
			-run "$(FILTER)" \
			-coverpkg=./... \
			-coverprofile=coverageunit.tmp.out \
			-covermode=atomic \
			-count=1 \
			-timeout=10m \
			${GO_PACKAGES}
	@cat coverageunit.tmp.out | grep -v "mocks" > coverageunit.out
	@rm coverageunit.tmp.out

#-----------------------------------------------------------------------------------------------------------------------
# Helpers
#-----------------------------------------------------------------------------------------------------------------------
define print
	@printf "${TEXT_ENABLE_STANDOUT}${COLOR_WHITE} 🚀 ${COLOR_BLUE} %-70s ${COLOR_WHITE} ${TEXT_DISABLE_STANDOUT}\n" $(1)
endef

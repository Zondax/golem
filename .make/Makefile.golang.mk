GOCMD=go
GOTEST=$(GOCMD) test

# Get all directories under cmd that contain a file main.go, but only 1 level deep
CMDS=$(shell find cmd -maxdepth 1 -type d -exec test -e '{}/main.go' \; -print)

# Strip cmd/ from directory names and generate output binary names
BINS=$(subst cmd/,output/,$(CMDS))

DEFAULT_APP_NAME ?= api

$(info Using $(shell nproc) CPUs)

## Golang
mod-tidy: ## Mod tidy
	@go mod tidy

mod-update: ## Mod Update
	@go get -u ./...

.PHONY: generate
generate: mod-tidy ## Mod generate
	@go generate ./internal/...

output/%: cmd/% FORCE | generate
	@echo "$(GREEN)Building $(notdir $<) binary...$(RESET)"
	@mkdir -p $(dir $@)
	@go build -o $@ ./$<

# Special target that forces rules to always run
.PHONY: FORCE
FORCE:

list: ## List all available binaries
	@for cmd in $(CMDS); do echo $$cmd; done

build: generate $(BINS) ## Build
	@# Check for main.go in the root and build if it exists too
	@[ -e main.go ] && go build -v -o output/$(DEFAULT_APP_NAME) || true
	
run: build ## Run
	@echo "Running $(DEFAULT_APP_NAME)"
	./output/$(DEFAULT_APP_NAME) start

version: build ## Get Version
	./output/$(DEFAULT_APP_NAME) version

clean: ## Go Clean
	go clean
	rm -rf internal/domain/entities/*

check-modtidy: ## Check Modtidy
	go mod tidy
	git diff --exit-code -- go.mod go.sum

lint: ## Lint
	golangci-lint --version
	golangci-lint run

# Dependency helpers
install-lint: ## Install go linter `golangci-lint`
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest

## Test
test: ## Run the tests of the project, excluding integration tests.
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/jstemmer/go-junit-report
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | go-junit-report -set-exit-code > junit-report.xml)
endif
	$(GOTEST) -v -race $(shell go list ./... | grep -v "/internal/tests/integration/") $(OUTPUT_OPTIONS)

coverage: ## Run the tests of the project and export the coverage, options: $EXPORT_RESULT
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/AlekSi/gocov-xml
	GO111MODULE=off go get -u github.com/axw/gocov/gocov
	gocov convert profile.cov | gocov-xml > coverage.xml
endif

.PHONY: *

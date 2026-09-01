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
GOLANGCI_LINT_VERSION ?= 2.13.2

# Installed by hand rather than via golangci-lint's install.sh: that script looks
# up the expected hash with `grep <tarball-name>` over the checksums file, which
# since v2.12.0 also matches the `<tarball-name>.sbom.json` entry. It then compares
# both hashes at once against the one real hash and always fails.
install-lint: ## Install go linter `golangci-lint`
	@set -eu; \
	ver="$(GOLANGCI_LINT_VERSION)"; \
	name="golangci-lint-$$ver-$$(go env GOOS)-$$(go env GOARCH)"; \
	base="https://github.com/golangci/golangci-lint/releases/download/v$$ver"; \
	bin="$$(go env GOPATH)/bin"; \
	tmp="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp"' EXIT; \
	curl -sSfL "$$base/$$name.tar.gz" -o "$$tmp/$$name.tar.gz"; \
	curl -sSfL "$$base/golangci-lint-$$ver-checksums.txt" -o "$$tmp/checksums.txt"; \
	want="$$(awk -v f="$$name.tar.gz" '$$2 == f { print $$1 }' "$$tmp/checksums.txt")"; \
	[ -n "$$want" ] || { echo "no checksum entry for $$name.tar.gz"; exit 1; }; \
	if command -v sha256sum >/dev/null 2>&1; then \
		got="$$(sha256sum "$$tmp/$$name.tar.gz" | awk '{ print $$1 }')"; \
	else \
		got="$$(shasum -a 256 "$$tmp/$$name.tar.gz" | awk '{ print $$1 }')"; \
	fi; \
	[ "$$want" = "$$got" ] || { echo "checksum mismatch for $$name.tar.gz: want $$want, got $$got"; exit 1; }; \
	tar -xzf "$$tmp/$$name.tar.gz" -C "$$tmp"; \
	mkdir -p "$$bin"; \
	install -m 0755 "$$tmp/$$name/golangci-lint" "$$bin/golangci-lint"

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

# Names expected by zondax/_workflows/_checks-golang.yaml
.PHONY: go-build go-mod-check go-lint-install go-lint go-test go-coverage
go-build: build
go-mod-check: check-modtidy
go-lint-install: install-lint
go-lint: lint
go-test: test
go-coverage: coverage

.PHONY: *

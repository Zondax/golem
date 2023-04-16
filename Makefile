#
# Generated by @zondax/cli
#
-include Makefile.settings.mk

# Get all directories under cmd
CMDS=$(shell find cmd -type d)

# Strip cmd/ from directory names and generate output binary names
BINS=$(subst cmd/,output/,$(CMDS))

default: build

mod-tidy:
	@go mod tidy

mod-update:
	@go get -u ./...

generate: mod-tidy
	go generate ./internal/...

output/%: cmd/%
	mkdir -p $(dir $@)
	go build -o $@ ./$<

build: generate $(BINS)
	@[ -e main.go ] && go build -v -o output/$(APP_NAME) || true

run: build
	./output/$(APP_NAME) start

version: build
	./output/$(APP_NAME) version

clean:
	go clean

gitclean:
	git clean -xfd
	git submodule foreach --recursive git clean -xfd

install_lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest

check-modtidy:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

lint:
	golangci-lint --version
	golangci-lint run

earthly:
	earthly +all

docker-bash:
	docker run --platform linux/amd64 -it zondax/${APP_NAME}:latest /bin/sh

docker-run:
	docker run --platform linux/amd64 -it zondax/${APP_NAME}:latest

########################################

zondax-update:
	@npx -y @zondax/cli@latest update

# include if exists
-include Makefile.local.mk

.PHONY: *

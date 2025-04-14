REMOTE_BUILD_KIT?=tcp://buildkit.zondax.dev:8372
REMOTE_BUILD_KIT_NAME?=zondax-builder

PLATFORMS?=linux/amd64,linux/arm64
DOCKER_IMAGE_NAME?=zondax/unknown

DOCKER_BUILD_PROGRESS?=auto

# Default to empty (builds final stage), can be overridden with a specific stage name
DOCKER_BUILD_TARGET?=

# Check if docker-bake.hcl exists
HAS_BAKE_FILE := $(shell test -f docker-bake.hcl && echo "YES" || echo "NO")

# Push settings for different methods
DOCKER_PUSH_ARGS_BAKE := --set "*.output=type=registry,push=true"
DOCKER_PUSH_ARGS_LEGACY := --push --output type=registry,registry.addr=docker.io

# Load settings for different methods
DOCKER_LOAD_ARGS_BAKE := --set "*.output=type=docker"
DOCKER_LOAD_ARGS_LEGACY := --load

# Automatically select the appropriate setter based on presence of bake file
DOCKER_PUSH_ARGS := $(if $(filter YES,$(HAS_BAKE_FILE)),$(DOCKER_PUSH_ARGS_BAKE),$(DOCKER_PUSH_ARGS_LEGACY))
DOCKER_LOAD_ARGS := $(if $(filter YES,$(HAS_BAKE_FILE)),$(DOCKER_LOAD_ARGS_BAKE),$(DOCKER_LOAD_ARGS_LEGACY))

# Git commit hash of the current HEAD, fallback to empty string
GIT_HASH?=$(shell git rev-parse HEAD 2>/dev/null || echo "")

# Git commit hash of the current HEAD, fallback to empty string
GIT_HASH_SHORT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "")

# Prioritize GitHub environment variable, fallback to local git branch
# Sanitizes to valid Docker tag characters and collapses consecutive invalid chars
GIT_BRANCH ?= $(shell \
	RAW_BRANCH=""; \
	if [ -n "$$GIT_BRANCH_GITHUB" ]; then \
		RAW_BRANCH="$$GIT_BRANCH_GITHUB"; \
	else \
		RAW_BRANCH=$$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo ""); \
	fi; \
	echo "$${RAW_BRANCH}" | sed -E 's/[^a-zA-Z0-9._-]+/-/g; s/^-+//; s/-+$$//' \
)

# Get and sanitize tags pointing to HEAD, handling multiple tags
GIT_TAGS ?= $(shell \
	TAGS=$$(git tag --points-at HEAD 2>/dev/null || echo ""); \
	echo "$$TAGS" | tr ' ' '\n' | sed -E 's/[^a-zA-Z0-9._-]+/-/g; s/^-+//; s/-+$$//' | paste -sd " " - \
)

# Appends "-dirty" if there are uncommitted changes
# Empty if working directory is clean or git not available
GIT_DIRTY?=$(shell git diff --quiet 2>/dev/null || echo "-dirty")

# Build timestamp in format YYYYMMDDHHMMSS
BUILD_DATE?=$(shell date +%Y%m%d%H%M%S)

# Directory for build information
INFO_TMP_DIR?=.build

# File for build information
INFO_TMP?=$(INFO_TMP_DIR)/.env

# First build list of image tags without repository prefix
IMAGE_TAGS=\
	$(BUILD_DATE)

# Add full Git hash tag with dirty flag if needed
ifneq ($(GIT_HASH),)
IMAGE_TAGS += $(GIT_HASH)$(GIT_DIRTY)
endif

# Add short Git hash tag with dirty flag if needed
ifneq ($(GIT_HASH_SHORT),)
IMAGE_TAGS += $(GIT_HASH_SHORT)$(GIT_DIRTY)
endif

# Add branch tag if exists
ifneq ($(GIT_BRANCH),)
IMAGE_TAGS += $(GIT_BRANCH)
endif

# Add Git tags if they exist
ifneq ($(GIT_TAGS),)
IMAGE_TAGS += $(GIT_TAGS)
endif

# Base image name without target prefix - generated from tags list
BASE_IMAGE_NAMES=$(foreach tag,$(IMAGE_TAGS),$(DOCKER_IMAGE_NAME):$(tag))

# Final image names - add target prefix if DOCKER_BUILD_TARGET is set
IMAGE_NAMES=$(if $(DOCKER_BUILD_TARGET),\
	$(foreach name,$(BASE_IMAGE_NAMES),$(subst :,:$(DOCKER_BUILD_TARGET)-,$(name))),\
	$(BASE_IMAGE_NAMES))

define docker_legacy
	@echo "🔨 Building Docker image..."
	@echo "🎯 Target platforms: $(PLATFORMS)"
	@if [ -n "$(DOCKER_BUILD_TARGET)" ]; then \
		echo "🎯 Building target stage: $(DOCKER_BUILD_TARGET)"; \
	fi
	@echo "🔧 Build args:"
	@echo "📦 Image tags:"
	@for tag in $(IMAGE_NAMES); do echo "   $$tag"; done
	@if [ -n "$(REMOTE_BUILD_KIT)" ]; then \
		echo "🌐 Using remote builder: $(REMOTE_BUILD_KIT_NAME)"; \
	else \
		echo "💻 Building locally"; \
	fi
	@echo "🚀 Starting build..."

	docker buildx build \
		--platform $(PLATFORMS) \
		$(if $(REMOTE_BUILD_KIT),--builder $(REMOTE_BUILD_KIT_NAME),) \
		--progress=$(DOCKER_BUILD_PROGRESS) \
		$(if $(DOCKER_BUILD_TARGET),--target $(DOCKER_BUILD_TARGET),) \
		$(foreach tag,$(IMAGE_NAMES),-t $(tag)) \
		$(1) .
endef

define docker_bake
	@echo "🔨 Building Docker image... (bake)"
	@echo "🎯 Target platforms: $(PLATFORMS)"
	@if [ -n "$(DOCKER_BUILD_TARGET)" ]; then \
		echo "🎯 Building target stage: $(DOCKER_BUILD_TARGET)"; \
	fi
	@echo "🔧 Build args:"
	@echo "📦 Image tags:"
	@for tag in $(IMAGE_NAMES); do echo "   $$tag"; done
	@if [ -n "$(REMOTE_BUILD_KIT)" ]; then \
		echo "🌐 Using remote builder: $(REMOTE_BUILD_KIT_NAME)"; \
	else \
		echo "💻 Building locally"; \
	fi
	@echo "🚀 Starting build..."

	GIT_BRANCH="$(GIT_BRANCH)" \
	EXTRA_TAGS="$(IMAGE_TAGS)" \
	docker buildx bake \
		$(if $(REMOTE_BUILD_KIT),--builder $(REMOTE_BUILD_KIT_NAME),) \
		--progress=$(DOCKER_BUILD_PROGRESS) \
		$(if $(DOCKER_BUILD_TARGET),--set "*.target=$(DOCKER_BUILD_TARGET)",) \
		$(1) \
		-f docker-bake.hcl -f .make/docker-common.hcl $(2)
endef

# Smart proxy function that automatically selects between legacy and bake
define docker_build
	@echo "HAS_BAKE_FILE: $(HAS_BAKE_FILE)"
	$(if $(filter YES,$(HAS_BAKE_FILE)),\
		$(info 📄 Using bake method with docker-bake.hcl...)\
		$(call docker_bake,$(1) $(2)),\
		$(info ℹ️ Using legacy build method...)\
		$(call docker_legacy,$(1)))
endef

# Helper for space substitution in tags
empty :=
space := $(empty) $(empty)
__IMAGE_NAMES = $(subst $(space),$(space),$(IMAGE_NAMES))

## Docker
docker-build: docker-gen-info docker-install ## Build multi-platform Docker images
	$(call docker_build)

# Build and load Docker images locally
docker-load: docker-gen-info docker-install ## Build and load Docker images locally
	$(call docker_build, $(DOCKER_LOAD_ARGS))

# Build and load Docker images locally
docker-print: docker-gen-info docker-install ## Build and load Docker images locally
	$(call docker_build, , --print)

# Build and publish Docker images to registry
docker-publish: docker-gen-info docker-install ## Build and publish to registry
	@echo "🔄 Using builder: $(REMOTE_BUILD_KIT_NAME)"
	@echo "🎯 Publishing to platforms: $(PLATFORMS)"
	@echo "📦 Publishing images:"
	@for tag in $(IMAGE_NAMES); do echo "   $$tag"; done
	@echo "🚀 Starting build and push..."
	$(call docker_build, $(DOCKER_PUSH_ARGS))

docker-targets:
	@echo "📑 Available build targets:"
	@grep "^FROM .* AS " Dockerfile | sed 's/FROM .* AS \(.*\)/  - \1/'
	@echo "  - <final stage> (default)"
	@echo
	@echo "You can set the DOCKER_BUILD_TARGET environment variable to build a specific target."

# Generate build information file
docker-gen-info: ## Generate build information file
	@mkdir -p $(INFO_TMP_DIR)
	@echo "DOCKER_IMAGE_NAME=$(DOCKER_IMAGE_NAME)" > $(INFO_TMP)
	@echo "GIT_HASH=$(GIT_HASH)" >> $(INFO_TMP)
	@echo "GIT_HASH_SHORT=$(GIT_HASH_SHORT)" >> $(INFO_TMP)
	@echo "GIT_BRANCH=$(GIT_BRANCH)" >> $(INFO_TMP)
	@echo "GIT_TAGS=$(GIT_TAGS)" >> $(INFO_TMP)
	@echo "GIT_DIRTY=$(GIT_DIRTY)" >> $(INFO_TMP)
	@echo "BUILD_DATE=$(BUILD_DATE)" >> $(INFO_TMP)

# Display build information and image tags
docker-info: docker-gen-info ## Show build info and tags
	@cat $(INFO_TMP) || echo "Error: Could not read $(INFO_TMP)"
	@echo
	@echo "🎯 Platforms"
	@echo "$(PLATFORMS)"
	@echo
	@echo "🔗 Images to publish"
	@for tag in $(IMAGE_NAMES); do echo "📦 $$tag"; done
	@rm -rf $(INFO_TMP_DIR)

# Install and configure Docker buildx remote builder
docker-install: docker-gen-info ## Configure buildx builder
ifdef REMOTE_BUILD_KIT
	@docker buildx rm $(REMOTE_BUILD_KIT_NAME) > /dev/null 2>&1 || true
	@docker buildx create \
		--driver remote \
		--name $(REMOTE_BUILD_KIT_NAME) \
		$(REMOTE_BUILD_KIT) > /dev/null 2>&1
endif

# Inspect Docker buildx builder configuration
docker-inspect: ## Inspect buildx configuration
	@docker buildx inspect $(REMOTE_BUILD_KIT_NAME)

docker-bash: ## (obsolete?) Start container with bash shell, requires $APP_NAME
	docker run --platform linux/amd64 -it zondax/${APP_NAME}:latest /bin/sh

.PHONY: *

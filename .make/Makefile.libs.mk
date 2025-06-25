# Shell colors
GREEN  := $(shell command -v tput >/dev/null 2>&1 && tput -Txterm setaf 2 || echo '')
YELLOW := $(shell command -v tput >/dev/null 2>&1 && tput -Txterm setaf 3 || echo '')
WHITE  := $(shell command -v tput >/dev/null 2>&1 && tput -Txterm setaf 7 || echo '')
CYAN   := $(shell command -v tput >/dev/null 2>&1 && tput -Txterm setaf 6 || echo '')
RESET  := $(shell command -v tput >/dev/null 2>&1 && tput -Txterm sgr0 || echo '')

## Misc

gitclean: ## Git Clean 
	git clean -xfd
	git submodule foreach --recursive git clean -xfd

zondax-update: ## (obsolete?) Update template using zondax cli
	@npx -y @zondax/cli@latest update
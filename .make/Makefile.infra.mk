KUBECONFORM_EXTRA_CRD_VERSION := main
KUBECONFORM_EXTRA_CRD_URL := https://raw.githubusercontent.com/datreeio/CRDs-catalog/$(KUBECONFORM_EXTRA_CRD_VERSION)

-include Makefile.yaml.mk

## Infra
infra-checks: ## Infra Checks
	@$(MAKE) yaml-lint
	@$(MAKE) infra-kubeconform

infra-kubeconform: ## Infra Kubeconform
	@echo "\nRunning kubeconform on Kubernetes manifests..."
	@kubeconform \
		-schema-location default \
		-schema-location '$(KUBECONFORM_EXTRA_CRD_URL)/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
		-ignore-missing-schemas \
		-n 16 \
		-ignore-filename-pattern '\.infra/.*/values\.yaml' \
		-ignore-filename-pattern '\.infra/.*/name-remapping\.yaml' \
		-summary .infra/

.PHONY: infra-checks infra-lint infra-kubeconform

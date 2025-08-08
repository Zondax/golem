## Reusable macro to ignore extra arguments for a given target with logging.
## Usage: $(call ignore_extra_args,<target_name>)
define ignore_extra_args
  $(if $(filter $1,$(MAKECMDGOALS)),\
    $(eval EXTRA_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS)))\
    $(eval override MAKECMDGOALS := $1)\
    $(if $(EXTRA_ARGS),\
      $(eval .PHONY: $(EXTRA_ARGS))\
      $(foreach t,$(EXTRA_ARGS),$(eval $(t): ; @true))\
    )\
  )
endef

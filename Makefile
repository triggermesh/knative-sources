BASE_DIR := $(CURDIR)
SUBDIRS  ?= slack
SOURCES  := $(shell awk 'BEGIN {FS = ":.*?";} /^[a-zA-Z0-9._-]+:.*?\#\# / {printf "%s ", $$1}' $(BASE_DIR)/scripts/inc.Source $(BASE_DIR)/scripts/inc.Codegen)

include $(BASE_DIR)/scripts/inc.Makefile

.PHONY: $(SUBDIRS) $(SOURCES) install-gotestsum install-golangci-lint

$(SOURCES): $(SUBDIRS)

$(SUBDIRS):
	@$(MAKE) -C $@ $(MAKECMDGOALS)

install-gotestsum:
ifndef HAS_GOTESTSUM
	curl -SL https://github.com/gotestyourself/gotestsum/releases/download/v0.4.2/gotestsum_0.4.2_linux_amd64.tar.gz | tar -C $(shell go env GOPATH)/bin -zxf -
endif

install-golangci-lint:
ifndef HAS_GOLANGCI_LINT
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.26.0
endif

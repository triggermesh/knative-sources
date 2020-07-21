COMMANDS           = adapter controller
SOURCES           ?= linux/amd64

BASE_DIR          ?= .

OUTPUT_DIR        ?= $(BASE_DIR)/_output

BIN_OUTPUT_DIR    ?= $(OUTPUT_DIR)
TEST_OUTPUT_DIR   ?= $(OUTPUT_DIR)
COVER_OUTPUT_DIR  ?= $(OUTPUT_DIR)
DIST_DIR          ?= $(OUTPUT_DIR)

RM                ?= rm
CP                ?= cp
MV                ?= mv
MKDIR             ?= mkdir

DOCKER            ?= docker
IMAGE_REPO        ?= gcr.io/triggermesh
IMAGE_NAME        ?= $(IMAGE_REPO)/$(KSOURCE)-source
IMAGE_TAG         ?= latest
IMAGE_SHA         ?= $(shell git rev-parse HEAD)

GO                ?= go
GOFMT             ?= gofmt
GOLINT            ?= golangci-lint run
GOTOOL            ?= go tool
GOTEST            ?= gotestsum --junitfile $(TEST_OUTPUT_DIR)/$(KSOURCE)-unit-tests.xml --format pkgname-and-test-fails --

LDFLAGS            =

HAS_GOTESTSUM     := $(shell command -v gotestsum;)
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)

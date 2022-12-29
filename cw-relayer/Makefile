BRANCH    := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DIR ?= $(CURDIR)/build
COMMIT    := $(shell git log -1 --format='%H')

all: test-unit install

.PHONY: all

###############################################################################
##                                  Version                                  ##
###############################################################################

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

###############################################################################
##                              Build / Install                              ##
###############################################################################

ldflags = -X github.com/ojo-network/contracts/cmd.Version=$(VERSION) \
		  -X github.com/ojo-network/contracts/cmd.Commit=$(COMMIT) \

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

build_tags += $(BUILD_TAGS)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

start:
	@make build
	build/cw-relayer config.toml

docker-build-relayer-e2e:
	@DOCKER_BUILDKIT=1 docker build -t cw-relayer -f tests/e2e/docker/relayer.Dockerfile .

.PHONY: build install start docker-build-api-e2e

###############################################################################
##                              Tests & Linting                              ##
###############################################################################
PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e')
PACKAGES_E2E=$(shell go list ./... | grep '/e2e')

test-unit:
	@echo "--> Running tests"
	@go test -mod=readonly -race $(PACKAGES_UNIT) -v

test-e2e:
	$(MAKE) docker-build-relayer-e2e
	@go test -mod=readonly -race $(PACKAGES_E2E) -v


.PHONY: test-unit test-e2e test-unit

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix --timeout=8m

.PHONY: lint
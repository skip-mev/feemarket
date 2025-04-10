#!/usr/bin/make -f

export VERSION := $(shell echo $(shell git describe --always --match "v*") | sed 's/^v//')
export COMMIT := $(shell git log -1 --format='%H')
export COMETBFT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')

BIN_DIR ?= $(GOPATH)/bin
BUILD_DIR ?= $(CURDIR)/build
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)
HTTPS_GIT := https://github.com/skip-mev/feemarket.git
DOCKER := $(shell which docker)
COVER_FILE := coverage.txt
COVER_HTML_FILE := cover.html

###############################################################################
##                                Workspaces                                 ##
###############################################################################


tidy:
	@go mod tidy

###############################################################################
###                                Test App                                 ###
###############################################################################

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=feemarket \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=feemarketd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(COMETBFT_VERSION)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

BUILD_TARGETS := build-test-app

build-test-app: BUILD_ARGS=-o $(BUILD_DIR)/

$(BUILD_TARGETS): $(BUILD_DIR)/
	cd $(CURDIR)/tests/app && go build -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILD_DIR)/:
	mkdir -p $(BUILD_DIR)/

# build-and-start-app builds a fee market simulation application binary in the build folder
# and initializes a single validator configuration. If desired, users can suppliment
# other addresses using "genesis add-genesis-account address 10000000000000000000000000stake".
# This will allow users to bootstrap their wallet with a balance.
build-and-start-app: build-test-app
	./build/feemarketd init validator1 --chain-id chain-id-0
	./build/feemarketd keys add validator1
	./build/feemarketd genesis add-genesis-account validator1 10000000000000000000000000stake
	./build/feemarketd genesis add-genesis-account cosmos1see0htr47uapjvcvh0hu6385rp8lw3em24hysg 10000000000000000000000000stake
	./build/feemarketd genesis gentx validator1 1000000000stake --chain-id chain-id-0
	./build/feemarketd genesis collect-gentxs
	./build/feemarketd start --api.enable true --api.enabled-unsafe-cors true --log_level info

.PHONY: build-test-app build-and-start-app

.PHONY: docker-build docker-build-e2e
###############################################################################
##                                  Docker                                   ##
###############################################################################

docker-build:
	@echo "Building E2E Docker image..."
	@DOCKER_BUILDKIT=1 docker build -t skip-mev/feemarket-e2e -f contrib/images/feemarket.e2e.Dockerfile .

docker-build-e2e:
	@echo "Building e2e-test Docker image..."
	@DOCKER_BUILDKIT=1 docker build -t feemarket-e2e -f contrib/images/feemarket.e2e.Dockerfile .

###############################################################################
###                                  Tests                                  ###
###############################################################################

TEST_E2E_DEPS = docker-build-e2e
TEST_E2E_TAGS = e2e

test-e2e: $(TEST_E2E_DEPS)
	@echo "Running e2e tests..."
	cd tests/e2e && go test . -timeout 30m -p 1 -race -v -tags='$(TEST_E2E_TAGS)'

test-unit:
	@go test -v -race $(shell go list ./... | grep -v tests/)

test-fuzz:
	@cd ./x/feemarket/fuzz && go test -v -race -rapid.checks=100_000 -p 4

test-integration:
	@go test -v -race ./tests/integration

test-cover:
	@echo Running unit tests and creating coverage report...
	@go test -mod=readonly -v -timeout 30m -coverprofile=$(COVER_FILE) -covermode=atomic $(shell go list ./... | grep -v tests/ | grep -v api/ | grep -v testutils/)
	@sed -i '/.pb.go/d' $(COVER_FILE)
	@sed -i '/.gw.go/d' $(COVER_FILE)
	@go tool cover -html=$(COVER_FILE) -o $(COVER_HTML_FILE)
	@rm $(COVER_FILE)

test-all: test-unit test-integration test-e2e test-fuzz

.PHONY: test-unit test-e2e test-integration test-cover test-all

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-pulsar-gen:
	@echo "Generating Dep-Inj Protobuf files"
	@$(protoImage) sh ./scripts/protocgen-pulsar.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking proto-update-deps

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run

lint-fix:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --fix --issues-exit-code=0

lint-markdown:
	@echo "--> Running markdown linter"
	@markdownlint **/*.md

.PHONY: lint lint-fix lint-markdown

###############################################################################
###                                Formatting                               ###
###############################################################################

format:
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run github.com/client9/misspell/cmd/misspell -w
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -name '*.pb.go' -not -name '*.pulsar.go' -not -name '*.gw.go' | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/skip-mev/feemarket

mocks:
	@echo "--> generating mocks"
	@go install github.com/vektra/mockery/v2
	@go generate ./...
	make format

.PHONY: format mocks

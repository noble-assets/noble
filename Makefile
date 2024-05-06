BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  ifeq (,$(VERSION))
    ifeq ($(shell git status --porcelain),)
    	VERSION := $(BRANCH)
    else
    	VERSION := $(BRANCH):dirty
    endif
  endif
endif

ldflags := $(LDFLAGS)
ldflags += -X github.com/cosmos/cosmos-sdk/version.Name=noble \
	-X github.com/cosmos/cosmos-sdk/version.AppName=nobled \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags 'netgo ledger' -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/nobled

build:
	@go build -mod=readonly $(BUILD_FLAGS) -o $(PWD)/bin/ ./cmd/nobled

###############################################################################
###                          Formatting & Linting                           ###
###############################################################################

gofumpt_cmd=mvdan.cc/gofumpt
golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint

format:
	@echo "🤖 Running formatter..."
	@go run $(gofumpt_cmd) -l -w .
	@echo "✅ Completed formatting!"

lint:
	@echo "🤖 Running linter..."
	@go run $(golangci_lint_cmd) run --timeout=10m
	@echo "✅ Completed linting!"

###############################################################################
###                                 Testing                                 ###
###############################################################################

# TODO(@john): Add "local-image" command!

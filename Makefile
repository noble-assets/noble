BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
VERSION := $(shell echo $(shell git describe --tags --always --dirty --match "v*") | sed 's/^v//')
LEDGER_ENABLED ?= true

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
	GCCEXE = $(shell where gcc.exe 2> NUL)
	ifeq ($(GCCEXE),)
		$(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
	else
		build_tags += ledger
	endif
	else
	UNAME_S = $(shell uname -s)
	ifeq ($(UNAME_S),OpenBSD)
		$(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
	else
		GCC = $(shell command -v gcc 2> /dev/null)
		ifeq ($(GCC),)
			$(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
		else
			build_tags += ledger
		endif
	endif
	endif
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Noble \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=nobled \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

build:
	@echo "🤖 Building nobled..."
	@go build -mod=readonly $(BUILD_FLAGS) -o "$(PWD)/build/nobled" ./cmd/nobled
	@echo "✅ Completed build!"

install:
	@echo "🤖 Installing nobled..."
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/nobled
	@echo "✅ Completed install!"

###############################################################################
###                                 Tooling                                 ###
###############################################################################

gofumpt_cmd=mvdan.cc/gofumpt
golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint
BUILDER_VERSION=0.15.3

FILES := $(shell find . -name "*.go")
license:
	@go-license --config .github/license.yml $(FILES)

format:
	@echo "🤖 Running formatter..."
	@go run $(gofumpt_cmd) -l -w .
	@echo "✅ Completed formatting!"

lint:
	@echo "🤖 Running linter..."
	@go run $(golangci_lint_cmd) run --timeout=10m
	@echo "✅ Completed linting!"

swagger:
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./api/generate.sh

###############################################################################
###                                 Testing                                 ###
###############################################################################

local-image:
ifeq (,$(shell which heighliner))
	@echo heighliner not found. https://github.com/strangelove-ventures/heighliner
else
	@echo "🤖 Building image..."
	@heighliner build --chain noble --local
	@echo "✅ Completed build!"
endif

.PHONY: license format lint build install local-image

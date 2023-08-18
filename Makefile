#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

CHAIN_NAME = noble
DAEMON_NAME = nobled

LEDGER_ENABLED ?= true
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in "github.com/cometbft/cometbft v0.37.2"
BUILDDIR ?= $(CURDIR)/build

export GO111MODULE = on

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

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(CHAIN_NAME) \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=$(DAEMON_NAME) \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TM_VERSION)

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

build:
	@echo "🤖 Building nobled..."
	@go build $(BUILD_FLAGS) -o "$(PWD)/build/" ./cmd/nobled && cd ..
	@echo "✅ Completed build!"

install:
	@echo "🤖 Installing nobled..."
	@go install $(BUILD_FLAGS) ./cmd/nobled && cd ..
	@echo "✅ Completed install!"

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
###                          INTERCHAINTEST (ictest)                        ###
###############################################################################

test:
	go test -v -race ./...

ictest-tkn-factory:
	cd interchaintest && go test -race -v -run ^TestNobleChain$$ .

ictest-packet-forward:
	cd interchaintest && go test -race -v -run ^TestPacketForwardMiddleware$$ .

ictest-paramauthority:
	cd interchaintest && go test -race -v -run ^TestNobleParamAuthority$$ .

ictest-chain-upgrade-grand-1:
	cd interchaintest && go test -race -v -timeout 15m -run ^TestGrand1ChainUpgrade$$ .

ictest-chain-upgrade-noble-1:
	cd interchaintest && go test -race -v -timeout 15m -run ^TestNoble1ChainUpgrade$$ .

ictest-globalFee:
	cd interchaintest && go test -race -v -run ^TestGlobalFee$$ .

ictest-ics20-bps-fees:
	cd interchaintest && go test -race -v -run ^TestICS20BPSFees$$ .

ictest-client-substitution:
	cd interchaintest && go test -race -v -run ^TestClientSubstitution$$ .

###############################################################################
###                                Build Image                              ###
###############################################################################
get-heighliner:
	git clone https://github.com/strangelove-ventures/heighliner.git
	cd heighliner && go install

local-image:
ifeq (,$(shell which heighliner))
	echo 'heighliner' binary not found. Consider running `make get-heighliner`
else
	heighliner build -c noble --local -f ./chains.yaml
endif

.PHONY: all build-linux install lint test \
	go-mod-cache build interchaintest get-heighliner local-image \

###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.26.1
BUILDER_VERSION=0.13.5

proto-all: proto-format proto-lint proto-gen

proto-format:
	@echo "🤖 Running protobuf formatter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "✅ Completed protobuf formatting!"

proto-gen:
	@echo "🤖 Generating code from protobuf..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./proto/generate.sh
	@echo "✅ Completed code generation!"

proto-lint:
	@echo "🤖 Running protobuf linter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "✅ Completed protobuf linting!"

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

ldflags := $(LDFLAGS)
ldflags += -X github.com/cosmos/cosmos-sdk/version.Name=Noble \
	-X github.com/cosmos/cosmos-sdk/version.AppName=nobled \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'


###############################################################################
###                              Building / Install                         ###
###############################################################################

install: go.sum
	@echo "🤖 Installing nobled..."
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/nobled
	@echo "✅ Completed install!"

build:
	@echo "🤖 Building nobled..."
	@go build $(BUILD_FLAGS) -o "$(PWD)/build/" ./...
	@echo "✅ Completed build!"

###############################################################################
###                                 Testing                                 ###
###############################################################################

local-image:
ifeq (,$(shell which heighliner))
	echo 'heighliner' binary not found. Please install: https://github.com/strangelove-ventures/heighliner
else
	heighliner build -c noble --local
endif


###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m 


.PHONY: install build local-image lint
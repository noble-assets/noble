BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  ifeq (,$(VERSION))
    ifeq ($(shell git status --porcelain),)
    	VERSION := $(BRANCH)
    else
    	VERSION := $(BRANCH)-dirty
    endif
  endif
endif

ldflags := $(LDFLAGS)
ldflags += -X github.com/cosmos/cosmos-sdk/version.Name=Noble \
	-X github.com/cosmos/cosmos-sdk/version.AppName=nobled \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)
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
	@go build -mod=readonly $(BUILD_FLAGS) -o "$(PWD)/build/" ./...
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
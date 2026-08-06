BINARY  := kramgo
MODULE  := github.com/crystalpoplar/kramgo
VERSION ?= dev
LDFLAGS := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION)"

# Supported OS/arch combinations for cross-platform builds.
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: all build test lint clean cross

all: build

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/ dist/

cross:
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		$(eval EXT  := $(if $(filter windows,$(OS)),.exe,)) \
		GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) \
			-o dist/$(BINARY)-$(OS)-$(ARCH)$(EXT) \
			./cmd/$(BINARY) ; \
	)
	@echo "Cross-platform binaries are in dist/"

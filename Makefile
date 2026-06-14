BINARY        := paterna
PKG_VERSION   := github.com/hugaojanuario/Paterna/internal/version
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE          ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(PKG_VERSION).Version=$(VERSION) \
	-X $(PKG_VERSION).Commit=$(COMMIT) \
	-X $(PKG_VERSION).Date=$(DATE)

.PHONY: build install run tidy clean release-dry

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/cli

install: build
	install -m 0755 $(BINARY) $${HOME}/.local/bin/$(BINARY)

run:
	go run ./cmd/cli

tidy:
	go mod tidy

clean:
	rm -rf $(BINARY) dist/

release-dry:
	goreleaser release --snapshot --clean --skip=publish

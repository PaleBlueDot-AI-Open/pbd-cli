.PHONY: build build-dev install clean

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "-X github.com/PaleBlueDot-AI-Open/pbd-cli/internal/config.env=prod \
                       -X main.version=$(VERSION) \
                       -X main.commit=$(COMMIT) \
                       -X main.date=$(DATE)"

LDFLAGS_DEV := -ldflags "-X github.com/PaleBlueDot-AI-Open/pbd-cli/internal/config.env=dev \
                          -X main.version=$(VERSION)-dev \
                          -X main.commit=$(COMMIT) \
                          -X main.date=$(DATE)"

build:
	go build $(LDFLAGS) -o pbd-cli

build-dev:
	go build $(LDFLAGS_DEV) -o pbd-cli-dev

install: build
	sudo mv pbd-cli /usr/local/bin/

clean:
	rm -f pbd-cli pbd-cli-dev

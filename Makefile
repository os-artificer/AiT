.PHONY: all build install test vet fmt clean test-cross dupl

BINARY_NAME := ait
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/ait
LDFLAGS_PKG := github.com/os-artificer/AiT/internal/ait/version

VERSION    ?= $(shell cat internal/ait/version/VERSION 2>/dev/null | tr -d ' \n')
GIT_COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X $(LDFLAGS_PKG).version=$(VERSION) \
           -X $(LDFLAGS_PKG).gitCommit=$(GIT_COMMIT) \
           -X $(LDFLAGS_PKG).buildDate=$(BUILD_DATE)

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

install:
	go install -ldflags "$(LDFLAGS)" $(MAIN_PKG)

test:
	go test ./...

test-cross:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" ./...
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

dupl:
	go run github.com/mibk/dupl@latest -threshold 15 $$(find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*')

clean:
	rm -rf $(BUILD_DIR)

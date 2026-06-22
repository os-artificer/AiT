.PHONY: all build build-cross install test vet fmt clean test-cross test-e2e test-e2e-smoke dupl check-generate

TOOLS       := ai-ssh
BINARY_NAME := ai-ssh
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/ai-ssh
LDFLAGS_PKG := github.com/os-artificer/AiT/internal/ai-ssh/version

VERSION    ?= $(shell cat internal/ai-ssh/version/VERSION 2>/dev/null | tr -d ' \n')
GIT_COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X $(LDFLAGS_PKG).version=$(VERSION) \
           -X $(LDFLAGS_PKG).gitCommit=$(GIT_COMMIT) \
           -X $(LDFLAGS_PKG).buildDate=$(BUILD_DATE)

CROSS_PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

build-cross:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for platform in $(CROSS_PLATFORMS); do \
		os=$${platform%%/*}; arch=$${platform#*/}; \
		ext=; [ "$$os" = windows ] && ext=.exe; \
		out=$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext; \
		echo ">> $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags "$(LDFLAGS)" -o $$out $(MAIN_PKG); \
	done

install:
	go install -ldflags "$(LDFLAGS)" $(MAIN_PKG)

test:
	go test ./...

test-cross:
	@set -e; \
	for platform in $(CROSS_PLATFORMS); do \
		os=$${platform%%/*}; arch=$${platform#*/}; \
		echo ">> test-cross $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags "$(LDFLAGS)" ./...; \
	done

check-generate:
	go generate ./internal/ai-ssh/registry/...
	git diff --exit-code -- internal/ai-ssh/registry/all/imports.go

test-e2e:
	@test -x bin/ai-ssh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m "not destructive"

test-e2e-smoke:
	@test -x bin/ai-ssh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m smoke

vet:
	go vet ./...

fmt:
	go fmt ./...

dupl:
	go run github.com/mibk/dupl@latest -threshold 15 $$(find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*')

clean:
	rm -rf $(BUILD_DIR) .gocache .tmp .test-fresh-gomodcache

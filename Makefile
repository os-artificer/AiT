.PHONY: all build build-cross install test vet fmt clean test-cross test-e2e test-e2e-smoke dupl check-generate

TOOLS     := aissh aiimg ailog aihost ait
BUILD_DIR := bin

MODULE := github.com/os-artificer/AiT

GIT_COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

CROSS_PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

PLUGIN_TOOLS := aissh aiimg ailog aihost

tool_version = $(shell cat $(1) 2>/dev/null | tr -d ' \n')
tool_ldflags = -X $(2).version=$(call tool_version,$(1)) -X $(2).gitCommit=$(GIT_COMMIT) -X $(2).buildDate=$(BUILD_DATE)

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aissh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/aissh ;; \
			aiimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/aiimg ;; \
			ailog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/ailog ;; \
			aihost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/aihost ;; \
			ait) verfile=internal/ait/version/VERSION; ldpkg=$(MODULE)/internal/ait/version; main=./cmd/ait ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		echo ">> $(BUILD_DIR)/$$tool"; \
		go build -ldflags "$$ldflags" -o $(BUILD_DIR)/$$tool $$main; \
	done; \
	for b in $(PLUGIN_TOOLS); do \
		mkdir -p $(BUILD_DIR)/plugins/$$b; \
		ln -sfn ../../$$b $(BUILD_DIR)/plugins/$$b/$$b 2>/dev/null || \
			cp -f $(BUILD_DIR)/$$b $(BUILD_DIR)/plugins/$$b/$$b 2>/dev/null || \
			cp -f $(BUILD_DIR)/$$b.exe $(BUILD_DIR)/plugins/$$b/$$b.exe; \
		cp internal/pkg/aitplugin/manifests/$$b/plugin.json $(BUILD_DIR)/plugins/$$b/plugin.json; \
	done

build-cross:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aissh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/aissh ;; \
			aiimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/aiimg ;; \
			ailog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/ailog ;; \
			aihost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/aihost ;; \
			ait) verfile=internal/ait/version/VERSION; ldpkg=$(MODULE)/internal/ait/version; main=./cmd/ait ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		for platform in $(CROSS_PLATFORMS); do \
			os=$${platform%%/*}; arch=$${platform#*/}; \
			ext=; [ "$$os" = windows ] && ext=.exe; \
			out=$(BUILD_DIR)/$$tool-$$os-$$arch$$ext; \
			echo ">> $$out"; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build -ldflags "$$ldflags" -o $$out $$main; \
		done; \
	done

install:
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aissh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/aissh ;; \
			aiimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/aiimg ;; \
			ailog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/ailog ;; \
			aihost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/aihost ;; \
			ait) verfile=internal/ait/version/VERSION; ldpkg=$(MODULE)/internal/ait/version; main=./cmd/ait ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		go install -ldflags "$$ldflags" $$main; \
	done; \
	gobin=$$(go env GOBIN); \
	if [ -z "$$gobin" ]; then gobin=$$(go env GOPATH)/bin; fi; \
	for b in $(PLUGIN_TOOLS); do \
		mkdir -p $$gobin/plugins/$$b; \
		ln -sfn ../../$$b $$gobin/plugins/$$b/$$b 2>/dev/null || \
			cp -f $$gobin/$$b $$gobin/plugins/$$b/$$b 2>/dev/null || \
			cp -f $$gobin/$$b.exe $$gobin/plugins/$$b/$$b.exe; \
		cp internal/pkg/aitplugin/manifests/$$b/plugin.json $$gobin/plugins/$$b/plugin.json; \
	done

test:
	go test ./...

test-cross:
	@set -e; \
	for platform in $(CROSS_PLATFORMS); do \
		os=$${platform%%/*}; arch=$${platform#*/}; \
		echo ">> test-cross $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build ./...; \
	done

check-generate:
	go generate ./internal/ssh/registry/...
	git diff --exit-code -- internal/ssh/registry/all/imports.go

# Full E2E including real SSH hosts from tests/config.local.yaml (local only; skipped on CI).
test-e2e:
	@test -x $(BUILD_DIR)/aissh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m "not destructive"

# CI-safe smoke: help / config / session CLI / experimental-AI gate. Never dials real hosts.
test-e2e-smoke:
	@test -x $(BUILD_DIR)/aissh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m smoke

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

dupl:
	@command -v dupl >/dev/null || go install github.com/golangci/dupl@latest
	dupl -t 100 ./internal/...

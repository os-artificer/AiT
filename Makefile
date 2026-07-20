.PHONY: all build build-cross install test vet fmt clean test-cross test-e2e test-e2e-smoke dupl check-generate

TOOLS     := aitssh aitimg
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

# tool_version_file,tool_ldflags_pkg,tool_main_pkg,tool_version
define TOOL_META_aitssh
internal/aissh/version/VERSION \
github.com/os-artificer/AiT/internal/aissh/version \
./cmd/aitssh
endef

define TOOL_META_aitimg
internal/aitimg/version/VERSION \
github.com/os-artificer/AiT/internal/aitimg/version \
./cmd/aitimg
endef

tool_version = $(shell cat $(1) 2>/dev/null | tr -d ' \n')
tool_ldflags = -X $(2).version=$(call tool_version,$(1)) -X $(2).gitCommit=$(GIT_COMMIT) -X $(2).buildDate=$(BUILD_DATE)

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aitssh) verfile=internal/aissh/version/VERSION; ldpkg=$(MODULE)/internal/aissh/version; main=./cmd/aitssh ;; \
			aitimg) verfile=internal/aitimg/version/VERSION; ldpkg=$(MODULE)/internal/aitimg/version; main=./cmd/aitimg ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		echo ">> $(BUILD_DIR)/$$tool"; \
		go build -ldflags "$$ldflags" -o $(BUILD_DIR)/$$tool $$main; \
	done

build-cross:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aitssh) verfile=internal/aissh/version/VERSION; ldpkg=$(MODULE)/internal/aissh/version; main=./cmd/aitssh ;; \
			aitimg) verfile=internal/aitimg/version/VERSION; ldpkg=$(MODULE)/internal/aitimg/version; main=./cmd/aitimg ;; \
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
			aitssh) verfile=internal/aissh/version/VERSION; ldpkg=$(MODULE)/internal/aissh/version; main=./cmd/aitssh ;; \
			aitimg) verfile=internal/aitimg/version/VERSION; ldpkg=$(MODULE)/internal/aitimg/version; main=./cmd/aitimg ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		go install -ldflags "$$ldflags" $$main; \
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
	go generate ./internal/aissh/registry/...
	git diff --exit-code -- internal/aissh/registry/all/imports.go

# Full E2E including real SSH hosts from tests/config.local.yaml (local only; skipped on CI).
test-e2e:
	@test -x $(BUILD_DIR)/aitssh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m "not destructive"

# CI-safe smoke: help / config / session CLI / experimental-AI gate. Never dials real hosts.
test-e2e-smoke:
	@test -x $(BUILD_DIR)/aitssh || $(MAKE) build
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

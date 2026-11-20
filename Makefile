.PHONY: all build build-cross install install-git-hooks release-archives test vet fmt clean test-cross test-e2e test-e2e-smoke test-e2e-mock dupl check-generate nestscan

TOOLS     := aitssh aitimg aitlog aithost aitkeep aitvid ait
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

PLUGIN_TOOLS := aitssh aitimg aitlog aithost aitkeep aitvid

tool_version = $(shell cat $(1) 2>/dev/null | tr -d ' \n')
tool_ldflags = -X $(2).version=$(call tool_version,$(1)) -X $(2).gitCommit=$(GIT_COMMIT) -X $(2).buildDate=$(BUILD_DATE)

all: build

build:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aitssh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/ssh ;; \
			aitimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/img ;; \
			aitlog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/log ;; \
			aithost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/host ;; \
			aitkeep) verfile=internal/pkg/keepalive/VERSION; ldpkg=$(MODULE)/internal/pkg/keepalive; main=./cmd/keep ;; \
			aitvid) verfile=internal/vid/version/VERSION; ldpkg=$(MODULE)/internal/vid/version; main=./cmd/vid ;; \
			ait) verfile=internal/ait/version/VERSION; ldpkg=$(MODULE)/internal/ait/version; main=./cmd/ait ;; \
			*) echo "unknown tool $$tool"; exit 1 ;; \
		esac; \
		ver=$$(cat $$verfile 2>/dev/null | tr -d ' \n'); \
		ldflags="-X $$ldpkg.version=$$ver -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
		echo ">> $(BUILD_DIR)/$$tool"; \
		go build -ldflags "$$ldflags" -o $(BUILD_DIR)/$$tool $$main; \
	done; \
	for b in $(PLUGIN_TOOLS); do \
		cmd=$$(sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' internal/pkg/aitplugin/manifests/$$b/plugin.json | head -1); \
		mkdir -p $(BUILD_DIR)/plugins/$$cmd; \
		ln -sfn ../../$$b $(BUILD_DIR)/plugins/$$cmd/$$b 2>/dev/null || \
			cp -f $(BUILD_DIR)/$$b $(BUILD_DIR)/plugins/$$cmd/$$b 2>/dev/null || \
			cp -f $(BUILD_DIR)/$$b.exe $(BUILD_DIR)/plugins/$$cmd/$$b.exe; \
		cp internal/pkg/aitplugin/manifests/$$b/plugin.json $(BUILD_DIR)/plugins/$$cmd/plugin.json; \
	done

build-cross:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aitssh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/ssh ;; \
			aitimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/img ;; \
			aitlog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/log ;; \
			aithost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/host ;; \
			aitkeep) verfile=internal/pkg/keepalive/VERSION; ldpkg=$(MODULE)/internal/pkg/keepalive; main=./cmd/keep ;; \
			aitvid) verfile=internal/vid/version/VERSION; ldpkg=$(MODULE)/internal/vid/version; main=./cmd/vid ;; \
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

# Release archives for GitHub (requires TAG=vX.Y.Z). Output: dist/ait-<tag>-<goos>-<arch>.tar.gz or .zip
release-archives:
ifndef TAG
	$(error TAG is required, e.g. make release-archives TAG=v0.1.0)
endif
	@set -e; \
	mkdir -p dist; \
	for platform in $(CROSS_PLATFORMS); do \
		os=$${platform%%/*}; arch=$${platform#*/}; \
		staging=dist/staging-$$os-$$arch; \
		rm -rf $$staging; \
		mkdir -p $$staging; \
		ext=; [ "$$os" = windows ] && ext=.exe; \
		for tool in $(TOOLS); do \
			case $$tool in \
				aitssh) ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/ssh ;; \
				aitimg) ldpkg=$(MODULE)/internal/img/version; main=./cmd/img ;; \
				aitlog) ldpkg=$(MODULE)/internal/log/version; main=./cmd/log ;; \
				aithost) ldpkg=$(MODULE)/internal/host/version; main=./cmd/host ;; \
				aitkeep) ldpkg=$(MODULE)/internal/pkg/keepalive; main=./cmd/keep ;; \
				aitvid) ldpkg=$(MODULE)/internal/vid/version; main=./cmd/vid ;; \
				ait) ldpkg=$(MODULE)/internal/ait/version; main=./cmd/ait ;; \
				*) echo "unknown tool $$tool"; exit 1 ;; \
			esac; \
			ldflags="-X $$ldpkg.version=$(TAG) -X $$ldpkg.gitCommit=$(GIT_COMMIT) -X $$ldpkg.buildDate=$(BUILD_DATE)"; \
			echo ">> release $$tool $$os/$$arch"; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build -ldflags "$$ldflags" -o $$staging/$$tool$$ext $$main; \
		done; \
		for b in $(PLUGIN_TOOLS); do \
			cmd=$$(sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' internal/pkg/aitplugin/manifests/$$b/plugin.json | head -1); \
			mkdir -p $$staging/plugins/$$cmd; \
			cp internal/pkg/aitplugin/manifests/$$b/plugin.json $$staging/plugins/$$cmd/plugin.json; \
			cp $$staging/$$b$$ext $$staging/plugins/$$cmd/$$b$$ext; \
		done; \
		if [ "$$os" = windows ]; then \
			( cd $$staging && zip -qr ../ait-$(TAG)-windows-amd64.zip . ); \
		else \
			COPYFILE_DISABLE=1 tar -czf dist/ait-$(TAG)-$$os-$$arch.tar.gz -C $$staging .; \
		fi; \
		rm -rf $$staging; \
	done

install:
	@set -e; \
	for tool in $(TOOLS); do \
		case $$tool in \
			aitssh) verfile=internal/ssh/version/VERSION; ldpkg=$(MODULE)/internal/ssh/version; main=./cmd/ssh ;; \
			aitimg) verfile=internal/img/version/VERSION; ldpkg=$(MODULE)/internal/img/version; main=./cmd/img ;; \
			aitlog) verfile=internal/log/version/VERSION; ldpkg=$(MODULE)/internal/log/version; main=./cmd/log ;; \
			aithost) verfile=internal/host/version/VERSION; ldpkg=$(MODULE)/internal/host/version; main=./cmd/host ;; \
			aitkeep) verfile=internal/pkg/keepalive/VERSION; ldpkg=$(MODULE)/internal/pkg/keepalive; main=./cmd/keep ;; \
			aitvid) verfile=internal/vid/version/VERSION; ldpkg=$(MODULE)/internal/vid/version; main=./cmd/vid ;; \
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
		cmd=$$(sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' internal/pkg/aitplugin/manifests/$$b/plugin.json | head -1); \
		mkdir -p $$gobin/plugins/$$cmd; \
		ln -sfn ../../$$b $$gobin/plugins/$$cmd/$$b 2>/dev/null || \
			cp -f $$gobin/$$b $$gobin/plugins/$$cmd/$$b 2>/dev/null || \
			cp -f $$gobin/$$b.exe $$gobin/plugins/$$cmd/$$b.exe; \
		cp internal/pkg/aitplugin/manifests/$$b/plugin.json $$gobin/plugins/$$cmd/plugin.json; \
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
	@test -x $(BUILD_DIR)/aitssh || $(MAKE) build
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m "not destructive"

# CI-safe smoke: help / config / session CLI / experimental-AI gate / in-repo testenv-mock.
# Never dials real hosts; SSH/HTTP scenarios use bin/testenv-mock (see docs/architecture/testenv-mock.md).
test-e2e-smoke:
	@test -x $(BUILD_DIR)/aitssh || $(MAKE) build
	go build -o $(BUILD_DIR)/testenv-mock ./mock/testenv-mock
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && ../tests/.venv/bin/pytest e2e -v -m smoke

# GitHub CI full e2e: start testenv-mock, inject ssh.primary, run @ssh against the mock.
# Skips tmux / zmodem / destructive; nested/AI keys still skip when unconfigured (MS20).
test-e2e-mock:
	@test -x $(BUILD_DIR)/aitssh || $(MAKE) build
	go build -o $(BUILD_DIR)/testenv-mock ./mock/testenv-mock
	@python3 -m venv tests/.venv 2>/dev/null || true
	tests/.venv/bin/pip install -r tests/requirements.txt -q
	cd tests && AIT_TEST_USE_MOCK=1 ../tests/.venv/bin/pytest e2e -v -m "not destructive and not tmux and not zmodem"

fmt:
	go fmt ./...

# Strip Cursor Co-authored-by / Made-with trailers (IDE may inject despite attribution: false).
install-git-hooks:
	chmod +x scripts/git-hooks/prepare-commit-msg scripts/git-hooks/commit-msg scripts/git-hooks/post-commit
	chmod +x scripts/git-hooks/strip-cursor-trailers-lib.sh
	git config core.hooksPath scripts/git-hooks
	@echo "git hooks installed (core.hooksPath=scripts/git-hooks)"
	@echo "Also disable Cursor IDE: Settings → Agent → Attribution (commits/PRs off)."

vet:
	go vet ./...

# Fail if production Go has control-flow nesting depth >= 3 (go-structure §9).
# Not wired into CI yet; run before merge: make nestscan
nestscan:
	go run ./scripts/nestscan.go ./internal ./cmd ./pkg

clean:
	rm -rf $(BUILD_DIR)

dupl:
	@command -v dupl >/dev/null || go install github.com/golangci/dupl@latest
	dupl -t 100 ./internal

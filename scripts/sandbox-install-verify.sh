#!/usr/bin/env bash
# Local sandbox setup and verification for ait install/update/upgrade (M1–M10).
# Usage:
#   Terminal 1: go run ./cmd/aitrelease-mock -tag v0.9.0-sandbox
#   Terminal 2: ./scripts/sandbox-install-verify.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SANDBOX="${SANDBOX:-/tmp/ait-sandbox}"
TAG="${TAG:-v0.9.0-sandbox}"
MOCK_URL="${AIT_RELEASE_TEST_BASE_URL:-http://127.0.0.1:8765}"

run_ait() {
	local input="$1"
	AIT_HOME="$SANDBOX/home" AIT_RELEASE_TEST_BASE_URL="$MOCK_URL" \
		"$SANDBOX/bin/ait" <<< "$input"
}

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

pass() {
	echo "PASS: $*"
}

setup_sandbox() {
	echo ">> setup sandbox at $SANDBOX"
	rm -rf "$SANDBOX"
	mkdir -p "$SANDBOX/bin" "$SANDBOX/home"
	( cd "$ROOT" && make build >/dev/null )
	cp "$ROOT/bin/ait" "$SANDBOX/bin/ait"
	chmod +x "$SANDBOX/bin/ait"
	touch "$SANDBOX/bin/.ait-install-probe" && rm "$SANDBOX/bin/.ait-install-probe"
}

check_mock() {
	if ! curl -sf "$MOCK_URL/repos/os-artificer/AiT/releases/latest" >/dev/null; then
		fail "mock not reachable at $MOCK_URL — start: go run ./cmd/aitrelease-mock -tag $TAG"
	fi
}

verify_m1() {
	local out
	out="$(run_ait $'help install\nexit\n')"
	echo "$out" | grep -q "Usage:" || fail "M1 missing Usage"
	echo "$out" | grep -q "install" || fail "M1 missing install"
	pass "M1 help install"
}

verify_m2() {
	local out
	out="$(run_ait $'install aissh\nexit\n')"
	echo "$out" | grep -q "fetching ait-" || fail "M2 missing fetching line"
	echo "$out" | grep -q "installed aissh -> $TAG" || fail "M2 missing installed line: $out"
	test -f "$SANDBOX/bin/plugins/aissh/plugin.json" || fail "M2 missing plugin.json"
	test -x "$SANDBOX/bin/plugins/aissh/aissh" || test -f "$SANDBOX/bin/plugins/aissh/aissh" || fail "M2 missing aissh plugin binary"
	pass "M2 install aissh"
}

verify_m3() {
	local out
	out="$(run_ait $'help\nexit\n')"
	echo "$out" | grep -q "aissh" || fail "M3 aissh not in help: $out"
	pass "M3 discover aissh in help"
}

verify_m4() {
	local out
	out="$(run_ait $'install nosuch\nexit\n' 2>&1)" || true
	echo "$out" | grep -qi "official plugins" || fail "M4 missing hint: $out"
	pass "M4 install unknown plugin"
}

verify_m5() {
	local out
	out="$(run_ait $'update\nexit\n')"
	echo "$out" | grep -q "updated aissh -> $TAG" || fail "M5 update failed: $out"
	pass "M5 update installed"
}

verify_m6() {
	local out
	out="$(run_ait $'update aissh ailog\nexit\n')"
	echo "$out" | grep -q "skipped ailog" || fail "M6 missing skip: $out"
	echo "$out" | grep -q "updated aissh -> $TAG" || fail "M6 missing update aissh: $out"
	pass "M6 update skip not installed"
}

verify_m7() {
	local fresh="/tmp/ait-sandbox-empty-$$"
	rm -rf "$fresh"
	mkdir -p "$fresh/bin" "$fresh/home"
	cp "$ROOT/bin/ait" "$fresh/bin/ait"
	chmod +x "$fresh/bin/ait"
	local out
	out="$(AIT_HOME="$fresh/home" AIT_RELEASE_TEST_BASE_URL="$MOCK_URL" "$fresh/bin/ait" <<< $'update\nexit\n' 2>&1)" || true
	echo "$out" | grep -qi "Nothing to update\|nothing to update\|No installed official" || fail "M7 expected nothing to update: $out"
	rm -rf "$fresh"
	pass "M7 update nothing installed"
}

verify_m8() {
	local out
	out="$(run_ait $'upgrade aissh\nexit\n' 2>&1)" || true
	echo "$out" | grep -qi "accepts no plugin names\|no plugin names" || fail "M8 upgrade should reject args: $out"
	pass "M8 upgrade rejects args"
}

verify_m9() {
	local out
	out="$(run_ait $'upgrade\nexit\n')"
	echo "$out" | grep -q "upgrade complete" || fail "M9 missing upgrade complete: $out"
	test -f "$SANDBOX/bin/ait" || fail "M9 missing ait binary"
	pass "M9 upgrade"
}

verify_m10() {
	cp "$ROOT/bin/ait" "$SANDBOX/bin/ait"
	chmod +x "$SANDBOX/bin/ait"
	local out
	out="$(AIT_HOME="$SANDBOX/home" AIT_RELEASE_TEST_BASE_URL=http://127.0.0.1:1 "$SANDBOX/bin/ait" <<< $'install aissh\nexit\n' 2>&1)" || true
	echo "$out" | grep -qiE "Try:|connection refused|No GitHub release|ErrNoRelease" || fail "M10 expected fetch error: $out"
	pass "M10 mock down errhint"
}

verify_dist_layout() {
	arch="$(go env GOOS)-$(go env GOARCH)"
	archive="$ROOT/dist/ait-${TAG}-${arch}.tar.gz"
	test -f "$archive" || fail "missing $archive (run: make release-archives TAG=$TAG)"
	tar tzf "$archive" | grep -qE '(^|/)ait$' || fail "dist archive missing ait binary"
	tar tzf "$archive" | grep -q 'plugins/aissh/aissh' || fail "dist archive missing plugins/aissh"
	pass "dist archive layout ($archive)"
}

verify_dist_apply() {
	local dist_mock="${DIST_MOCK_URL:-http://127.0.0.1:9887}"
	local arch="$(go env GOOS)-$(go env GOARCH)"
	local archive="$ROOT/dist/ait-${TAG}-${arch}.tar.gz"
	test -f "$archive" || fail "missing $archive"
	if ! curl -sf "$dist_mock/repos/os-artificer/AiT/releases/latest" >/dev/null 2>&1; then
		fail "dist mock not reachable at $dist_mock — start: go run ./cmd/aitrelease-mock -addr :9887 -archive $archive"
	fi
	local fresh="/tmp/ait-sandbox-dist"
	rm -rf "$fresh"
	mkdir -p "$fresh/bin" "$fresh/home"
	cp "$ROOT/bin/ait" "$fresh/bin/ait"
	chmod +x "$fresh/bin/ait"
	local out
	out="$(AIT_HOME="$fresh/home" AIT_RELEASE_TEST_BASE_URL="$dist_mock" "$fresh/bin/ait" <<< $'install aissh\nexit\n')"
	echo "$out" | grep -q "installed aissh -> $TAG" || fail "dist apply install failed: $out"
	test -f "$fresh/bin/plugins/aissh/aissh" || fail "dist apply missing plugin binary"
	pass "dist archive apply via mock"
}

main() {
	case "${1:-all}" in
	setup)
		setup_sandbox
		echo "export SANDBOX=$SANDBOX"
		echo "export AIT_HOME=$SANDBOX/home"
		echo "export AIT_RELEASE_TEST_BASE_URL=$MOCK_URL"
		;;
	verify)
		check_mock
		verify_m1
		verify_m2
		verify_m3
		verify_m4
		verify_m5
		verify_m6
		verify_m7
		verify_m8
		verify_m9
		verify_m10
		echo ">> all M1–M10 passed"
		;;
	all)
		setup_sandbox
		check_mock
		verify_m1
		verify_m2
		verify_m3
		verify_m4
		verify_m5
		verify_m6
		verify_m7
		verify_m8
		verify_m9
		verify_m10
		echo ">> sandbox verify complete ($SANDBOX)"
		;;
	verify-dist)
		verify_dist_layout
		verify_dist_apply
		echo ">> dist verify complete"
		;;
	*)
		echo "usage: $0 {setup|verify|all|verify-dist}" >&2
		exit 2
		;;
	esac
}

main "$@"

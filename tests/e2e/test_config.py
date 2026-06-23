"""config init/view/validate/set/get E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.cli import assert_exit
from tests.lib.hints import assert_contains, assert_hint

pytestmark = pytest.mark.smoke


def test_config_init_creates_file(runner):
    proc = runner.run_ok(["config", "init"])
    assert_contains(proc, "config", stream="stdout")


def test_config_init_dry_run(runner):
    proc = runner.run_ok(["config", "init", "--dry-run"])
    assert "Would" in proc.stdout or "would" in proc.stdout.lower()


def test_config_view_after_init(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run_ok(["config", "view"])
    assert "session:" in proc.stdout
    assert "ai:" in proc.stdout


def test_config_validate_ok(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run_ok(["config", "validate"])
    assert_contains(proc, "config OK", stream="stdout")


def test_config_validate_missing_file(runner):
    proc = runner.run(["config", "validate"])
    assert_exit(1, proc)


def test_config_get_remote_prompt_path(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run_ok(["config", "get", "session.remote_prompt.path"])
    assert "short" in proc.stdout or "absolute" in proc.stdout


def test_config_set_remote_prompt_path(runner):
    runner.run_ok(["config", "init"])
    runner.run_ok(["config", "set", "session.remote_prompt.path", "short"])
    proc = runner.run_ok(["config", "get", "session.remote_prompt.path"])
    assert "short" in proc.stdout


def test_config_unknown_key_hint(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run(["config", "get", "bad.key"])
    assert_exit(1, proc)
    assert_hint(
        proc,
        explanation_substr="unknown config key",
        try_commands=("aitssh config set --help",),
    )


def test_config_edit_with_true_editor(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run_ok(
        ["config", "edit"],
        extra_env={"VISUAL": "true", "EDITOR": "true"},
    )
    # edit runs true (no-op) then validate
    assert proc.returncode == 0

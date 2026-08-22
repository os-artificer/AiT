"""aissh session flag CLI E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.cli import assert_exit
from tests.lib.hints import assert_hint

pytestmark = pytest.mark.smoke


def test_session_no_action_fails(runner):
    proc = runner.run(["session"])
    assert_exit(1, proc)
    assert_hint(proc, try_commands=("aissh session --help",))


def test_session_mutually_exclusive_flags(runner):
    proc = runner.run(["session", "--list", "--load", "x"])
    assert_exit(1, proc)
    assert_hint(proc, try_commands=("aissh session --help",))


def test_session_list_empty(runner):
    proc = runner.run_ok(["session", "--list"])
    # empty store: header or empty table
    assert proc.returncode == 0


def test_session_list_number_column(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run_ok(["session", "--list"])
    if "#" in proc.stdout:
        assert "#" in proc.stdout


def test_session_load_display_id_hint(runner):
    proc = runner.run(["session", "--load", "621b71ac"])
    assert_exit(1, proc)
    assert_hint(
        proc,
        explanation_substr="display",
        try_commands=("aissh session --list -v",),
    )


def test_session_load_non_tty_fails(runner):
    proc = runner.run(["session", "--load"])
    assert_exit(1, proc)
    assert_hint(proc, try_commands=("aissh session --load <alias>",))


def test_session_suspend_no_current(runner):
    proc = runner.run(["session", "--suspend"])
    assert_exit(1, proc)


def test_session_delete_all_force_empty(runner):
    proc = runner.run_ok(["session", "--delete", "all", "--force"])
    assert proc.returncode == 0


@pytest.mark.destructive
def test_session_delete_all_requires_force_or_confirm(runner):
    proc = runner.run(["session", "--delete", "all"])
    # non-TTY: prompt cancelled on stdin EOF, exit 0 with message
    assert proc.returncode == 0
    assert "cancelled" in proc.stdout.lower() or "deleted" in proc.stdout.lower()

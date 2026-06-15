"""Local AI CLI non-interactive smoke tests."""

from __future__ import annotations

import pytest

from tests.lib.cli import assert_exit
from tests.lib.hints import assert_hint

pytestmark = [pytest.mark.smoke]


def test_ai_model_list_non_interactive(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run(["ai", "model", "list"], timeout=5)
    # may fail without keys but must not hang
    assert proc.returncode in (0, 1)


def test_ai_ask_ping_non_tty(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run(["ai", "ask", "ping"], timeout=5)
    # without API key: fail fast with hint; with key: 0
    assert proc.returncode in (0, 1)


def test_ai_ask_empty_non_tty(runner):
    proc = runner.run(["ai", "ask"], timeout=5)
    assert_exit(1, proc)
    assert_hint(proc, try_commands=("ait ai ask <message>",))


def test_ai_ask_stdin_non_tty(runner):
    runner.run_ok(["config", "init"])
    proc = runner.run(["ai", "ask"], input_text="ping\n", timeout=5)
    assert proc.returncode in (0, 1)


def test_ai_ask_help(runner):
    proc = runner.run_ok(["ai", "ask", "--help"])
    assert "Usage:" in proc.stdout


def test_ai_agent_help(runner):
    proc = runner.run_ok(["ai", "agent", "--help"])
    assert "Usage:" in proc.stdout


@pytest.mark.ai
def test_ai_ask_with_configured_key(runner, test_config):
    if not test_config.ai.deepseek_key:
        pytest.skip("deepseek_key not set")
    runner.run_ok(["config", "init"])
    runner.run_ok(
        ["config", "set", "ai.deepseek.api_key", test_config.ai.deepseek_key],
    )
    proc = runner.run(["ai", "ask", "reply with exactly: pong"], timeout=60)
    assert proc.returncode == 0
    assert "pong" in proc.stdout.lower() or len(proc.stdout.strip()) > 0

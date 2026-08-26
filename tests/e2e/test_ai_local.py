"""Local AI CLI non-interactive smoke tests (bin/ai)."""

from __future__ import annotations

import pytest

from tests.lib.cli import AITRunner, assert_exit
from tests.lib.hints import assert_hint

pytestmark = [pytest.mark.smoke]


def test_ai_experimental_disabled_by_default(runner, ai_bin):
    runner.run_ok(["config", "init"])
    off = AITRunner(ai_bin, runner.ait_home, base_env=runner.base_env)
    proc = off.run(["ask", "ping"], timeout=5)
    assert_exit(1, proc)
    assert_hint(
        proc,
        sentinel="AI experimental features are disabled",
        try_commands=(
            "ai --ai-experimental",
            "aissh config set ai.experimental.enabled true",
        ),
    )


def test_ai_help_without_experimental(ai_bin, runner):
    off = AITRunner(ai_bin, runner.ait_home, base_env=runner.base_env)
    proc = off.run_ok(["--help"])
    assert "Usage:" in proc.stdout
    assert "experimental" in proc.stdout.lower() or "ask" in proc.stdout


def test_ai_model_list_non_interactive(runner, ai_runner):
    runner.run_ok(["config", "init"])
    proc = ai_runner.run(["model", "list"], timeout=5)
    # may fail without keys but must not hang
    assert proc.returncode in (0, 1)


def test_ai_ask_ping_non_tty(runner, ai_runner):
    runner.run_ok(["config", "init"])
    proc = ai_runner.run(["ask", "ping"], timeout=5)
    # without API key: fail fast with hint; with key: 0
    assert proc.returncode in (0, 1)


def test_ai_ask_empty_non_tty(ai_runner):
    proc = ai_runner.run(["ask"], timeout=5)
    assert_exit(1, proc)
    assert_hint(proc, try_commands=("ai ask <message>",))


def test_ai_ask_stdin_non_tty(runner, ai_runner):
    runner.run_ok(["config", "init"])
    proc = ai_runner.run(["ask"], input_text="ping\n", timeout=5)
    assert proc.returncode in (0, 1)


def test_ai_ask_help(ai_runner):
    proc = ai_runner.run_ok(["ask", "--help"])
    assert "Usage:" in proc.stdout


def test_ai_agent_help(ai_runner):
    proc = ai_runner.run_ok(["agent", "--help"])
    assert "Usage:" in proc.stdout


@pytest.mark.ai
def test_ai_ask_with_configured_key(runner, ai_runner, test_config):
    if not test_config.ai.deepseek_key:
        pytest.skip("deepseek_key not set")
    runner.run_ok(["config", "init"])
    runner.run_ok(
        ["config", "set", "ai.deepseek.api_key", test_config.ai.deepseek_key],
    )
    proc = ai_runner.run(["ask", "reply with exactly: pong"], timeout=60)
    assert proc.returncode == 0
    assert "pong" in proc.stdout.lower() or len(proc.stdout.strip()) > 0

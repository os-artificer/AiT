"""AI provider and ollama CLI smoke tests (bin/ai)."""

from __future__ import annotations

import pytest

pytestmark = pytest.mark.smoke


def test_ai_provider_help(ai_runner):
    proc = ai_runner.run_ok(["provider", "--help"])
    assert "enable" in proc.stdout and "disable" in proc.stdout


def test_ai_ollama_status(runner, ai_runner):
    runner.run_ok(["config", "init"])
    proc = ai_runner.run(["ollama", "status"], timeout=30)
    assert proc.returncode in (0, 1)


def test_ai_ollama_setup_dry_run(runner, ai_runner):
    runner.run_ok(["config", "init"])
    proc = ai_runner.run(["ollama", "setup", "--dry-run"], timeout=30)
    assert proc.returncode in (0, 1)


@pytest.mark.ai
def test_ai_provider_disable_enable_ollama(runner, ai_runner, test_config):
    runner.run_ok(["config", "init"])
    proc_off = ai_runner.run(["provider", "disable", "ollama"], timeout=30)
    assert proc_off.returncode in (0, 1)
    proc_on = ai_runner.run(
        ["provider", "enable", "ollama", "--yes", "--no-install", "--no-pull"],
        timeout=30,
    )
    assert proc_on.returncode in (0, 1)

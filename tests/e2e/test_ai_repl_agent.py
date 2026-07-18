"""Local AI REPL agent mode smoke (A20-A22 subset)."""

from __future__ import annotations

import pexpect
import pytest

from tests.lib.expect import runner_env_from

pytestmark = [pytest.mark.ai]


@pytest.fixture
def ai_repl_session(test_config, ait_home, ai_runner):
    ai_runner.run_ok(["config", "init"])
    if test_config.ai.deepseek_key:
        ai_runner.run_ok(["config", "set", "ai.deepseek.api_key", test_config.ai.deepseek_key])
    env = runner_env_from(
        test_config,
        __import__("pathlib").Path(ait_home),
        ai_experimental=True,
    )
    child = pexpect.spawn(
        str(test_config.binary),
        ["ai"],
        encoding="utf-8",
        timeout=60,
        env=env,
    )
    child.expect(r"AiT AI\(Agent\)", timeout=30)
    yield child
    try:
        child.sendline("exit")
        child.expect(pexpect.EOF, timeout=10)
    except Exception:
        child.close(force=True)


def test_ai_repl_mode_status(ai_repl_session):
    ai_repl_session.sendline("mode")
    ai_repl_session.expect(r"agent|ask|approve", timeout=15)


def test_ai_repl_mode_ask_switch(ai_repl_session):
    ai_repl_session.sendline("mode ask")
    ai_repl_session.expect(r"Ask", timeout=15)
    ai_repl_session.sendline("mode agent")
    ai_repl_session.expect(r"Agent", timeout=15)


def test_ai_repl_help_lists_commands(ai_repl_session):
    ai_repl_session.sendline("help")
    ai_repl_session.expect(r"mode|approve|run|history", timeout=15)

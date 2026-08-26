"""Local ait agent Loop smoke (v4; replaces legacy aissh ai REPL)."""

from __future__ import annotations

import pexpect
import pytest

from tests.lib.expect import runner_env_from

pytestmark = [pytest.mark.ai]


@pytest.fixture
def agent_repl_session(test_config, ait_home, runner, toolbox_bin):
    runner.run_ok(["config", "init"])
    if test_config.ai.deepseek_key:
        runner.run_ok(["config", "set", "ai.deepseek.api_key", test_config.ai.deepseek_key])
    env = runner_env_from(
        test_config,
        __import__("pathlib").Path(ait_home),
        ai_experimental=True,
    )
    child = pexpect.spawn(
        toolbox_bin,
        ["agent"],
        encoding="utf-8",
        timeout=60,
        env=env,
    )
    child.expect(r"AiT Agent>", timeout=30)
    yield child
    try:
        child.sendline("exit")
        child.expect(pexpect.EOF, timeout=10)
    except Exception:
        child.close(force=True)


def test_agent_repl_prompt(agent_repl_session):
    agent_repl_session.sendline("model")
    # Status line or empty is fine; must not hang or crash.
    agent_repl_session.expect(r"AiT Agent>|provider|model|deepseek|ollama|minimax|AI", timeout=15)


def test_agent_repl_history_meta(agent_repl_session):
    agent_repl_session.sendline("history")
    agent_repl_session.expect(r"AiT Agent>", timeout=15)


def test_agent_repl_exit(agent_repl_session):
    agent_repl_session.sendline("exit")
    agent_repl_session.expect(pexpect.EOF, timeout=10)

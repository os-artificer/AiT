"""Remote AI via !! and !ai-ssh ai ask (R19)."""

from __future__ import annotations

import pytest

from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = [pytest.mark.ssh, pytest.mark.ai]


@pytest.fixture
def ai_remote_session(test_config, ait_home, runner):
    runner.run_ok(["config", "init"])
    if test_config.ai.deepseek_key:
        runner.run_ok(["config", "set", "ai.deepseek.api_key", test_config.ai.deepseek_key])
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()


def test_remote_ai_ask_one_shot(ai_remote_session):
    out = ai_remote_session.meta("ai ask ping")
    assert out is not None
    assert len(out.strip()) >= 0


def test_remote_ai_bang_enters_repl(ai_remote_session):
    ai_remote_session.bang_ai()
    out = ai_remote_session.send_line("mode")
    assert "agent" in out.lower() or "ask" in out.lower() or "mode" in out.lower()
    ai_remote_session.send_line("exit")


def test_remote_ai_bare_enters_repl(ai_remote_session):
    ai_remote_session.child.sendline("!ai-ssh ai")
    ai_remote_session.child.expect(r"AiT AI\([^)]+\)", timeout=120)
    ai_remote_session.child.sendline("mode")
    ai_remote_session.child.expect(r"agent", timeout=30)
    ai_remote_session.child.sendline("exit")


def test_remote_ai_agent_bare_enters_repl(ai_remote_session):
    ai_remote_session.child.sendline("!ai-ssh ai agent")
    ai_remote_session.child.expect(r"AiT AI\([^)]+\)", timeout=120)
    ai_remote_session.child.sendline("mode")
    ai_remote_session.child.expect(r"agent", timeout=30)
    ai_remote_session.child.sendline("exit")


def test_remote_ai_utf8_query(ai_remote_session):
    out = ai_remote_session.meta('ai ask 你好')
    assert out is not None

"""aissh shell REPL smoke tests (pexpect)."""

from __future__ import annotations

import pytest

from tests.lib.config import load_config
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.smoke


@pytest.fixture
def shell_session(test_config, ait_home):
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_shell(test_config, env)
    yield sess
    sess.close()


def test_shell_help(shell_session):
    out = shell_session.send_line("help")
    assert "Usage:" in out or "Keys" in out or "help" in out.lower()


def test_shell_list_empty(shell_session):
    out = shell_session.send_line("list")
    # empty store or no sessions message
    assert "session" in out.lower() or "#" in out or out.strip() == ""


def test_shell_ll_lists_cwd(shell_session, tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    out = shell_session.send_line("ll")
    assert "total" in out.lower() or "." in out


def test_shell_history(shell_session):
    shell_session.send_line("echo-marker-test")
    out = shell_session.send_line("history")
    assert "echo-marker-test" in out or "history" in out.lower()


def test_shell_config_view(shell_session):
    out = shell_session.send_line("config view")
    assert "session:" in out or "ai:" in out


def test_shell_unknown_command_hint(shell_session):
    out = shell_session.send_line("nosuchcmd-e2e-xyz")
    assert "Try:" in out or "unknown" in out.lower() or "not found" in out.lower()

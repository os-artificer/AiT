"""Interactive SSH login and remote prompt E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.config import load_config
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.ssh


@pytest.fixture
def login_session(test_config, ait_home):
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()


def test_login_shows_ait_prompt(login_session, test_config):
    buf = login_session.visible_output(login_session.buffer())
    assert test_config.primary.user in buf
    if login_session.native_prompt:
        assert "➜" in buf or "#" in buf or test_config.primary.host.split(".")[0] in buf
    else:
        assert "AiT" in buf


def test_remote_ls_returns_output(login_session):
    out = login_session.send_line("ls")
    assert out is not None


def test_remote_pwd_then_prompt(login_session):
    login_session.send_line("pwd")
    # second command confirms prompt returned (P11)
    out = login_session.send_line("echo e2e-prompt-ok", expect_substr="e2e-prompt-ok", retries=2)
    assert "e2e-prompt-ok" in login_session.visible_output(out)


def test_login_one_shot_remote_command(runner, test_config):
    proc = runner.run(
        test_config.primary.one_shot_login_argv(["echo", "e2e-oneshot"]),
        timeout=60,
    )
    assert proc.returncode == 0
    assert "e2e-oneshot" in proc.stdout

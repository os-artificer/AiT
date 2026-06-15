"""login --tmux E2E (TM* subset)."""

from __future__ import annotations

import uuid

import pytest

from tests.lib.cli import AITRunner
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = [pytest.mark.ssh, pytest.mark.tmux]


@pytest.fixture
def tmux_runner(test_config, tmp_path):
    home = tmp_path / "tmux_home"
    home.mkdir()
    env = runner_env_from(test_config, home)
    runner = AITRunner(
        test_config.binary,
        home,
        base_env={
            "AIT_KEYRING_FILE_PASSWORD": test_config.keyring_password,
            "LANG": test_config.lang,
        },
    )
    return runner, test_config, env


def test_login_tmux_attach(tmux_runner):
    _runner, cfg, env = tmux_runner
    alias = f"tmux-{uuid.uuid4().hex[:6]}"
    sess = AITExpectSession.spawn_login(
        cfg,
        env,
        extra_login_args=["--tmux", "--alias", alias],
        timeout=120,
    )
    try:
        out = sess.send_line("echo tmux-e2e-ok")
        assert "tmux-e2e-ok" in out
    finally:
        sess.close()

"""Chain login/load/switch via !ait meta (N*)."""

from __future__ import annotations

import uuid

import pytest

from tests.lib.cli import AITRunner
from tests.lib.config import cli_base_env
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.ssh


def _alias() -> str:
    return f"chain-{uuid.uuid4().hex[:6]}"


@pytest.fixture
def chain_env(test_config, tmp_path):
    home = tmp_path / "chain_home"
    home.mkdir()
    return test_config, runner_env_from(test_config, home)


def test_meta_suspend_lists_session(chain_env):
    cfg, env = chain_env
    alias = _alias()
    from tests.lib.cli import AITRunner

    runner = AITRunner(
        cfg.binary,
        __import__("pathlib").Path(env["AIT_HOME"]),
        base_env=cli_base_env(cfg, env["AIT_HOME"]),
    )
    runner.run_ok(
        cfg.primary.one_shot_login_argv(["echo", "chain"], ["--alias", alias]),
        timeout=90,
    )
    sess = AITExpectSession.spawn_login(cfg, env, extra_login_args=["--alias", alias])
    try:
        out = sess.visible_output(sess.meta("session --list"))
        if alias not in out:
            out = sess.visible_output(sess.meta("session --list"))
        assert alias in out
    finally:
        sess.close()

"""Nested ssh exit restores outer AiT prompt (N6)."""

from __future__ import annotations

import pytest

from tests.lib.config import load_config
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.ssh


@pytest.fixture
def nested_session(test_config, ait_home):
    if not test_config.features.nested_ssh or not test_config.inner.host:
        pytest.skip("features.nested_ssh disabled or ssh.inner not configured")
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()


def test_nested_ssh_exit_restores_outer_prompt(nested_session, test_config):
    inner = test_config.inner
    ssh_line = f"ssh -p {inner.port} {inner.user}@{inner.host}" if inner.port != 22 else f"ssh {inner.user}@{inner.host}"
    nested_session.send_line(ssh_line, timeout=120)
    nested_session.send_line("exit", timeout=60)
    nested_session.child.expect("Connection to", timeout=30)
    # outer dial host prompt
    nested_session.expect_ait_prompt(
        test_config.primary.user,
        test_config.primary.host.split(".")[0],
        timeout=60,
    )
    out = nested_session.send_line("echo outer-ok")
    assert "outer-ok" in out

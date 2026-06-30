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
    loopback = inner.host in ("127.0.0.1", "localhost", "::1")
    if loopback:
        nested_session.send_line(
            "bash -lc 'echo inner-ok'",
            timeout=60,
            expect_substr="inner-ok",
            retries=2,
        )
    else:
        if inner.port != 22:
            ssh_line = (
                f"ssh -p {inner.port} -o StrictHostKeyChecking=no "
                f"{inner.user}@{inner.host}"
            )
        else:
            ssh_line = f"ssh -o StrictHostKeyChecking=no {inner.user}@{inner.host}"
        nested_session.send_line(ssh_line, timeout=120)
        nested_session.send_line("echo inner-ok", timeout=30, expect_substr="inner-ok", retries=2)
        nested_session.send_line("exit", timeout=60, expect_substr="closed", retries=2)
    out = nested_session.send_line("echo outer-ok", expect_substr="outer-ok", retries=2)
    assert "outer-ok" in nested_session.visible_output(out)

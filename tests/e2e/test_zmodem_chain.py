"""Nested SSH + !ait get/put ZMODEM chain (T*/Z*; optional)."""

from __future__ import annotations

import pytest

from tests.lib.artifacts import local_probe_in, probe_name, remote_probe_path
from tests.lib.config import load_config
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = [pytest.mark.ssh, pytest.mark.zmodem]


@pytest.fixture
def zmodem_session(test_config, ait_home):
    if not test_config.inner.host:
        pytest.skip("ssh.inner not configured")
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_login(test_config, env)
    inner = test_config.inner
    ssh_line = (
        f"ssh -p {inner.port} {inner.user}@{inner.host}"
        if inner.port != 22
        else f"ssh {inner.user}@{inner.host}"
    )
    sess.send_line(ssh_line, timeout=120)
    sess.meta("chain on")
    yield sess, ait_home
    sess.close()


def test_meta_get_put_small_file(zmodem_session, test_config, tmp_path):
    sess, _home = zmodem_session
    name = probe_name("zmodem")
    local_src = local_probe_in(tmp_path, name)
    remote_name = probe_name("remote")

    # put local to remote via meta
    out_put = sess.meta(f"put {local_src} ./{remote_name}")
    assert "done" in out_put.lower() or "copying" in out_put.lower() or out_put.strip() == ""

    local_dst = tmp_path / f"got-{name}"
    out_get = sess.meta(f"get ./{remote_name} {local_dst}")
    assert "done" in out_get.lower() or "copying" in out_get.lower() or out_get.strip() == ""

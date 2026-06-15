"""Standalone ait get/put SFTP E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.artifacts import local_probe_in, probe_name, remote_probe_path

pytestmark = pytest.mark.ssh


def test_get_put_roundtrip(runner, test_config, tmp_path):
    name = probe_name()
    local_src = local_probe_in(tmp_path, name)
    remote = remote_probe_path(name)

    put_args = test_config.primary.login_argv()
    put_args = ["put", *put_args[1:]] if put_args[0] == "login" else ["put"]
    # put: local remote
    put_argv = ["put"]
    if test_config.primary.port != 22:
        put_argv.extend(["-p", str(test_config.primary.port)])
    if test_config.primary.identity:
        put_argv.extend(["-i", test_config.primary.identity])
    put_argv.extend([str(local_src), f"{test_config.primary.destination}:{remote}"])
    proc_put = runner.run(put_argv, timeout=120)
    assert proc_put.returncode == 0, proc_put.stderr

    local_dst = tmp_path / f"down-{name}"
    get_argv = ["get"]
    if test_config.primary.port != 22:
        get_argv.extend(["-p", str(test_config.primary.port)])
    if test_config.primary.identity:
        get_argv.extend(["-i", test_config.primary.identity])
    get_argv.extend([f"{test_config.primary.destination}:{remote}", str(local_dst)])
    proc_get = runner.run(get_argv, timeout=120)
    assert proc_get.returncode == 0, proc_get.stderr
    assert local_dst.is_file()
    assert "ait-e2e-probe" in local_dst.read_text(encoding="utf-8")


def test_get_single_path_default_basename(runner, test_config, tmp_path):
    """TR5: standalone get single remote path defaults to ./basename."""
    name = probe_name()
    local_src = local_probe_in(tmp_path, name)
    remote = remote_probe_path(name)

    put_argv = ["put"]
    if test_config.primary.port != 22:
        put_argv.extend(["-p", str(test_config.primary.port)])
    if test_config.primary.identity:
        put_argv.extend(["-i", test_config.primary.identity])
    put_argv.extend([str(local_src), f"{test_config.primary.destination}:{remote}"])
    runner.run(put_argv, timeout=120)

    cwd = tmp_path / "dl"
    cwd.mkdir()
    get_argv = ["get"]
    if test_config.primary.port != 22:
        get_argv.extend(["-p", str(test_config.primary.port)])
    if test_config.primary.identity:
        get_argv.extend(["-i", test_config.primary.identity])
    get_argv.append(f"{test_config.primary.destination}:{remote}")
    proc = runner.run(get_argv, timeout=120, cwd=cwd)
    assert proc.returncode == 0, proc.stderr
    assert (cwd / name).is_file()

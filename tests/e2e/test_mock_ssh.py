"""CLI user scenarios against the in-repo testenv-mock SSH+SFTP daemon (MS13–MS16)."""

from __future__ import annotations

from pathlib import Path

import pytest

from tests.lib.mockenv import ssh_client_env, ssh_connect_args, start_mock

pytestmark = [pytest.mark.smoke, pytest.mark.mock]


@pytest.fixture(scope="module")
def mock_env():
    env = start_mock()
    yield env
    env.close()


def test_oneshot_echo(runner, mock_env, tmp_path: Path):
    ssh = mock_env.ssh
    proc = runner.run(
        ssh_connect_args(ssh, extra=["echo", "e2e-oneshot"]),
        extra_env=ssh_client_env(tmp_path),
        timeout=20,
    )
    assert proc.returncode == 0, proc.output
    assert "e2e-oneshot" in proc.stdout


def test_oneshot_unknown_command(runner, mock_env, tmp_path: Path):
    ssh = mock_env.ssh
    proc = runner.run(
        ssh_connect_args(ssh, extra=["nosuchcmd"]),
        extra_env=ssh_client_env(tmp_path),
        timeout=20,
    )
    assert proc.returncode != 0


def test_get_welcome(runner, mock_env, tmp_path: Path):
    ssh = mock_env.ssh
    dest = tmp_path / "welcome.out"
    proc = runner.run(
        [
            "get",
            "-p",
            str(ssh["port"]),
            "--password",
            ssh["password"],
            f"{ssh['user']}@{ssh['host']}:{ssh['welcome']}",
            str(dest),
        ],
        extra_env=ssh_client_env(tmp_path),
        timeout=20,
    )
    assert proc.returncode == 0, proc.output
    assert dest.is_file()
    assert "hello from mock ssh" in dest.read_text(encoding="utf-8")


def test_get_missing(runner, mock_env, tmp_path: Path):
    ssh = mock_env.ssh
    dest = tmp_path / "missing.out"
    proc = runner.run(
        [
            "get",
            "-p",
            str(ssh["port"]),
            "--password",
            ssh["password"],
            f"{ssh['user']}@{ssh['host']}:{ssh['root']}/no-such-file.bin",
            str(dest),
        ],
        extra_env=ssh_client_env(tmp_path),
        timeout=20,
    )
    assert proc.returncode != 0


def test_put_then_get_roundtrip(runner, mock_env, tmp_path: Path):
    ssh = mock_env.ssh
    src = tmp_path / "up.txt"
    src.write_text("ait-e2e-probe\n", encoding="utf-8")
    remote = f"{ssh['root']}/from-cli.txt"
    dest_spec = f"{ssh['user']}@{ssh['host']}:{remote}"
    extra = ssh_client_env(tmp_path)

    put = runner.run(
        ["put", "-p", str(ssh["port"]), "--password", ssh["password"], str(src), dest_spec],
        extra_env=extra,
        timeout=20,
    )
    assert put.returncode == 0, put.output

    got = tmp_path / "down.txt"
    get = runner.run(
        ["get", "-p", str(ssh["port"]), "--password", ssh["password"], dest_spec, str(got)],
        extra_env=extra,
        timeout=20,
    )
    assert get.returncode == 0, get.output
    assert got.read_text(encoding="utf-8") == "ait-e2e-probe\n"


def test_bad_password(runner, tmp_path: Path):
    env = start_mock(reject_password=True)
    try:
        ssh = env.ssh
        proc = runner.run(
            [
                "-p",
                str(ssh["port"]),
                "-l",
                ssh["user"],
                "--password",
                "wrong-password",
                ssh["host"],
                "echo",
                "should-not-run",
            ],
            extra_env=ssh_client_env(tmp_path),
            timeout=20,
        )
        assert proc.returncode != 0
        assert "should-not-run" not in proc.stdout
    finally:
        env.close()


def test_connection_refused(runner, tmp_path: Path):
    env = start_mock()
    ssh = dict(env.ssh)
    env.close()
    proc = runner.run(
        ssh_connect_args(ssh, extra=["echo", "nope"]),
        extra_env=ssh_client_env(tmp_path),
        timeout=15,
    )
    assert proc.returncode != 0
    assert "nope" not in proc.stdout

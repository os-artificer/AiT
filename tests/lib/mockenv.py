"""Start the in-repo testenv-mock daemon for smoke E2E tests."""

from __future__ import annotations

import json
import os
import subprocess
import time
from pathlib import Path
from typing import Any


def repo_root() -> Path:
    return Path(__file__).resolve().parents[2]


def mock_binary() -> Path:
    root = repo_root()
    return root / "bin" / "testenv-mock"


def ensure_mock_binary() -> Path:
    path = mock_binary()
    if path.is_file():
        return path
    subprocess.run(
        ["go", "build", "-o", str(path), "./cmd/testenv-mock"],
        cwd=str(repo_root()),
        check=True,
    )
    return path


def ssh_client_env(home: Path) -> dict[str, str]:
    """Env for aissh against loopback mock sshd (no agent, isolated HOME)."""
    return {
        "AIT_SSH_INSECURE": "1",
        "SSH_AUTH_SOCK": "",
        "HOME": str(home),
    }


class TestEnvMock:
    """Subprocess wrapper around bin/testenv-mock."""

    def __init__(self, proc: subprocess.Popen[str], info: dict[str, Any]):
        self.proc = proc
        self.info = info

    @property
    def ssh(self) -> dict[str, Any]:
        return self.info["ssh"]

    @property
    def http(self) -> dict[str, Any]:
        return self.info["http"]

    def close(self) -> None:
        if self.proc.poll() is not None:
            return
        self.proc.send_signal(2)
        try:
            self.proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            self.proc.kill()


def start_mock(*, reject_password: bool = False, http_scenario: str = "ok") -> TestEnvMock:
    bin_path = ensure_mock_binary()
    args = [str(bin_path), "-http-scenario", http_scenario]
    if reject_password:
        args.append("-reject-password")
    proc = subprocess.Popen(
        args,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=str(repo_root()),
        env={**os.environ, "SSH_AUTH_SOCK": ""},
    )
    assert proc.stdout is not None
    line = ""
    deadline = time.time() + 10
    while time.time() < deadline:
        if proc.poll() is not None:
            err = proc.stderr.read() if proc.stderr else ""
            raise RuntimeError(f"testenv-mock exited {proc.returncode}: {err}")
        line = proc.stdout.readline()
        if line:
            break
        time.sleep(0.05)
    if not line:
        proc.kill()
        raise RuntimeError("testenv-mock produced no JSON banner")
    info = json.loads(line)
    return TestEnvMock(proc, info)


def ssh_connect_args(ssh: dict[str, Any], extra: list[str] | None = None) -> list[str]:
    args = [
        "-p",
        str(ssh["port"]),
        "-l",
        ssh["user"],
        "--password",
        ssh["password"],
        f"{ssh['host']}",
    ]
    if extra:
        args.extend(extra)
    return args


def apply_mock_target(cfg: Any, ssh: dict[str, Any]) -> None:
    """Replace TestConfig.primary with the loopback testenv-mock sshd (MS20)."""
    from tests.lib.config import SSHTarget

    cfg.primary = SSHTarget(
        host=str(ssh["host"]),
        port=int(ssh["port"]),
        user=str(ssh["user"]),
        identity="",
        password=str(ssh.get("password") or ""),
    )
    cfg.inner = SSHTarget()
    cfg.jump_hosts = []
    cfg.features.insecure_ssh = True
    cfg.features.nested_ssh = False
    cfg.features.tmux = False
    cfg.features.lrzsz = False
    cfg.features.zmodem_chain = False
    cfg.using_mock = True

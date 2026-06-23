"""pexpect helpers for interactive ai-ssh login, shell, and meta sessions."""

from __future__ import annotations

import os
import re
import time
from pathlib import Path
from typing import Sequence

import pexpect

from tests.lib.config import SSHTarget, TestConfig


class AITExpectSession:
    """Interactive PTY session around ai-ssh login or shell."""

    def __init__(self, child: pexpect.spawn, *, prompt_pattern: str | bytes):
        self.child = child
        self.prompt_pattern = prompt_pattern
        self._before_last = ""

    @classmethod
    def spawn_login(
        cls,
        cfg: TestConfig,
        runner_env: dict[str, str],
        target: SSHTarget | None = None,
        extra_login_args: Sequence[str] | None = None,
        *,
        timeout: int = 90,
    ) -> AITExpectSession:
        tgt = target or cfg.primary
        argv = [str(cfg.binary), *tgt.login_argv(list(extra_login_args or []))]
        if tgt.password:
            argv.append("--password")
            argv.append(tgt.password)
        child = pexpect.spawn(
            argv[0],
            argv[1:],
            encoding="utf-8",
            timeout=timeout,
            env=runner_env,
        )
        if tgt.password:
            idx = child.expect([r"[Pp]assword:", r"Password:", pexpect.TIMEOUT], timeout=30)
            if idx < 2:
                child.sendline(tgt.password)
        # AiT remote prompt: AiT user@host:...
        pattern = r"AiT\s+" + re.escape(tgt.user) + r"@" + r"[^\s:]+:"
        child.expect(pattern, timeout=timeout)
        return cls(child, prompt_pattern=pattern)

    @classmethod
    def spawn_shell(cls, cfg: TestConfig, runner_env: dict[str, str], *, timeout: int = 30) -> AITExpectSession:
        child = pexpect.spawn(
            str(cfg.binary),
            ["shell"],
            encoding="utf-8",
            timeout=timeout,
            env=runner_env,
        )
        child.expect(r"AiT>\s*", timeout=timeout)
        return cls(child, prompt_pattern=r"AiT>\s*")

    def send_line(self, line: str, *, timeout: int = 60) -> str:
        self.child.sendline(line)
        idx = self.child.expect([self.prompt_pattern, pexpect.TIMEOUT], timeout=timeout)
        if idx == 1:
            raise TimeoutError(f"timeout waiting for prompt after: {line!r}\n{self.child.before}")
        self._before_last = self.child.before or ""
        return self._before_last

    def meta(self, cmd: str, *, timeout: int = 90) -> str:
        return self.send_line(f"!ait {cmd}", timeout=timeout)

    def bang_ai(self, first_line: str = "", *, timeout: int = 120) -> str:
        line = "!!" if not first_line else f"!! {first_line}"
        self.child.sendline(line)
        self.child.expect(r"AiT AI\([^)]+\)", timeout=timeout)
        return self.child.before or ""

    def expect_output(self, substr: str, *, timeout: int = 30) -> None:
        self.child.expect(re.escape(substr), timeout=timeout)

    def expect_ait_prompt(self, user: str, host_fragment: str, *, timeout: int = 60) -> None:
        pat = rf"AiT\s+{re.escape(user)}@{re.escape(host_fragment)}"
        self.child.expect(pat, timeout=timeout)

    def assert_no_native_prompt(self, text: str) -> None:
        """Fail when oh-my-zsh or bracket-style native prompts appear in output."""
        native_markers = ("➜", "->", " [%] ", " [~] ")
        for marker in native_markers:
            if marker in text:
                raise AssertionError(f"native prompt marker {marker!r} in output: {text!r}")

    def assert_ait_remote_prompt(self, user: str, host_fragment: str, text: str) -> None:
        if "AiT " not in text and f"AiT {user}@" not in text:
            raise AssertionError(f"missing AiT prompt for {user}@{host_fragment}: {text!r}")
        self.assert_no_native_prompt(text)

    def close(self, *, timeout: int = 30) -> None:
        try:
            self.send_line("exit", timeout=timeout)
        except (pexpect.TIMEOUT, pexpect.EOF):
            pass
        try:
            self.child.expect(pexpect.EOF, timeout=10)
        except pexpect.TIMEOUT:
            self.child.close(force=True)
        if self.child.isalive():
            self.child.terminate(force=True)

    def buffer(self) -> str:
        return (self.child.before or "") + (self.child.after or "")


def runner_env_from(cfg: TestConfig, ait_home: Path) -> dict[str, str]:
    env = os.environ.copy()
    env["AIT_HOME"] = str(ait_home)
    env["AIT_KEYRING_BACKEND"] = "file"
    env["AIT_KEYRING_DIR"] = str(ait_home / "keyring")
    env["AIT_KEYRING_FILE_PASSWORD"] = cfg.keyring_password
    env["LANG"] = cfg.lang
    env["LC_ALL"] = cfg.lang
    if cfg.features.insecure_ssh:
        env["AIT_SSH_INSECURE"] = "1"
    return env

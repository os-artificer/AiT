"""pexpect helpers for interactive aissh login, shell, and meta sessions."""

from __future__ import annotations

import os
import re
import time
from pathlib import Path
from typing import Sequence

import pexpect

from tests.lib.config import SSHTarget, TestConfig, cli_base_env

# oh-my-zsh / MOTD prompts that block the shell prompt after SSH login.
_LOGIN_BLOCKER_PATTERNS = [
    r"Would you like to update",
    r"\[Y/n\]",
    r"\[y/N\]",
]

_ANSI_ESCAPE_RE = re.compile(r"\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\].*?\x07")


def strip_ansi(text: str) -> str:
    """Remove ANSI escape sequences for stable substring assertions."""
    return _ANSI_ESCAPE_RE.sub("", text)


def _login_prompt_patterns(tgt: SSHTarget) -> list[str]:
    """Prompt patterns for interactive login (AiT-managed or native shell)."""
    host_short = tgt.host.split(".")[0]
    patterns = [
        r"AiT\s+" + re.escape(tgt.user) + r"@" + r"[^\s:]+:\S*?>\s*",
        re.escape(tgt.user) + r"@" + re.escape(host_short) + r"[^\n]*➜",
        re.escape(tgt.user) + r"@.*➜",
        re.escape(tgt.user) + r"@.*#\s",
    ]
    return patterns


def _expect_prompt_dismissing_blockers(
    child: pexpect.spawn,
    patterns: list[str],
    *,
    timeout: int,
) -> int:
    """Wait for a shell prompt, answering common interactive blockers."""
    deadline = time.monotonic() + timeout
    blockers = list(_LOGIN_BLOCKER_PATTERNS)
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise pexpect.TIMEOUT(
                f"timeout waiting for login prompt\n{child.before}",
            )
        expect_list = patterns + blockers + [pexpect.TIMEOUT]
        idx = child.expect(expect_list, timeout=min(remaining, 15))
        if idx < len(patterns):
            return idx
        if idx < len(patterns) + len(blockers):
            child.sendline("n")
            continue
        raise pexpect.TIMEOUT(
            f"timeout waiting for login prompt\n{child.before}",
        )


def _drain_duplicate_prompts(child: pexpect.spawn, patterns: list[str], *, quiet: float = 0.8) -> None:
    """Consume extra prompt matches so send_line does not return on leftover PS1."""
    deadline = time.monotonic() + 5.0
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            return
        idx = child.expect(patterns + [pexpect.TIMEOUT], timeout=min(quiet, remaining))
        if idx == len(patterns):
            return


class AITExpectSession:
    """Interactive PTY session around aissh login or shell."""

    def __init__(
        self,
        child: pexpect.spawn,
        *,
        prompt_pattern: str | bytes,
        prompt_patterns: list[str] | None = None,
        native_prompt: bool = False,
    ):
        self.child = child
        self.prompt_pattern = prompt_pattern
        self.prompt_patterns = prompt_patterns or [str(prompt_pattern)]
        self.native_prompt = native_prompt
        self.login_banner = ""
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
        child = pexpect.spawn(
            argv[0],
            argv[1:],
            encoding="utf-8",
            timeout=timeout,
            env=runner_env,
        )
        # --password on argv is consumed by aissh; do not wait 30s for a prompt.
        if tgt.password and "--password" not in argv:
            idx = child.expect([r"[Pp]assword:", r"Password:", pexpect.TIMEOUT], timeout=30)
            if idx < 2:
                child.sendline(tgt.password)
        patterns = _login_prompt_patterns(tgt)
        idx = _expect_prompt_dismissing_blockers(child, patterns, timeout=timeout)
        after = child.after if isinstance(child.after, str) else ""
        banner = (child.before or "") + after
        _drain_duplicate_prompts(child, patterns)
        native_prompt = idx != 0
        prompt_pattern = patterns[idx]
        sess = cls(
            child,
            prompt_pattern=prompt_pattern,
            prompt_patterns=patterns,
            native_prompt=native_prompt,
        )
        sess.login_banner = banner
        return sess

    @classmethod
    def spawn_shell(cls, cfg: TestConfig, runner_env: dict[str, str], *, timeout: int = 30) -> AITExpectSession:
        child = pexpect.spawn(
            str(cfg.binary),
            ["shell"],
            encoding="utf-8",
            timeout=timeout,
            env=runner_env,
        )
        pat = r"AiT>\s*"
        child.expect(pat, timeout=timeout)
        return cls(child, prompt_pattern=pat, prompt_patterns=[pat])

    def send_line(
        self,
        line: str,
        *,
        timeout: int = 60,
        expect_substr: str | None = None,
        retries: int = 1,
    ) -> str:
        last = ""
        for attempt in range(retries):
            self.child.sendline(line)
            if expect_substr:
                substr_idx = self.child.expect(
                    [re.escape(expect_substr), pexpect.TIMEOUT],
                    timeout=timeout,
                )
                if substr_idx == 1:
                    if attempt + 1 >= retries:
                        raise TimeoutError(
                            f"timeout waiting for {expect_substr!r} after: {line!r}\n{self.child.before}",
                        )
                    time.sleep(0.25)
                    continue
            idx = self.child.expect(self.prompt_patterns + [pexpect.TIMEOUT], timeout=timeout)
            if idx == len(self.prompt_patterns):
                raise TimeoutError(f"timeout waiting for prompt after: {line!r}\n{self.child.before}")
            last = self.child.before or ""
            if expect_substr is None and not strip_ansi(last).strip():
                # Late duplicate PS1 (setup redraw) can match before meta/CLI output arrives.
                idx = self.child.expect(self.prompt_patterns + [pexpect.TIMEOUT], timeout=timeout)
                if idx == len(self.prompt_patterns):
                    raise TimeoutError(f"timeout waiting for prompt after: {line!r}\n{self.child.before}")
                last = self.child.before or ""
            if expect_substr is None or expect_substr in strip_ansi(last):
                self._before_last = last
                return last
            if attempt + 1 < retries:
                time.sleep(0.25)
        self._before_last = last
        return last

    def visible_output(self, text: str | None = None) -> str:
        """Return recent PTY output with ANSI stripped (meta notices use AiT: on stderr)."""
        raw = text if text is not None else self._before_last
        return strip_ansi(raw or "")

    def meta(self, cmd: str, *, timeout: int = 90) -> str:
        token = (cmd.split() or ["!ait"])[0]
        return self.send_line(f"!ait {cmd}", timeout=timeout, expect_substr=token)

    def bang_ai(self, first_line: str = "", *, timeout: int = 120) -> str:
        line = "!!" if not first_line else f"!! {first_line}"
        self.child.sendline(line)
        self.child.expect(r"AiT AI\([^)]+\)", timeout=timeout)
        return self.child.before or ""

    def expect_output(self, substr: str, *, timeout: int = 30) -> None:
        self.child.expect(re.escape(substr), timeout=timeout)

    def expect_ait_prompt(self, user: str, host_fragment: str, *, timeout: int = 60) -> None:
        if self.native_prompt:
            self.child.expect(self.prompt_patterns, timeout=timeout)
            return
        pat = rf"AiT\s+{re.escape(user)}@{re.escape(host_fragment)}"
        self.child.expect(pat, timeout=timeout)

    def assert_no_native_prompt(self, text: str) -> None:
        """Fail when oh-my-zsh or bracket-style native prompts appear in output."""
        native_markers = ("➜", "->", " [%] ", " [~] ")
        for marker in native_markers:
            if marker in text:
                raise AssertionError(f"native prompt marker {marker!r} in output: {text!r}")

    def assert_ait_remote_prompt(self, user: str, host_fragment: str, text: str) -> None:
        if self.native_prompt:
            if user not in text and "➜" not in text and "#" not in text:
                raise AssertionError(f"missing native prompt markers for {user}@{host_fragment}: {text!r}")
            return
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
        before = self.child.before or ""
        after = self.child.after
        if not isinstance(after, str):
            after = ""
        return self.login_banner + before + after


def runner_env_from(
    cfg: TestConfig,
    ait_home: Path,
    *,
    ai_experimental: bool = False,
) -> dict[str, str]:
    env = os.environ.copy()
    env.update(cli_base_env(cfg, ait_home))
    env["AIT_HOME"] = str(ait_home)
    env["AIT_KEYRING_BACKEND"] = "file"
    env["AIT_KEYRING_DIR"] = str(ait_home / "keyring")
    env["AIT_KEYRING_FILE_PASSWORD"] = cfg.keyring_password
    env["LANG"] = cfg.lang
    env["LC_ALL"] = cfg.lang
    # Clear inherited shell export so smoke/SSH fixtures control the gate explicitly.
    env["AIT_AI_EXPERIMENTAL"] = "1" if ai_experimental else ""
    if cfg.features.insecure_ssh:
        env["AIT_SSH_INSECURE"] = "1"
    return env

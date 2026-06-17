"""Subprocess wrapper for non-interactive ait CLI invocations."""

from __future__ import annotations

import os
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Mapping, Sequence


@dataclass
class Completed:
    argv: list[str]
    returncode: int
    stdout: str
    stderr: str

    @property
    def output(self) -> str:
        return self.stdout + self.stderr


class AITRunner:
    """Run ait with isolated AIT_HOME and optional env."""

    def __init__(self, binary: Path, ait_home: Path, base_env: Mapping[str, str] | None = None):
        self.binary = Path(binary)
        self.ait_home = Path(ait_home)
        self.base_env = dict(base_env or {})

    def env(self, extra: Mapping[str, str] | None = None) -> dict[str, str]:
        env = os.environ.copy()
        env.update(self.base_env)
        env["AIT_HOME"] = str(self.ait_home)
        env["AIT_KEYRING_BACKEND"] = "file"
        env["AIT_KEYRING_DIR"] = str(self.ait_home / "keyring")
        if extra:
            env.update(extra)
        return env

    def run(
        self,
        args: Sequence[str],
        *,
        timeout: float = 30,
        input_text: str | None = None,
        extra_env: Mapping[str, str] | None = None,
        cwd: Path | None = None,
    ) -> Completed:
        argv = [str(self.binary), *args]
        proc = subprocess.run(
            argv,
            input=input_text,
            capture_output=True,
            text=True,
            timeout=timeout,
            env=self.env(extra_env),
            cwd=str(cwd) if cwd else None,
        )
        return Completed(argv=argv, returncode=proc.returncode, stdout=proc.stdout, stderr=proc.stderr)

    def run_ok(self, args: Sequence[str], **kwargs) -> Completed:
        result = self.run(args, **kwargs)
        assert_exit(0, result)
        return result


def assert_exit(code: int, proc: Completed) -> None:
    if proc.returncode != code:
        raise AssertionError(
            f"exit {proc.returncode} != {code}\n"
            f"argv: {' '.join(proc.argv)}\n"
            f"stdout:\n{proc.stdout}\n"
            f"stderr:\n{proc.stderr}"
        )


def has_tty() -> bool:
    return sys.stdin.isatty() and os.environ.get("AIT_E2E_FORCE_TTY") == "1"

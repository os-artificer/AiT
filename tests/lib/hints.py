"""errhint template assertions for CLI stderr."""

from __future__ import annotations

import re

from tests.lib.cli import Completed


def assert_no_cli_prefix(proc: Completed) -> None:
    """CLI stderr must not contain AiT: prefix (H0c)."""
    if "AiT:" in proc.stderr:
        raise AssertionError(f"CLI stderr must not contain AiT: prefix\n{proc.stderr}")


def assert_try_block(text: str, *commands: str) -> None:
    if "Try:" not in text:
        raise AssertionError(f"missing Try: block in\n{text}")
    for cmd in commands:
        needle = f"  {cmd}"
        if needle not in text and cmd not in text:
            raise AssertionError(f"missing try line {cmd!r} in\n{text}")


def assert_hint(
    proc: Completed,
    *,
    explanation_substr: str | None = None,
    try_commands: tuple[str, ...] = (),
    sentinel: str | None = None,
) -> None:
    assert_no_cli_prefix(proc)
    combined = proc.stderr
    if sentinel and sentinel not in combined:
        raise AssertionError(f"missing sentinel {sentinel!r} in stderr:\n{combined}")
    if explanation_substr and explanation_substr not in combined:
        raise AssertionError(f"missing explanation {explanation_substr!r} in stderr:\n{combined}")
    if try_commands:
        assert_try_block(combined, *try_commands)


def assert_unknown_command(proc: Completed, cmd: str) -> None:
    if proc.returncode == 0:
        raise AssertionError(f"expected non-zero exit for unknown command {cmd!r}")
    out = proc.output.lower()
    if "unknown" not in out and "not found" not in out:
        raise AssertionError(f"expected unknown/not found for {cmd!r}:\n{proc.output}")


def assert_contains(proc: Completed, *substrings: str, stream: str = "any") -> None:
    hay = proc.stdout if stream == "stdout" else proc.stderr if stream == "stderr" else proc.output
    for s in substrings:
        if s not in hay:
            raise AssertionError(f"missing {s!r} in {stream}:\n{hay}")


def assert_matches(proc: Completed, pattern: str, stream: str = "any") -> None:
    hay = proc.stdout if stream == "stdout" else proc.stderr if stream == "stderr" else proc.output
    if not re.search(pattern, hay, re.MULTILINE):
        raise AssertionError(f"pattern {pattern!r} not found in {stream}:\n{hay}")

"""version and --help smoke tests."""

from __future__ import annotations

import pytest

from tests.lib.hints import assert_contains, assert_unknown_command

pytestmark = pytest.mark.smoke

TOP_LEVEL_COMMANDS = [
    "shell",
    "session",
    "get",
    "put",
    "config",
    "version",
]

REMOVED_TOP_LEVEL = ["list", "load", "suspend", "models", "ll", "login", "ai"]


def test_version(runner):
    proc = runner.run_ok(["version"])
    assert_contains(proc, "Version:", stream="stdout")


def test_root_help(runner):
    proc = runner.run_ok(["--help"])
    assert_contains(proc, "Usage:", "aissh [command]")
    for cmd in TOP_LEVEL_COMMANDS:
        assert cmd in proc.stdout, f"missing {cmd} in root help"
    for gone in ("  list       ", "  load       ", "  suspend    ", "  login       "):
        assert gone not in proc.stdout, f"root help still lists removed command {gone!r}"
    # v4: Ops AI is the separate `ai` binary, not an aissh subcommand.
    assert "\n  ai " not in proc.stdout, "root help must not list ai as aissh subcommand"


@pytest.mark.parametrize("cmd", TOP_LEVEL_COMMANDS)
def test_subcommand_help(runner, cmd):
    proc = runner.run_ok([cmd, "--help"])
    assert_contains(proc, "Usage:")


@pytest.mark.parametrize("cmd", REMOVED_TOP_LEVEL)
def test_removed_top_level_unknown(runner, cmd):
    proc = runner.run([cmd])
    assert_unknown_command(proc, cmd)


def test_get_help_openssh_block(runner):
    proc = runner.run_ok(["get", "--help"])
    assert_contains(proc, "OpenSSH-compatible flags:")


def test_put_help_openssh_block(runner):
    proc = runner.run_ok(["put", "--help"])
    assert_contains(proc, "OpenSSH-compatible flags:")


def test_config_help_lists_subcommands(runner):
    proc = runner.run_ok(["config", "--help"])
    for sub in ("init", "view", "edit", "validate", "set", "get"):
        assert sub in proc.stdout, f"missing config {sub}"


def test_ai_help_lists_subcommands(ai_runner):
    proc = ai_runner.run_ok(["--help"])
    for sub in ("ask", "agent", "provider", "ollama", "model", "connect"):
        assert sub in proc.stdout, f"missing ai {sub}"

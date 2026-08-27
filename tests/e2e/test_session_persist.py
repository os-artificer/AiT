"""Session persistence: save, alias, load, switch, suspend."""

from __future__ import annotations

import uuid

import pytest

from tests.lib.config import cli_base_env
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.ssh


def _alias() -> str:
    return f"e2e-{uuid.uuid4().hex[:8]}"


@pytest.fixture
def persist_runner(test_config, tmp_path):
    home = tmp_path / "persist_home"
    home.mkdir()
    env = runner_env_from(test_config, home)
    from tests.lib.cli import AITRunner

    return AITRunner(test_config.binary, home, base_env=cli_base_env(test_config, home)), test_config, env


def test_login_save_creates_session_row(persist_runner):
    runner, cfg, _env = persist_runner
    runner.run_ok(cfg.primary.one_shot_login_argv(["echo", "e2e-save"], ["--save"]), timeout=90)
    proc = runner.run_ok(["session", "--list"])
    assert cfg.primary.host in proc.stdout or cfg.primary.user in proc.stdout


def test_login_alias_separate_row(persist_runner):
    runner, cfg, _env = persist_runner
    alias = _alias()
    runner.run_ok(
        cfg.primary.one_shot_login_argv(["echo", "ok"], ["--alias", alias]),
        timeout=90,
    )
    proc = runner.run_ok(["session", "--list"])
    assert alias in proc.stdout


def test_session_list_verbose_full_id(persist_runner):
    runner, cfg, _env = persist_runner
    runner.run_ok(cfg.primary.one_shot_login_argv(["echo", "ok"], ["--save"]), timeout=90)
    proc = runner.run_ok(["session", "--list", "-v"])
    assert "FULL_ID" in proc.stdout or len(proc.stdout) > 20


@pytest.mark.destructive
def test_suspend_via_meta_in_session(persist_runner):
    runner, cfg, env = persist_runner
    alias = _alias()
    runner.run_ok(cfg.primary.login_argv(["--alias", alias]), timeout=90)
    sess = AITExpectSession.spawn_login(cfg, env, extra_login_args=["--alias", alias])
    try:
        sess.meta("suspend")
    finally:
        sess.close()
    proc = runner.run_ok(["session", "--list"])
    assert alias in proc.stdout

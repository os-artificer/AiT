"""pytest fixtures for AiT E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.cli import AITRunner
from tests.lib.config import TestConfig, config_skip_reason, load_config
from tests.lib.expect import AITExpectSession, runner_env_from


@pytest.fixture(scope="session")
def test_config() -> TestConfig:
    return load_config()


@pytest.fixture(scope="session")
def ait_bin(test_config: TestConfig) -> str:
    if not test_config.binary.is_file():
        pytest.skip(f"Binary not found: {test_config.binary}; run make build")
    return str(test_config.binary)


@pytest.fixture
def ait_home(tmp_path) -> str:
    home = tmp_path / "ait_home"
    home.mkdir()
    return str(home)


@pytest.fixture
def runner(test_config: TestConfig, ait_bin: str, ait_home: str) -> AITRunner:
    base = {
        "LANG": test_config.lang,
        "LC_ALL": test_config.lang,
        "AIT_KEYRING_FILE_PASSWORD": test_config.keyring_password,
    }
    if test_config.features.insecure_ssh:
        base["AIT_SSH_INSECURE"] = "1"
    return AITRunner(ait_bin, ait_home, base_env=base)


@pytest.fixture
def expect_env(test_config: TestConfig, ait_home: str) -> dict[str, str]:
    return runner_env_from(test_config, __import__("pathlib").Path(ait_home))


def pytest_collection_modifyitems(config, items):
    cfg = load_config()
    for item in items:
        if item.get_closest_marker("ssh") and not cfg.is_ssh_configured():
            item.add_marker(pytest.mark.skip(reason=config_skip_reason(cfg) or "SSH not configured"))
        if item.get_closest_marker("ai") and not cfg.is_ai_configured():
            item.add_marker(pytest.mark.skip(reason="AI not configured (keys or require_ollama)"))
        if item.get_closest_marker("tmux") and not cfg.features.tmux:
            item.add_marker(pytest.mark.skip(reason="features.tmux disabled in config"))
        if item.get_closest_marker("zmodem") and not cfg.features.zmodem_chain:
            item.add_marker(pytest.mark.skip(reason="features.zmodem_chain disabled in config"))


@pytest.fixture(scope="module")
def e2e_ssh_session(test_config: TestConfig, tmp_path_factory):
    """Shared login session for a module (one dial per file)."""
    if not test_config.is_ssh_configured():
        pytest.skip(config_skip_reason(test_config))
    home = tmp_path_factory.mktemp("ssh_module_home")
    env = runner_env_from(test_config, home)
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()

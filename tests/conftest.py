"""pytest fixtures for AiT E2E tests."""

from __future__ import annotations

import pytest

from tests.lib.cli import AITRunner
from tests.lib.config import TestConfig, cli_base_env, config_skip_reason, load_config
from tests.lib.envflags import AI_EXPERIMENTAL_ENV, running_on_ci, use_testenv_mock
from tests.lib.expect import AITExpectSession, runner_env_from
from tests.lib.mockenv import apply_mock_target, start_mock


@pytest.fixture(scope="session")
def testenv_mock():
    """Session-wide in-repo testenv-mock when AIT_TEST_USE_MOCK=1 (CI full e2e)."""
    if not use_testenv_mock():
        yield None
        return
    env = start_mock()
    try:
        yield env
    finally:
        env.close()


@pytest.fixture(scope="session")
def test_config(testenv_mock) -> TestConfig:
    cfg = load_config()
    if testenv_mock is not None:
        apply_mock_target(cfg, testenv_mock.ssh)
    return cfg


@pytest.fixture(scope="session")
def ait_bin(test_config: TestConfig) -> str:
    if not test_config.binary.is_file():
        pytest.skip(f"Binary not found: {test_config.binary}; run make build")
    return str(test_config.binary)


@pytest.fixture(scope="session")
def ai_bin(test_config: TestConfig) -> str:
    """Sibling `bin/ai` next to the configured aissh binary (v4 Ops AI CLI)."""
    path = test_config.binary.parent / "ai"
    if not path.is_file():
        pytest.skip(f"Binary not found: {path}; run make build")
    return str(path)


@pytest.fixture(scope="session")
def toolbox_bin(test_config: TestConfig) -> str:
    """Sibling `bin/ait` for toolbox REPL / `ait agent` Loop tests."""
    path = test_config.binary.parent / "ait"
    if not path.is_file():
        pytest.skip(f"Binary not found: {path}; run make build")
    return str(path)


@pytest.fixture
def ait_home(tmp_path) -> str:
    home = tmp_path / "ait_home"
    home.mkdir()
    return str(home)


@pytest.fixture
def runner(test_config: TestConfig, ait_bin: str, ait_home: str) -> AITRunner:
    return AITRunner(ait_bin, ait_home, base_env=cli_base_env(test_config, ait_home))


@pytest.fixture
def ai_runner(runner: AITRunner, ai_bin: str) -> AITRunner:
    """`bin/ai` with experimental AI enabled (local AI smoke / @ai tests).

    Config remains on aissh (`runner`); share AIT_HOME so keys and experimental
    YAML settings apply to both binaries.
    """
    return AITRunner(
        ai_bin,
        runner.ait_home,
        base_env={**runner.base_env, **AI_EXPERIMENTAL_ENV},
    )


@pytest.fixture
def expect_env(test_config: TestConfig, ait_home: str) -> dict[str, str]:
    return runner_env_from(test_config, __import__("pathlib").Path(ait_home))


def pytest_collection_modifyitems(config, items):
    cfg = load_config()
    on_ci = running_on_ci()
    mock = use_testenv_mock()
    for item in items:
        if item.get_closest_marker("ssh"):
            if on_ci and not mock:
                item.add_marker(
                    pytest.mark.skip(reason="SSH/host e2e runs only locally (not on CI)"),
                )
            elif not mock and not cfg.is_ssh_configured():
                item.add_marker(
                    pytest.mark.skip(reason=config_skip_reason(cfg) or "SSH not configured"),
                )
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
        pytest.skip(config_skip_reason(test_config) or "SSH not configured")
    home = tmp_path_factory.mktemp("ssh_module_home")
    env = runner_env_from(test_config, home)
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()

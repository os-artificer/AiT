"""Load E2E test configuration from YAML and environment variables."""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Sequence

import yaml

TESTS_DIR = Path(__file__).resolve().parent.parent
DEFAULT_CONFIG_PATH = TESTS_DIR / "config.local.yaml"
EXAMPLE_CONFIG_PATH = TESTS_DIR / "config.example.yaml"


@dataclass
class SSHTarget:
    host: str = ""
    port: int = 22
    user: str = "root"
    identity: str = ""
    password: str = ""

    @property
    def destination(self) -> str:
        if self.user and self.host:
            return f"{self.user}@{self.host}"
        return self.host

    def connect_flags(self) -> list[str]:
        """OpenSSH-style flags shared by connect, get, and put."""
        argv: list[str] = []
        if self.port and self.port != 22:
            argv.extend(["-p", str(self.port)])
        if self.identity:
            argv.extend(["-i", os.path.expanduser(self.identity)])
        if self.password:
            argv.extend(["--password", self.password])
        return argv

    def login_argv(self, extra: list[str] | None = None) -> list[str]:
        argv: list[str] = self.connect_flags()
        if extra:
            argv.extend(extra)
        argv.append(self.destination)
        return argv

    def one_shot_login_argv(
        self,
        remote_cmd: Sequence[str],
        extra: list[str] | None = None,
    ) -> list[str]:
        argv = self.login_argv(extra)
        argv.extend(remote_cmd)
        return argv


@dataclass
class Features:
    insecure_ssh: bool = False
    nested_ssh: bool = False
    tmux: bool = False
    lrzsz: bool = False
    zmodem_chain: bool = False


@dataclass
class AIConfig:
    deepseek_key: str = ""
    minimax_key: str = ""
    require_ollama: bool = False

    @property
    def has_cloud_key(self) -> bool:
        return bool(self.deepseek_key or self.minimax_key)


@dataclass
class TestConfig:
    binary: Path = TESTS_DIR.parent / "bin" / "aissh"
    primary: SSHTarget = field(default_factory=SSHTarget)
    inner: SSHTarget = field(default_factory=SSHTarget)
    jump_hosts: list[str] = field(default_factory=list)
    ai: AIConfig = field(default_factory=AIConfig)
    features: Features = field(default_factory=Features)
    keyring_password: str = "ait-e2e-test"
    lang: str = "en_US.UTF-8"
    config_path: Path | None = None
    loaded: bool = False
    using_mock: bool = False

    def is_ssh_configured(self) -> bool:
        return bool(self.primary.host)

    def is_ai_configured(self) -> bool:
        if self.ai.require_ollama:
            return True
        return self.ai.has_cloud_key


def _merge_ssh(data: dict[str, Any] | None) -> SSHTarget:
    data = data or {}
    return SSHTarget(
        host=str(data.get("host", "")).strip(),
        port=int(data.get("port", 22) or 22),
        user=str(data.get("user", "root")).strip() or "root",
        identity=str(data.get("identity", "")).strip(),
        password=str(data.get("password", "")).strip(),
    )


def load_config(path: Path | None = None) -> TestConfig:
    """Load config from YAML; apply AIT_TEST_* environment overrides."""
    cfg_path = Path(path or os.environ.get("AIT_TEST_CONFIG", DEFAULT_CONFIG_PATH))
    raw: dict[str, Any] = {}
    if cfg_path.is_file():
        with cfg_path.open(encoding="utf-8") as f:
            raw = yaml.safe_load(f) or {}

    ait_block = raw.get("ait") or {}
    ssh_block = raw.get("ssh") or {}
    ai_block = raw.get("ai") or {}
    feat_block = raw.get("features") or {}
    env_block = raw.get("env") or {}

    binary = Path(str(ait_block.get("binary", "../bin/aissh")))
    if not binary.is_absolute():
        binary = (TESTS_DIR / binary).resolve()

    jump = ssh_block.get("jump") or {}
    jump_hosts = list(jump.get("hosts") or [])

    cfg = TestConfig(
        binary=binary,
        primary=_merge_ssh(ssh_block.get("primary")),
        inner=_merge_ssh(ssh_block.get("inner")),
        jump_hosts=[str(h).strip() for h in jump_hosts if str(h).strip()],
        ai=AIConfig(
            deepseek_key=str(ai_block.get("deepseek_key", "")).strip(),
            minimax_key=str(ai_block.get("minimax_key", "")).strip(),
            require_ollama=bool(ai_block.get("require_ollama", False)),
        ),
        features=Features(
            insecure_ssh=bool(feat_block.get("insecure_ssh", False)),
            nested_ssh=bool(feat_block.get("nested_ssh", False)),
            tmux=bool(feat_block.get("tmux", False)),
            lrzsz=bool(feat_block.get("lrzsz", False)),
            zmodem_chain=bool(feat_block.get("zmodem_chain", False)),
        ),
        keyring_password=str(
            env_block.get("keyring_password")
            or os.environ.get("AIT_KEYRING_FILE_PASSWORD", "ait-e2e-test")
        ),
        lang=str(env_block.get("lang", "en_US.UTF-8")),
        config_path=cfg_path if cfg_path.is_file() else None,
        loaded=cfg_path.is_file(),
    )

    if os.environ.get("AIT_TEST_SSH_HOST"):
        cfg.primary.host = os.environ["AIT_TEST_SSH_HOST"].strip()
    if os.environ.get("AIT_TEST_SSH_PORT"):
        cfg.primary.port = int(os.environ["AIT_TEST_SSH_PORT"])
    if os.environ.get("AIT_TEST_SSH_INNER_HOST"):
        cfg.inner.host = os.environ["AIT_TEST_SSH_INNER_HOST"].strip()
    if os.environ.get("AIT_TEST_DEEPSEEK_KEY"):
        cfg.ai.deepseek_key = os.environ["AIT_TEST_DEEPSEEK_KEY"].strip()

    return cfg


def cli_base_env(cfg: TestConfig, ait_home: Path | str) -> dict[str, str]:
    """Env overlay for AITRunner / pexpect against the configured SSH target."""
    env = {
        "LANG": cfg.lang,
        "LC_ALL": cfg.lang,
        "AIT_KEYRING_FILE_PASSWORD": cfg.keyring_password,
        "AIT_AI_EXPERIMENTAL": "",
    }
    if cfg.features.insecure_ssh:
        env["AIT_SSH_INSECURE"] = "1"
    if cfg.using_mock:
        env["SSH_AUTH_SOCK"] = ""
        env["HOME"] = str(ait_home)
    return env


def config_skip_reason(cfg: TestConfig) -> str:
    if not cfg.loaded:
        return (
            f"Missing {DEFAULT_CONFIG_PATH}; copy tests/config.example.yaml "
            "to tests/config.local.yaml and configure SSH/AI targets."
        )
    if not cfg.binary.is_file():
        return f"Binary not found: {cfg.binary}; run make build"
    return ""

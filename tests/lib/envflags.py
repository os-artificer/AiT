"""Environment helpers for E2E (CI vs local SSH)."""

from __future__ import annotations

import os

# Enable experimental AI for subprocesses that exercise ai ask/agent/provider paths.
AI_EXPERIMENTAL_ENV = {"AIT_AI_EXPERIMENTAL": "1"}


def running_on_ci() -> bool:
    """True on GitHub Actions / generic CI — never dial real SSH hosts."""
    if os.environ.get("GITHUB_ACTIONS", "").lower() in ("1", "true", "yes"):
        return True
    return os.environ.get("CI", "").lower() in ("1", "true", "yes")


def use_testenv_mock() -> bool:
    """True when CI/local e2e should inject in-repo testenv-mock as ssh.primary."""
    return os.environ.get("AIT_TEST_USE_MOCK", "").lower() in ("1", "true", "yes")

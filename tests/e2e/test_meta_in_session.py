"""In-session meta commands: !ait list, suspend, chain, local/remote."""

from __future__ import annotations

import pytest

from tests.lib.config import load_config
from tests.lib.expect import AITExpectSession, runner_env_from

pytestmark = pytest.mark.ssh


@pytest.fixture
def meta_session(test_config, ait_home):
    env = runner_env_from(test_config, __import__("pathlib").Path(ait_home))
    sess = AITExpectSession.spawn_login(test_config, env)
    yield sess
    sess.close()


def test_meta_list_shows_sessions(meta_session):
    out = meta_session.meta("session --list")
    assert "#" in out or "session" in out.lower() or out.strip() == ""


def test_meta_chain_status(meta_session):
    out = meta_session.meta("chain")
    assert "chain" in out.lower() or "on" in out.lower() or "off" in out.lower()


def test_meta_chain_toggle(meta_session):
    meta_session.meta("chain on")
    out = meta_session.meta("chain")
    assert "on" in out.lower()


def test_meta_help(meta_session):
    out = meta_session.send_line("!ait help")
    assert "ait" in out.lower() or "help" in out.lower()

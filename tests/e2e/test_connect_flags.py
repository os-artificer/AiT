"""SSHTarget.connect_flags / login_argv include port and --password (MS22)."""

from __future__ import annotations

import pytest

from tests.lib.config import SSHTarget

pytestmark = pytest.mark.smoke


def test_login_argv_includes_password_and_port():
    tgt = SSHTarget(host="127.0.0.1", port=2222, user="dev", password="devpass")
    argv = tgt.login_argv()
    assert argv == ["-p", "2222", "--password", "devpass", "dev@127.0.0.1"]
    assert tgt.connect_flags() == ["-p", "2222", "--password", "devpass"]


def test_login_argv_omits_password_when_empty():
    tgt = SSHTarget(host="host.example.com", port=22, user="dev", identity="~/.ssh/id_ed25519")
    argv = tgt.login_argv(["--save"])
    assert "--password" not in argv
    assert argv[-1] == "dev@host.example.com"
    assert "--save" in argv

"""Remote Ops AI via `ai connect` (v4). Legacy !! / !ait ai are not on aissh."""

from __future__ import annotations

import pytest

pytestmark = [pytest.mark.ssh, pytest.mark.ai]


def test_ai_connect_help(ai_runner):
    proc = ai_runner.run_ok(["connect", "--help"])
    assert "Usage:" in proc.stdout
    assert "connect" in proc.stdout.lower()


@pytest.mark.skip(reason="v4 remote Agent uses ai connect; !! / !ait ai removed from aissh")
def test_remote_ai_legacy_meta_removed():
    pass

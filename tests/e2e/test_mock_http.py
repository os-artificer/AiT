"""HTTP mock user-error and success probes (MS10–MS12) via in-repo testenv-mock."""

from __future__ import annotations

import json
import urllib.error
import urllib.request

import pytest

from tests.lib.mockenv import start_mock

pytestmark = [pytest.mark.smoke, pytest.mark.mock]


@pytest.fixture(scope="module")
def mock_env():
    env = start_mock()
    yield env
    env.close()


def test_nominatim_ok(mock_env):
    url = mock_env.http["reverse_url"] + "?lat=31.24&lon=121.49"
    with urllib.request.urlopen(url, timeout=5) as resp:
        body = json.loads(resp.read().decode())
    assert "外滩" in body["display_name"]


def test_ollama_tags_ok(mock_env):
    with urllib.request.urlopen(mock_env.http["tags_url"], timeout=5) as resp:
        body = json.loads(resp.read().decode())
    assert body["models"][0]["name"]


def test_openai_chat_ok(mock_env):
    req = urllib.request.Request(
        mock_env.http["base_url"] + "/v1/chat/completions",
        data=b"{}",
        method="POST",
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=5) as resp:
        body = json.loads(resp.read().decode())
    assert body["choices"][0]["message"]["content"]


def test_http_unauthorized_chat():
    env = start_mock(http_scenario="unauthorized")
    try:
        req = urllib.request.Request(
            env.http["base_url"] + "/v1/chat/completions",
            data=b"{}",
            method="POST",
            headers={"Content-Type": "application/json"},
        )
        with pytest.raises(urllib.error.HTTPError) as caught:
            urllib.request.urlopen(req, timeout=5)
        assert caught.value.code == 401
    finally:
        env.close()


def test_http_server_error():
    env = start_mock(http_scenario="server_error")
    try:
        with pytest.raises(urllib.error.HTTPError) as caught:
            urllib.request.urlopen(env.http["tags_url"], timeout=5)
        assert caught.value.code == 500
    finally:
        env.close()

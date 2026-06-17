"""Temporary files for get/put and transfer probes."""

from __future__ import annotations

import os
import uuid
from pathlib import Path


def probe_name(prefix: str = "ait-e2e") -> str:
    return f"{prefix}-{uuid.uuid4().hex[:8]}.txt"


def write_probe(path: Path, content: str = "ait-e2e-probe\n") -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    return path


def remote_probe_path(name: str, directory: str = "/tmp") -> str:
    return f"{directory.rstrip('/')}/{name}"


def local_probe_in(tmp: Path, name: str | None = None) -> Path:
    return write_probe(tmp / (name or probe_name()))

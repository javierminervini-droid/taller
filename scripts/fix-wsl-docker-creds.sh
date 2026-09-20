#!/usr/bin/env bash
set -euo pipefail

BIN="/mnt/c/Users/jacka/AppData/Local/Programs/DockerDesktop/resources/bin"
PROFILE="${HOME}/.bashrc"
MARKER="# Docker Desktop CLI (Windows)"

if [[ ! -x "${BIN}/docker-credential-desktop.exe" ]]; then
  echo "Helper not found at ${BIN}/docker-credential-desktop.exe" >&2
  exit 1
fi

if ! grep -qF "${MARKER}" "${PROFILE}" 2>/dev/null; then
  {
    echo ""
    echo "${MARKER}"
    echo "export PATH=\"\$PATH:${BIN}\""
  } >> "${PROFILE}"
  echo "Added Docker bin to ~/.bashrc"
else
  echo "bashrc already has Docker PATH"
fi

mkdir -p "${HOME}/.docker"
python3 - <<'PY'
import json
from pathlib import Path
p = Path.home() / ".docker" / "config.json"
cfg = json.loads(p.read_text()) if p.exists() else {}
cfg.pop("credsStore", None)
cfg.pop("credStore", None)
p.write_text(json.dumps(cfg, indent=2) + "\n")
print("Updated", p, "->", cfg)
PY

export PATH="${PATH}:${BIN}"
cd /mnt/c/Users/jacka/Projects/taller-gestion
echo "Testing docker compose build metadata..."
docker compose build --pull 2>&1 | tail -30

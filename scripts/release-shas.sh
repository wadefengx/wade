#!/usr/bin/env bash
set -euo pipefail

repo="wadefengx/wade"
api="https://api.github.com/repos/$repo/releases/latest"
mode="${1:-}"

if [[ "$mode" != "" && "$mode" != "--update" ]]; then
  echo "usage: $0 [--update]" >&2
  exit 2
fi

release_json="$(mktemp)"
trap 'rm -f "$release_json"' EXIT
curl --fail --silent --show-error --location "$api" -o "$release_json"

python3 - "$mode" "$(dirname "$0")" "$release_json" <<'PY'
import json
import pathlib
import re
import sys

mode, scripts_dir, release_json = sys.argv[1:]
with open(release_json) as file:
    release = json.load(file)
version = release["tag_name"].removeprefix("v")
assets = {
    asset["name"]: asset["digest"].removeprefix("sha256:")
    for asset in release["assets"]
    if asset.get("digest", "").startswith("sha256:")
}
names = [
    "wade-darwin-arm64.tar.gz",
    "wade-darwin-amd64.tar.gz",
    "wade-linux-amd64.tar.gz",
    "wade-linux-arm64.tar.gz",
    "wade-windows-amd64.zip",
]

print(f"version: {version}")
for name in names:
    digest = assets.get(name)
    print(f"{name}: {digest or 'TODO: fill from release'}")

if mode != "--update":
    raise SystemExit

formula = pathlib.Path(scripts_dir) / "wade.rb"
manifest = pathlib.Path(scripts_dir) / "wade.json"
formula_text = formula.read_text()
manifest_text = manifest.read_text()

formula_text = re.sub(r'  version "[^"]+"', f'  version "{version}"', formula_text)
formula_text = re.sub(
    r'https://github\.com/wadefengx/wade/releases/download/v[^/"]+/',
    f'https://github.com/wadefengx/wade/releases/download/v{version}/',
    formula_text,
)
for name, digest in assets.items():
    if not name.endswith(".tar.gz"):
        continue
    pattern = rf'(url "#\{{BASE_URL\}}{re.escape(name)}"\n\s+sha256 )"[^"]+"'
    formula_text = re.sub(pattern, rf'\g<1>"{digest}"', formula_text)

manifest_text = re.sub(r'"version": "[^"]+"', f'"version": "{version}"', manifest_text)
manifest_text = re.sub(
    r'("url": "https://github\.com/wadefengx/wade/releases/download/)v[^/]+(/wade-windows-amd64\.zip")',
    rf'\g<1>v{version}\g<2>',
    manifest_text,
    count=1,
)
windows_digest = assets.get("wade-windows-amd64.zip")
if windows_digest:
    manifest_text = re.sub(
        r'("url": "[^"]*wade-windows-amd64\.zip",\n\s+"hash": )"[^"]+"',
        rf'\g<1>"{windows_digest}"',
        manifest_text,
    )

formula.write_text(formula_text)
manifest.write_text(manifest_text)
PY

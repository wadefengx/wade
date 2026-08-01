#!/usr/bin/env bash
# Wade AI Harness — 质量门禁
# 用法: .ai/harness/verify.sh [--skip-race]
# 输出: PASS/FAIL 计数, 非零退出码表示有失败项
set -uo pipefail

cd "$(dirname "$0")/../.."
PASS=0
FAIL=0
FAILED_ITEMS=()

check() {
  local name="$1" cmd="$2"
  echo -n "▶ $name ... "
  if eval "$cmd" >/tmp/wade-verify.log 2>&1; then
    echo "PASS"
    PASS=$((PASS + 1))
  else
    echo "FAIL"
    FAIL=$((FAIL + 1))
    FAILED_ITEMS+=("$name")
    tail -5 /tmp/wade-verify.log | sed 's/^/    /'
  fi
}

echo "═══════════ wade quality harness ═══════════"

check "gofmt (no output expected)"   "test -z \"\$(gofmt -l .)\""
check "go vet"                       "go vet ./..."
check "go build"                     "go build ./..."
check "go test ./..."                "go test ./..."
check "cross-compile windows"        "GOOS=windows go build ./..."
check "cross-compile linux/arm64"    "GOOS=linux GOARCH=arm64 go build ./..."
check "shellcheck install.sh"        "bash -n scripts/install.sh"
check "release-shas.sh syntax"       "bash -n scripts/release-shas.sh"
check "scoop manifest JSON valid"    "python3 -m json.tool scripts/wade.json >/dev/null"
check "docs HTML well-formed"        "python3 -c \"from html.parser import HTMLParser; HTMLParser().feed(open('docs/index.html').read())\""

echo "════════════════════════════════════════════"
echo "RESULT: $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
  echo "FAILED: ${FAILED_ITEMS[*]}"
  exit 1
fi
echo "ALL GREEN ✅"
exit 0

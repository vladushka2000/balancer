#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-$ROOT/pii-src.zip}"

cd "$ROOT"

rm -f "$OUT"

zip -r "$OUT" . \
  -x '*.git*' \
  -x 'target/*' \
  -x 'build/*' \
  -x '.idea/*' \
  -x '__pycache__/*' \
  -x '*.pyc' \
  -x 'node_modules/*' \
  -x '.venv/*' \
  -x 'dist/*' \
  -x 'out/*' \
  -x 'bin/*' \
  -x 'obj/*' \
  -x '*.exe' \
  -x '*.test' \
  -x '*.out' \
  -x '*.pdb' \
  -x 'logs/*' \
  -x '.cursor/*' \
  -x '*.zip'

echo "packed: $OUT"
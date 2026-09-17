#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if command -v npm >/dev/null 2>&1; then
  (cd web && npm run build)
fi

mkdir -p pkg/server/dist
# Refresh embedded UI from web/dist (air-gapped binary has no runtime asset server).
if command -v rsync >/dev/null 2>&1; then
  rsync -a --delete web/dist/ pkg/server/dist/
else
  rm -rf pkg/server/dist
  cp -R web/dist pkg/server/dist
fi

OUT="${OUT:-mesosphere}"
if [[ "${GOOS:-$(go env GOOS)}" == "windows" && "$OUT" != *.exe ]]; then
  OUT="${OUT}.exe"
fi
go build -o "$OUT" .
echo "built $OUT"

#!/usr/bin/env bash
# FST 生产构建（build.bat 的 bash 对应）
set -euo pipefail
cd "$(dirname "$0")"

CHOICE="${1:-}"
if [[ -z "$CHOICE" ]]; then
  echo "============================================================"
  echo "  [FST] Production Build Tool"
  echo "============================================================"
  echo
  echo "  [1] Embedded - Single binary"
  echo "  [2] External - Separate frontend assets"
  echo
  read -r -p "Select [1 or 2] (default 1): " CHOICE
fi
CHOICE="${CHOICE:-1}"

BMODE=embedded
if [[ "$CHOICE" == "2" ]]; then
  BMODE=external
fi

echo
echo "[1/4] Cleaning old build artifacts..."
rm -rf build dist
mkdir -p build dist
echo "building..." > dist/index.html

echo
echo "[2/4] Building frontend (pnpm build)..."
(
  cd frontend
  export VITE_BUILD_MODE="$BMODE"
  pnpm build
)

if [[ -f embedded_assets/share/cd2d5f2a2f5be-y.jpg ]]; then
  cp -f embedded_assets/share/cd2d5f2a2f5be-y.jpg dist/cd2d5f2a2f5be-y.jpg
fi
echo "Frontend assets ready in ./dist/"

echo
echo "[3/4] Cross-compiling Go backend (Mode: $BMODE)..."
export CGO_ENABLED=0

build_target() {
  local GOOS="$1" GOARCH="$2" EXT="$3" LABEL="$4"
  local OUTDIR="build/${GOOS}_${GOARCH}"
  echo "  - Building: ${LABEL}..."
  mkdir -p "$OUTDIR"
  GOOS="$GOOS" GOARCH="$GOARCH" go build -tags embedded \
    -ldflags "-X main.BuildMode=${BMODE} -s -w" \
    -o "${OUTDIR}/fst${EXT}" .
  if [[ "$BMODE" == "external" ]]; then
    mkdir -p "${OUTDIR}/dist"
    cp -a dist/. "${OUTDIR}/dist/"
  fi
  if [[ -f .env.example ]]; then
    cp -f .env.example "${OUTDIR}/.env.example"
  fi
}

build_target windows amd64 ".exe" "Windows x64"
build_target linux amd64 "" "Linux x64"

echo
echo "[4/4] Build complete!"
echo "============================================================"
echo "  Output: ./build/"
echo "  Mode:   [${BMODE}]"
echo "============================================================"

#!/usr/bin/env bash
# FST 开发启动（dev.bat 的 bash 对应）：扫插件 + Swagger + 前后端
set -euo pipefail
cd "$(dirname "$0")"

export GO_ENV=development

echo
echo "============================================================"
echo "  FST - Dev Mode (Auto Docs)"
echo "============================================================"
echo

echo "[1/3] Scanning plugins + Generating Swagger..."
go run backend/app/plugins/gen_plugins.go

(
  cd backend
  if ! swag init -g ../main.go -o docs --parseDependency --parseInternal --quiet; then
    echo "[WARN] Swagger generation failed. Install swag:"
    echo "  go install github.com/swaggo/swag/cmd/swag@latest"
  fi
)

echo
echo "[2/3] Starting frontend dev server..."
(
  cd frontend
  pnpm dev
) &
FRONTEND_PID=$!

cleanup() {
  if kill -0 "$FRONTEND_PID" 2>/dev/null; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

echo
echo "[3/3] Starting backend server (root main.go)..."
echo "============================================================"
echo
echo "  Backend:  http://localhost:8080"
echo "  Frontend: http://localhost:9980"
echo "  Scalar:   http://localhost:8080/scalar"
echo
echo "  Press Ctrl+C to stop"
echo "============================================================"
echo

go run ./main.go

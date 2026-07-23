#!/usr/bin/env bash
# FST 测试脚本（test.bat 的 bash 对应）
set -euo pipefail
cd "$(dirname "$0")"

ERROR_COUNT=0
RUN_BUILD=0
if [[ "${1:-}" == "--build" ]]; then
  RUN_BUILD=1
fi

echo "========================================"
echo "  FST 测试（go test / go vet / pnpm lint）"
echo "========================================"
echo

echo "[1/4] go test ./backend/... -count=1"
if ! go test ./backend/... -count=1; then
  echo "✗ go test 失败"
  ERROR_COUNT=$((ERROR_COUNT + 1))
else
  echo "✓ go test 通过"
fi

echo
echo "[2/4] go vet ./backend/..."
if ! go vet ./backend/...; then
  echo "✗ go vet 失败"
  ERROR_COUNT=$((ERROR_COUNT + 1))
else
  echo "✓ go vet 通过"
fi

echo
echo "[3/4] 前端 pnpm lint"
if [[ -f frontend/package.json ]]; then
  pushd frontend >/dev/null
  if [[ ! -d node_modules ]]; then
    echo "! 安装前端依赖..."
    pnpm install || ERROR_COUNT=$((ERROR_COUNT + 1))
  fi
  if [[ $ERROR_COUNT -eq 0 ]] || [[ -d node_modules ]]; then
    if ! pnpm lint; then
      echo "✗ pnpm lint 失败"
      ERROR_COUNT=$((ERROR_COUNT + 1))
    else
      echo "✓ pnpm lint 通过"
    fi
  fi
  popd >/dev/null
else
  echo "! 无 frontend 目录，跳过 lint"
fi

echo
if [[ "$RUN_BUILD" -eq 1 ]]; then
  echo "[4/4] 可选编译冒烟（--build）"
  mkdir -p build
  if ! go build -o build/test_backend ./; then
    echo "✗ 后端编译失败"
    ERROR_COUNT=$((ERROR_COUNT + 1))
  else
    echo "✓ 后端编译成功"
    rm -f build/test_backend
  fi
  if [[ -f frontend/package.json ]]; then
    pushd frontend >/dev/null
    if ! pnpm build; then
      echo "✗ 前端编译失败"
      ERROR_COUNT=$((ERROR_COUNT + 1))
    else
      echo "✓ 前端编译成功"
    fi
    popd >/dev/null
  fi
else
  echo "[4/4] 跳过编译冒烟（需要时请：./test.sh --build）"
fi

echo
echo "========================================"
if [[ "$ERROR_COUNT" -eq 0 ]]; then
  echo "✓ 全部通过"
else
  echo "✗ 发现 ${ERROR_COUNT} 个失败"
fi
echo "========================================"
exit "$ERROR_COUNT"

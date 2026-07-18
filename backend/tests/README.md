# backend/tests

> **最后更新**：2026-07-17

## 说明

本目录用于放「测试相关说明 / 集成用例索引」，**不替代**各包旁的 `*_test.go`。

Go 语言规则：

- `package xxx` 的白盒单元测试文件必须写在该包目录下（例如 `backend/app/services/foo_test.go`）。
- 强行挪到别处后，`go test ./backend/app/services` 将无法编译同包未导出符号，也不符合官方惯例。

## 推荐分工

| 内容 | 位置 |
|------|------|
| 单元测试 | 各包目录 `*_test.go` |
| 运维脚本、本地集成验证 | 项目根 [`tools/`](../../tools/) |
| 本说明 | `backend/tests/README.md` |

## 常用命令

```bash
# 后端主要单元测试
go test ./backend/app/services/ ./backend/app/models/ ./backend/utils/ ./backend/pkg/... ./backend/app/plugins/... -count=1

# 自动任务包（当前无 *_test.go，仅编译检查）
go build ./backend/internal/task/
```

自动任务管理器核心在 [`../internal/task`](../internal/task)，管理 API 在 `app/controllers/admin/auto_job_controller.go`。

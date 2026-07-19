# backend/tests

> **最后更新**：2026-07-19

## 说明

本目录用于放「测试相关说明 / 集成用例索引」，**不替代**各包旁的 `*_test.go`。

Go 语言规则：

- `package xxx` 的白盒单元测试文件必须写在该包目录下（例如 `backend/app/services/foo_test.go`）。
- 强行挪到别处后，`go test ./backend/app/services` 将无法编译同包未导出符号，也不符合官方惯例。

## 推荐分工

| 内容 | 位置 |
|------|------|
| 单元测试 / httptest 控制器测试 | 各包目录 `*_test.go` |
| SQLite 临时库测试工具 | [`internal/testutil`](../internal/testutil)（`testutil.SetupSQLite`） |
| 运维脚本、本地集成验证 | 项目根 [`tools/`](../../tools/) |
| 本说明 | `backend/tests/README.md` |

## 常用命令

```bash
# 全量后端测试（23 个包均已覆盖，含 controllers/models/services/middleware/pkg/utils 等）
go test ./backend/... -count=1

# 仅跑 SQLite 适配相关测试（CRUD / 索引 / 迁移 / 关键 DML 方言转换）
go test ./backend/pkg/db/... -count=1 -run SQLite
```

自动任务管理器核心在 [`../internal/task`](../internal/task)，管理 API 在 `app/controllers/admin/auto_job_controller.go`，已有 `internal/task` 包测试覆盖。

# backend/tests

> **最后更新**：2026-07-23

本目录只放测试说明，**不替代**各包旁的 `*_test.go`。

| 内容 | 位置 |
|------|------|
| 单元 / httptest | 各包 `*_test.go` |
| SQLite 临时库工具 | `internal/testutil` |
| 运维脚本 | 根目录 `tools/` |

```bash
go test ./backend/... -count=1

# pkg/db：表/索引探测 + 跨包锁行/聚合冒烟
go test ./backend/pkg/db/... -count=1

# Postgres 真机（需 FST_PG_DSN）
go test -tags integration ./backend/pkg/db/ -run Postgres -count=1
```

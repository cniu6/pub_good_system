# normalize_collations

统一 MySQL 数据库、表、字符串列的字符集与排序规则为 `utf8mb4`，自动选择最合适的 `utf8mb4` collation。

## 用途

- 解决两表字符/排序规则不一致导致的 `Illegal mix of collations (utf8mb4_xxx,IMPLICIT) ...` 错误。
- 全库统一后，所有字符串列统一排序比较规则。
- `utf8mb4` 字符集完整支持 emoji、中文、多国语言等 4 字节字符。

## 运行

在项目根目录执行：

```bash
# 1. 先预览（默认 dry-run）
go run ./tools/normalize_collations

# 2. 确认计划后执行
go run ./tools/normalize_collations -apply

# 3. 手动指定目标排序规则（MySQL 8.0+ 推荐）
go run ./tools/normalize_collations -collation utf8mb4_0900_ai_ci -apply

# 4. 并发处理（仅表数很多且磁盘 IO 充足时使用，大表会锁表）
go run ./tools/normalize_collations -apply -workers 4
```

## 参数

- `-apply`：真正执行，否则只打印计划。
- `-collation`：手动指定目标排序规则，留空则根据 MySQL 版本自动选择。
  - MySQL 8.0+：`utf8mb4_0900_ai_ci`
  - MySQL 5.7：`utf8mb4_unicode_ci`
  - MariaDB：`utf8mb4_unicode_ci`
- `-workers`：并发处理表数量，默认 `1`。
- `-skip-db`：跳过 `ALTER DATABASE`。
- `-skip-tables`：跳过指定表，逗号分隔。

## 注意事项

- **只支持 MySQL**。SQLite / PostgreSQL 不需要也不支持此操作。
- 生产环境执行前先备份数据库。
- 工具会读取仓库根 `.env` 中的 `DB_DRIVER` / `DB_DSN`（或其他数据库配置）。

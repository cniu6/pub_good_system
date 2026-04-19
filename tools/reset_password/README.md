# reset_password

用于在项目根目录下通过独立工具直接修改用户密码。

## 功能说明

- 按 `id`、`username`、`email` 或 `account`（用户名/邮箱）定位用户
- 复用项目现有密码强度校验与 bcrypt 哈希逻辑
- 修改密码后自动撤销该用户现有登录会话
- 默认同时清除登录失败次数与锁定状态
- 兼容历史 `users.admin_remark = NULL` 的旧数据场景

## 运行前提

请在项目根目录执行命令，工具会自动读取当前仓库根目录的 `.env` 数据库配置。

## 用法示例

### 按用户名修改密码

```powershell
go run ./tools/reset_password -username admin -password "NewPass123"
```

### 按邮箱修改密码

```powershell
go run ./tools/reset_password -email admin@example.com -password "NewPass123"
```

### 按账号修改密码（用户名或邮箱）

```powershell
go run ./tools/reset_password -account admin@example.com -password "NewPass123"
```

### 按用户 ID 修改密码

```powershell
go run ./tools/reset_password -id 1 -password "NewPass123"
```

### 从标准输入读取密码

```powershell
Write-Output "NewPass123" | go run ./tools/reset_password -username admin -password-stdin
```

### 修改密码但不清除锁定状态

```powershell
go run ./tools/reset_password -username admin -password "NewPass123" -clear-lock=false
```

## 参数说明

- `-id`：按用户 ID 查找
- `-username`：按用户名查找
- `-email`：按邮箱查找
- `-account`：按登录账号查找，支持用户名或邮箱
- `-password`：直接传入新密码
- `-password-stdin`：从标准输入读取新密码
- `-clear-lock`：是否同时清理登录失败次数和锁定状态，默认 `true`

## 密码规则

沿用后端现有密码强度校验：

- 至少 8 位
- 至少包含 1 个字母
- 至少包含 1 个数字
- 最多 72 字节

## 失败排查

如果运行时报数据库连接错误，请先确认：

- 当前执行目录是项目根目录
- 根目录 `.env` 中数据库配置正确
- 数据库服务已启动

如果之前遇到过 `admin_remark` 为 `NULL` 导致的扫描错误，现在代码已兼容该场景，并会在数据库迁移时自动把历史 `NULL` 清理为空字符串。

# register_admin

用于在项目根目录下通过独立工具快速注册管理员用户。

## 功能说明

- 创建角色为 `admin` 的管理员用户
- 复用项目现有密码强度校验与 bcrypt 哈希逻辑
- 自动检查用户名与邮箱是否已存在
- 昵称默认同用户名，也可单独指定

## 运行前提

请在项目根目录执行命令，工具会自动读取当前仓库根目录的 `.env` 数据库配置。

## 用法示例

### 基本用法

```powershell
go run ./tools/register_admin -username admin -email admin@example.com -password "Admin123"
```

### 指定昵称与手机号

```powershell
go run ./tools/register_admin -username superadmin -email super@example.com -password "Admin123" -nickname 超级管理员 -mobile 13800138000
```

### 从标准输入读取密码

```powershell
Write-Output "Admin123" | go run ./tools/register_admin -username admin -email admin@example.com -password-stdin
```

## 参数说明

- `-username`：管理员用户名（必填）
- `-email`：管理员邮箱（必填）
- `-password`：直接传入密码
- `-password-stdin`：从标准输入读取密码
- `-nickname`：昵称，默认同用户名
- `-mobile`：手机号（可选）

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

# FST 项目知识库 - AI 调用指南

> 📚 **文档目的**: 本文档用于为 AI 助手提供项目核心知识，防止在代码生成和修改时出现错误。
> 
> 🎯 **使用场景**: 
> - AI 编写代码时需要了解项目结构和 API 用法
> - 防止重复造轮子，应使用已有的工具和封装
> - 确保代码风格和实践符合项目规范

---

## 📖 文档目录

| 文档 | 内容 | 重要性 |
|------|------|--------|
| [邮件系统](./邮件系统.md) | SMTP配置、邮件发送、模板系统 | ⭐⭐⭐⭐⭐ |
| [JWT认证](./JWT认证.md) | Token生成、验证、刷新机制 | ⭐⭐⭐⭐⭐ |
| [插件系统](./插件系统.md) | 插件接口、注册、管理 | ⭐⭐⭐⭐ |
| [数据库模型](./数据库模型.md) | 表结构、模型方法、查询 | ⭐⭐⭐⭐⭐ |
| [配置系统](./配置系统.md) | 环境变量、配置加载 | ⭐⭐⭐⭐ |
| [API路由](./API路由.md) | 路由定义、中间件使用 | ⭐⭐⭐⭐⭐ |
| [前端请求](./前端请求.md) | Alova封装、API调用 | ⭐⭐⭐⭐ |
| [极验验证](./极验验证.md) | 验证码集成、验证流程 | ⭐⭐⭐ |
| [响应规范](./响应规范.md) | 统一响应格式、错误处理 | ⭐⭐⭐⭐ |
| [目录结构](./目录结构.md) | 项目文件组织规范 | ⭐⭐⭐ |

---

## 🚀 快速参考

### 1. 发送邮件（后端）

```go
import "fst/backend/utils"

// 基础发送
err := utils.SendEmail(utils.EmailMessage{
    To:      "user@example.com",
    Subject: "验证码",
    Body:    "您的验证码是：123456",
})

// 使用模板（推荐）
tpl, _ := models.GetEmailTemplate("register_code", "zh-CN")
subject := strings.ReplaceAll(tpl.Subject, "{app_name}", config.GlobalConfig.AppName)
body := strings.ReplaceAll(tpl.Content, "{code}", code)
body = strings.ReplaceAll(body, "{app_name}", config.GlobalConfig.AppName)
```

**关键要点**:
- 必须在 `.env` 配置 `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`
- 支持 SSL 和普通 SMTP
- 使用 `SystemEmailName` 和 `SystemEmail` 设置发件人
- 始终异步记录邮件日志到 `email_logs` 表

---

### 2. JWT Token 操作（后端）

```go
import "fst/backend/utils"

// 生成Token（默认24小时）
token, err := utils.GenerateToken(userID, role)

// 生成指定有效期的Token
token, err := utils.GenerateTokenWithTTL(userID, role, 7*24*time.Hour)

// 解析Token
claims, err := utils.ParseToken(tokenString)
if err != nil {
    // Token无效或过期
}
userID := claims.UserID
role := claims.Role
```

**关键要点**:
- 访问令牌默认24小时有效期
- 刷新令牌建议7天有效期
- JWT密钥从 `config.GlobalConfig.JWTSecret` 获取
- 解析失败时返回错误，不要 panic

---

### 3. 数据库操作（后端）

```go
import "fst/backend/app/models"

// 用户相关
user, err := models.GetUserByUsername(username)
user, err := models.GetUserByEmail(email)
user, err := models.GetUserByUsernameOrEmail(identifier) // 支持用户名或邮箱
err := models.CreateUser(user)

// 验证码
err := models.CreateVerificationCode(email, code, "register", expiresAt)
valid, codeID, err := models.VerifyCode(email, code, "register")
err := models.MarkVerificationCodeAsUsed(codeID)

// 邮件模板
tpl, err := models.GetEmailTemplate("register_code", "zh-CN")
// 如果不存在，会自动回退到 "zh-CN"

// 邮件日志
err := models.CreateEmailLog(to, subject, content, tplName, status, errorMsg)
```

**关键要点**:
- 使用 `sqlx` 库进行数据库操作
- 用户查询会自动排除 `delete_time IS NOT NULL` 的记录
- 验证码表使用软删除（is_deleted）
- 邮件模板支持多语言，自动回退到中文

---

### 4. 路由定义（后端）

```go
// 文件: backend/routes/routes.go

// 公开路由
v1 := api.Group("/v1/user")
{
    v1.POST("/register", authCtrl.Register)
    v1.POST("/send-register-code", authCtrl.SendRegisterCode)
}

// 需要认证的路由
protected := api.Group("/")
protected.Use(middleware.AuthMiddleware())
{
    protected.GET("/profile", handler)
}

// 管理员路由
admin := protected.Group(config.GlobalConfig.AdminPath)
admin.Use(middleware.AdminOnly())
{
    admin.GET("/dashboard", handler)
}
```

**关键要点**:
- 认证中间件从 `Authorization: Bearer <token>` 头解析
- 管理员路径可配置，默认 `/admin`
- Swagger 文档可通过 `ENABLE_SWAGGER=true` 开启

---

### 5. 统一响应（后端）

```go
import "fst/backend/utils"

// 成功响应
utils.Success(c, data)

// 失败响应
utils.Fail(c, 400, "错误信息")

// 标准响应格式
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

**关键要点**:
- 始终使用 `utils.Success` 和 `utils.Fail`
- 不要直接返回 `c.JSON`
- 错误码遵循 HTTP 状态码规范

---

### 6. 配置读取（后端）

```go
import "fst/backend/internal/config"

// 全局配置访问
cfg := config.GlobalConfig

// 常用配置
cfg.AppName         // 应用名称
cfg.JWTSecret       // JWT密钥
cfg.SMTPHost        // SMTP服务器
cfg.GeetestEnabled  // 极验是否启用
cfg.AdminPath       // 管理员路径
```

**关键要点**:
- 配置在 `main.go` 启动时通过 `config.InitConfig()` 加载
- 支持 `.env` 文件（KEY=VALUE 或 JSON 格式）
- 环境变量优先级高于 `.env` 文件

---

### 7. 前端 API 调用

```typescript
// 使用封装好的请求
import { request } from '@/service'

// GET 请求
const { data } = await request.Get<Service.ResponseResult<User[]>>('/api/users')

// POST 请求
const { data } = await request.Post<Service.ResponseResult<User>>('/api/users', {
  name: 'test',
  email: 'test@example.com'
})

// 使用特定 API 封装
import { fetchLogin, fetchSendRegisterCode } from '@/service'
const res = await fetchLogin({ username, password })
```

**关键要点**:
- 使用 Alova 进行请求管理
- 类型定义在 `Service.ResponseResult<T>`
- API 封装函数位于 `src/service/api/` 目录

---

## ⚠️ 常见错误预防

### 错误 1: 直接操作数据库而不使用模型方法

❌ **错误做法**:
```go
// 不要这样写
db.DB.Exec("SELECT * FROM users WHERE username = ?", username)
```

✅ **正确做法**:
```go
// 使用封装好的模型方法
user, err := models.GetUserByUsername(username)
```

### 错误 2: 手动设置响应格式

❌ **错误做法**:
```go
c.JSON(200, gin.H{"status": "ok", "result": data})
```

✅ **正确做法**:
```go
utils.Success(c, data)
```

### 错误 3: 不检查配置就发送邮件

❌ **错误做法**:
```go
utils.SendEmail(msg) // 直接发送，不检查SMTP配置
```

✅ **正确做法**:
```go
if config.GlobalConfig.SMTPHost == "" {
    // 处理未配置的情况
    return fmt.Errorf("SMTP not configured")
}
err := utils.SendEmail(msg)
```

### 错误 4: 前端硬编码 API 路径

❌ **错误做法**:
```typescript
fetch('/api/v1/user/register') // 不要硬编码
```

✅ **正确做法**:
```typescript
import { fetchRegister } from '@/service'
fetchRegister(data) // 使用封装好的API
```

---

## 📁 文件位置速查

| 功能 | 文件路径 |
|------|----------|
| 邮件发送 | `backend/utils/email.go` |
| JWT工具 | `backend/utils/jwt.go` |
| 密码处理 | `backend/utils/password.go` |
| 响应封装 | `backend/utils/response.go` |
| 用户模型 | `backend/app/models/user.go` |
| 验证码模型 | `backend/app/models/verification_code.go` |
| 邮件模板 | `backend/app/models/email.go` |
| 认证控制器 | `backend/app/controllers/auth_controller.go` |
| 系统控制器 | `backend/app/controllers/system_controller.go` |
| 认证中间件 | `backend/internal/middleware/auth.go` |
| CORS中间件 | `backend/internal/middleware/cors.go` |
| 配置管理 | `backend/internal/config/config.go` |
| 数据库初始化 | `backend/internal/db/mysql.go` |
| 路由定义 | `backend/routes/routes.go` |
| 插件接口 | `backend/app/plugins/interface.go` |
| 前端请求封装 | `frontend/src/service/request.ts` |
| 前端API定义 | `frontend/src/service/api/*.ts` |

---

## 🔗 外部资源

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [sqlx 文档](https://jmoiron.github.io/sqlx/)
- [Alova 文档](https://alova.js.org/)
- [Naive UI 文档](https://www.naiveui.com/)

---

> 💡 **提示**: 本文档由 AI 维护，每次重大功能更新后应同步更新相关章节。

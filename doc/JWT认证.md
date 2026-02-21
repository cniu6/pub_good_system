# JWT 认证系统 - 完整使用指南

> 🔐 **文档位置**: `doc/JWT认证.md`
> 
> **关联文件**:
> - `backend/utils/jwt.go` - JWT 生成和验证
> - `backend/internal/middleware/auth.go` - 认证中间件
> - `backend/internal/config/config.go` - JWT 密钥配置

---

## 📋 目录

1. [架构概览](#架构概览)
2. [配置说明](#配置说明)
3. [Token 结构](#token-结构)
4. [生成 Token](#生成-token)
5. [验证 Token](#验证-token)
6. [Token 刷新](#token-刷新)
7. [中间件使用](#中间件使用)
8. [前端集成](#前端集成)
9. [安全最佳实践](#安全最佳实践)
10. [故障排查](#故障排查)

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        JWT 认证流程                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────┐    登录凭证     ┌──────────┐    JWT Token       │
│   │  Client  │ ──────────────► │  Server  │ ─────────────────► │
│   └──────────┘                 └──────────┘                    │
│        │                            │                          │
│        │ Authorization: Bearer xxx  │                          │
│        │◄───────────────────────────┤                          │
│        │                            │                          │
│        │    请求受保护资源          │                          │
│        ├───────────────────────────►│                          │
│        │                            │                          │
│        │                            ▼                          │
│        │                   AuthMiddleware                      │
│        │                   (验证Token)                         │
│        │                            │                          │
│        │         响应结果           │                          │
│        │◄───────────────────────────┤                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 配置说明

### 环境变量

在 `.env` 文件中配置 JWT 密钥：

```env
# JWT 配置
JWT_SECRET=your-super-secret-key-change-in-production  # JWT签名密钥
JWT_EXPIRE_HOURS=24                                    # Token有效期（小时）
```

### 配置代码

**文件**: `backend/internal/config/config.go`

```go
type Config struct {
    JWTSecret string  // JWT签名密钥
    // ... 其他配置
}

// 加载配置
JWTSecret: getEnv("JWT_SECRET", "secret"),
```

### 安全警告

⚠️ **生产环境必须修改默认密钥！**

```env
# 不要这样
JWT_SECRET=secret  # 默认密钥，极不安全！

# 应该这样
JWT_SECRET=your-256-bit-secret-key-here-min-32-chars
```

生成强密钥：
```bash
# Linux/macOS
openssl rand -base64 32

# 或使用 Go
go run -e 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

---

## Token 结构

### Claims 定义

**文件**: `backend/utils/jwt.go`

```go
type Claims struct {
    UserID uint64 `json:"user_id"`  // 用户ID
    Role   string `json:"role"`      // 用户角色 (admin/user)
    jwt.RegisteredClaims             // 标准JWT声明
}
```

### 标准声明 (RegisteredClaims)

```go
type RegisteredClaims struct {
    Issuer    string    `json:"iss,omitempty"`   // 签发者
    Subject   string    `json:"sub,omitempty"`   // 主题
    Audience  []string  `json:"aud,omitempty"`   // 受众
    ExpiresAt *NumericDate `json:"exp,omitempty"` // 过期时间
    NotBefore *NumericDate `json:"nbf,omitempty"` // 生效时间
    IssuedAt  *NumericDate `json:"iat,omitempty"` // 签发时间
    ID        string    `json:"jti,omitempty"`   // 唯一标识
}
```

### JWT Token 示例

**Header**:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload**:
```json
{
  "user_id": 123,
  "role": "admin",
  "exp": 1704067200,
  "iat": 1703980800
}
```

**完整的 Token**:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInJvbGUiOiJhZG1pbiIsImV4cCI6MTcwNDA2NzIwMCwiaWF0IjoxNzAzOTgwODAwfQ.xxxxxx
```

---

## 生成 Token

### API 函数

**文件**: `backend/utils/jwt.go`

```go
// GenerateToken 生成默认24小时有效期的Token
func GenerateToken(userID uint64, role string) (string, error)

// GenerateTokenWithTTL 生成指定有效期的Token
func GenerateTokenWithTTL(userID uint64, role string, ttl time.Duration) (string, error)
```

### 基础使用

```go
package controllers

import (
    "fst/backend/utils"
    "time"
)

func (ctrl *AuthController) Login(c *gin.Context) {
    // ... 验证用户 ...
    
    userID := uint64(123)
    role := "user"
    
    // 生成访问令牌（24小时）
    accessToken, err := utils.GenerateToken(userID, role)
    if err != nil {
        utils.Fail(c, 500, "Failed to generate token")
        return
    }
    
    // 生成刷新令牌（7天）
    refreshToken, err := utils.GenerateTokenWithTTL(userID, role, 7*24*time.Hour)
    if err != nil {
        utils.Fail(c, 500, "Failed to generate refresh token")
        return
    }
    
    utils.Success(c, gin.H{
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
    })
}
```

### 双 Token 策略

```go
// 登录时返回双 Token
func (ctrl *AuthController) Login(c *gin.Context) {
    // ... 验证用户 ...
    
    // 短有效期访问令牌
    accessToken, _ := utils.GenerateTokenWithTTL(user.ID, user.Role, 2*time.Hour)
    
    // 长有效期刷新令牌
    refreshToken, _ := utils.GenerateTokenWithTTL(user.ID, user.Role, 7*24*time.Hour)
    
    utils.Success(c, gin.H{
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
        "tokenType":    "Bearer",
        "expiresIn":    7200, // 2小时（秒）
    })
}
```

### 不同场景的 Token 有效期

| 场景 | 有效期 | 用途 |
|------|--------|------|
| 访问令牌 | 15分钟 - 2小时 | 日常API调用 |
| 刷新令牌 | 7天 - 30天 | 刷新访问令牌 |
| 密码重置 | 15分钟 | 安全敏感操作 |
| 邮箱验证 | 24小时 | 验证邮箱地址 |

---

## 验证 Token

### API 函数

```go
// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string) (*Claims, error)
```

### 基础使用

```go
claims, err := utils.ParseToken(tokenString)
if err != nil {
    // Token无效或过期
    log.Printf("Invalid token: %v", err)
    return
}

// 获取声明信息
userID := claims.UserID
role := claims.Role
expiresAt := claims.ExpiresAt
```

### 错误处理

```go
claims, err := utils.ParseToken(tokenString)
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrTokenExpired):
        // Token过期
        return nil, fmt.Errorf("token expired")
    case errors.Is(err, jwt.ErrTokenMalformed):
        // Token格式错误
        return nil, fmt.Errorf("token malformed")
    case errors.Is(err, jwt.ErrTokenSignatureInvalid):
        // 签名无效
        return nil, fmt.Errorf("invalid token signature")
    default:
        // 其他错误
        return nil, fmt.Errorf("invalid token: %w", err)
    }
}
```

### 检查 Token 是否即将过期

```go
func isTokenNearExpiry(claims *utils.Claims, threshold time.Duration) bool {
    if claims.ExpiresAt == nil {
        return false
    }
    return time.Until(claims.ExpiresAt.Time) < threshold
}

// 使用示例
if isTokenNearExpiry(claims, 5*time.Minute) {
    // Token将在5分钟内过期，提示刷新
}
```

---

## Token 刷新

### 刷新流程

```
┌─────────┐         ┌─────────┐         ┌─────────┐
│ Client  │         │ Server  │         │  DB     │
└────┬────┘         └────┬────┘         └────┬────┘
     │                   │                   │
     │ POST /refresh     │                   │
     │ {refreshToken}    │                   │
     ├──────────────────►│                   │
     │                   │                   │
     │                   │ 验证 refreshToken │
     │                   ├──────────────────►│
     │                   │◄──────────────────┤
     │                   │                   │
     │                   │ 生成新Token       │
     │                   │                   │
     │ {newTokens}       │                   │
     │◄──────────────────┤                   │
     │                   │                   │
```

### 刷新控制器

**文件**: `backend/app/controllers/auth_controller.go`

```go
func (ctrl *AuthController) UpdateToken(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Fail(c, 400, err.Error())
        return
    }
    
    // 验证刷新令牌
    claims, err := utils.ParseToken(req.RefreshToken)
    if err != nil {
        utils.Fail(c, 401, "Invalid or expired refresh token")
        return
    }
    
    // 生成新的双令牌
    accessToken, err := utils.GenerateTokenWithTTL(claims.UserID, claims.Role, 24*time.Hour)
    if err != nil {
        utils.Fail(c, 500, "Failed to generate access token")
        return
    }
    
    refreshToken, err := utils.GenerateTokenWithTTL(claims.UserID, claims.Role, 7*24*time.Hour)
    if err != nil {
        utils.Fail(c, 500, "Failed to generate refresh token")
        return
    }
    
    utils.Success(c, gin.H{
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
    })
}
```

### 刷新请求格式

```http
POST /api/updateToken
Content-Type: application/json

{
    "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 刷新响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "accessToken": "eyJhbGciOiJIUzI1NiIs...",
        "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
    }
}
```

---

## 中间件使用

### 认证中间件

**文件**: `backend/internal/middleware/auth.go`

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            c.Abort()
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if !(len(parts) == 2 && parts[0] == "Bearer") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
            c.Abort()
            return
        }
        
        claims, err := utils.ParseToken(parts[1])
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }
        
        // 将用户信息存入上下文
        c.Set("userID", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### 管理员权限中间件

```go
func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Admin access only"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 路由使用示例

**文件**: `backend/routes/routes.go`

```go
func SetupRoutes(router *gin.Engine) {
    api := router.Group("/api")
    {
        // 公开路由
        api.POST("/login", authCtrl.Login)
        api.POST("/register", authCtrl.Register)
        
        // 受保护路由
        protected := api.Group("/")
        protected.Use(middleware.AuthMiddleware())
        {
            protected.GET("/profile", getProfile)
            protected.POST("/updateToken", authCtrl.UpdateToken)
            
            // 管理员路由
            admin := protected.Group(config.GlobalConfig.AdminPath)
            admin.Use(middleware.AdminOnly())
            {
                admin.GET("/dashboard", adminDashboard)
            }
        }
    }
}
```

### 从上下文获取用户信息

```go
func getProfile(c *gin.Context) {
    // 从上下文中获取
    userID, exists := c.Get("userID")
    if !exists {
        utils.Fail(c, 401, "User ID not found in context")
        return
    }
    
    role, _ := c.Get("role")
    
    utils.Success(c, gin.H{
        "userID": userID,
        "role":   role,
    })
}
```

---

## 前端集成

### 存储 Token

```typescript
// 登录成功后存储
function handleLogin(response: LoginResponse) {
    const { accessToken, refreshToken } = response.data
    
    // 存储到 localStorage（或更安全的存储方式）
    localStorage.setItem('accessToken', accessToken)
    localStorage.setItem('refreshToken', refreshToken)
}
```

### 请求拦截器添加 Token

```typescript
// frontend/src/service/request.ts
import { createAlova } from 'alova'

const request = createAlova({
    baseURL: '/api',
    beforeRequest(method) {
        // 添加认证头
        const token = localStorage.getItem('accessToken')
        if (token) {
            method.config.headers.Authorization = `Bearer ${token}`
        }
    },
    responded: {
        onSuccess: async (response) => {
            return response.json()
        },
        onError: async (error, method) => {
            if (error.status === 401) {
                // Token过期，尝试刷新
                await refreshToken()
                // 重试原请求
                return method.send()
            }
            throw error
        }
    }
})
```

### Token 刷新逻辑

```typescript
let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

async function refreshToken(): Promise<string> {
    if (isRefreshing) {
        // 等待刷新完成
        return new Promise((resolve) => {
            refreshSubscribers.push(resolve)
        })
    }
    
    isRefreshing = true
    
    try {
        const refreshToken = localStorage.getItem('refreshToken')
        const { data } = await fetchUpdateToken({ refreshToken })
        
        localStorage.setItem('accessToken', data.accessToken)
        localStorage.setItem('refreshToken', data.refreshToken)
        
        // 通知等待的请求
        refreshSubscribers.forEach(callback => callback(data.accessToken))
        refreshSubscribers = []
        
        return data.accessToken
    } catch (error) {
        // 刷新失败，跳转登录
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
        throw error
    } finally {
        isRefreshing = false
    }
}
```

### API 封装示例

**文件**: `frontend/src/service/api/auth.ts`

```typescript
import { request } from '../request'

// 登录
export function fetchLogin(data: { username: string; password: string }) {
    return request.Post<Service.ResponseResult<LoginResult>>('/api/login', data)
}

// 刷新Token
export function fetchUpdateToken(data: { refreshToken: string }) {
    return request.Post<Service.ResponseResult<TokenResult>>('/api/updateToken', data)
}

// 获取用户信息
export function fetchUserProfile() {
    return request.Get<Service.ResponseResult<UserProfile>>('/api/profile')
}
```

---

## 安全最佳实践

### 1. 密钥管理

```go
// ✅ 正确：从环境变量读取
JWTSecret: getEnv("JWT_SECRET", ""),

// ❌ 错误：硬编码密钥
JWTSecret: "my-secret-key",
```

### 2. 使用 HTTPS

```go
// 生产环境强制 HTTPS
if config.GlobalConfig.AppMode == "production" {
    // 检查请求是否通过 HTTPS
    if c.Request.TLS == nil {
        utils.Fail(c, 403, "HTTPS required")
        return
    }
}
```

### 3. Token 存储安全

**前端**:
```typescript
// 更安全的存储方式
// 1. httpOnly Cookie（推荐）
// 2. Memory storage
// 3. 避免 localStorage（XSS风险）

// 如果需要使用 localStorage，添加额外保护
const encryptedToken = encrypt(token) // 使用 Web Crypto API
localStorage.setItem('token', encryptedToken)
```

**后端**:
```go
// 设置 Cookie 选项（如果使用 Cookie 存储）
c.SetCookie("token", token, 3600, "/", "", true, // Secure
    true, // HttpOnly
)
```

### 4. Token 绑定设备/IP

```go
// 在 Claims 中添加设备信息
type Claims struct {
    UserID   uint64 `json:"user_id"`
    Role     string `json:"role"`
    DeviceID string `json:"device_id"`  // 设备标识
    IP       string `json:"ip"`          // IP地址
    jwt.RegisteredClaims
}

// 验证时检查
func ValidateTokenWithContext(tokenString, deviceID, ip string) (*Claims, error) {
    claims, err := ParseToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    // 检查设备/IP是否匹配
    if claims.DeviceID != deviceID || claims.IP != ip {
        return nil, fmt.Errorf("token context mismatch")
    }
    
    return claims, nil
}
```

### 5. Token 黑名单（登出）

```go
// 使用 Redis 存储已注销的 Token
var tokenBlacklist = make(map[string]time.Time)

func InvalidateToken(tokenID string, expiry time.Time) {
    tokenBlacklist[tokenID] = expiry
}

func IsTokenInvalidated(tokenID string) bool {
    expiry, exists := tokenBlacklist[tokenID]
    if !exists {
        return false
    }
    
    // 清理过期条目
    if time.Now().After(expiry) {
        delete(tokenBlacklist, tokenID)
        return false
    }
    
    return true
}
```

### 6. 定期轮换密钥

```go
// 支持多个密钥（新旧密钥同时有效）
var jwtSecrets = []string{
    os.Getenv("JWT_SECRET_V2"),  // 新密钥
    os.Getenv("JWT_SECRET"),     // 旧密钥
}

func ParseTokenWithKeyRotation(tokenString string) (*Claims, error) {
    for _, secret := range jwtSecrets {
        claims, err := parseTokenWithSecret(tokenString, secret)
        if err == nil {
            return claims, nil
        }
    }
    return nil, fmt.Errorf("invalid token")
}
```

---

## 故障排查

### 问题 1: "Invalid or expired token"

**排查步骤**:

1. 检查 Token 格式
```bash
# 解码 JWT 查看内容
echo "eyJhbGciOiJIUzI1NiIs..." | base64 -d
```

2. 检查密钥是否一致
```go
// 调试：打印实际使用的密钥
fmt.Printf("JWT Secret: %s\n", config.GlobalConfig.JWTSecret)
```

3. 检查 Token 过期时间
```go
claims, _ := utils.ParseToken(token)
fmt.Printf("Token expires at: %v\n", claims.ExpiresAt)
fmt.Printf("Current time: %v\n", time.Now())
```

### 问题 2: 前端请求 401

**排查步骤**:

1. 检查请求头
```javascript
console.log(localStorage.getItem('accessToken'))
// 确保请求头包含: Authorization: Bearer <token>
```

2. 检查 Token 是否过期
```go
// 后端添加调试日志
claims, err := utils.ParseToken(token)
if err != nil {
    log.Printf("Token parse error: %v", err)
    log.Printf("Token: %s", token) // 注意：生产环境不要记录完整Token
}
```

### 问题 3: Token 刷新失败

**排查步骤**:

1. 检查刷新令牌是否过期
2. 检查刷新令牌是否在黑名单中
3. 检查用户状态（是否被禁用）

---

## API 参考

### utils/jwt.go

| 函数 | 签名 | 说明 |
|------|------|------|
| GenerateToken | `func GenerateToken(userID uint64, role string) (string, error)` | 生成24小时有效期的Token |
| GenerateTokenWithTTL | `func GenerateTokenWithTTL(userID uint64, role string, ttl time.Duration) (string, error)` | 生成指定有效期的Token |
| ParseToken | `func ParseToken(tokenString string) (*Claims, error)` | 解析并验证Token |

### middleware/auth.go

| 函数 | 签名 | 说明 |
|------|------|------|
| AuthMiddleware | `func AuthMiddleware() gin.HandlerFunc` | JWT认证中间件 |
| AdminOnly | `func AdminOnly() gin.HandlerFunc` | 管理员权限中间件 |

---

## 扩展阅读

- [JWT.io](https://jwt.io/) - JWT 调试工具
- [RFC 7519](https://tools.ietf.org/html/rfc7519) - JWT 规范
- [Golang-JWT 文档](https://github.com/golang-jwt/jwt)

---

> 📝 **最后更新**: 2026-02-04
> 
> 如有疑问，请参考 `backend/utils/jwt.go` 和 `backend/internal/middleware/auth.go` 源代码。

# API 路由系统 - 完整使用指南

> 🌐 **文档位置**: `doc/API路由.md`
> 
> **关联文件**:
> - `backend/routes/routes.go` - 路由定义主文件
> - `backend/internal/middleware/*.go` - 中间件
> - `backend/app/controllers/*.go` - 控制器

---

## 📋 目录

1. [架构概览](#架构概览)
2. [路由定义](#路由定义)
3. [中间件使用](#中间件使用)
4. [控制器编写](#控制器编写)
5. [路由分组](#路由分组)
6. [Swagger 文档](#swagger-文档)
7. [常见模式](#常见模式)

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        请求处理流程                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   HTTP Request                                                   │
│       │                                                          │
│       ▼                                                          │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐                 │
│   │  Logger  │───►│   CORS   │───►│  Recovery│                 │
│   │  日志    │    │  跨域    │    │  错误恢复 │                 │
│   └──────────┘    └──────────┘    └──────────┘                 │
│       │                                                          │
│       ▼                                                          │
│   Router (路由匹配)                                               │
│       │                                                          │
│       ├──► /api/login ──► AuthMiddleware? ──► LoginHandler     │
│       │                                                          │
│       ├──► /api/v1/user/* ──► (公开路由)                         │
│       │                                                          │
│       └──► /api/* ──► AuthMiddleware ──► Protected Handlers    │
│                       │                                          │
│                       └──► /admin/* ──► AdminOnly               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 路由定义

### 主路由文件

**文件**: `backend/routes/routes.go`

```go
package routes

import (
    "fst/backend/app/controllers"
    _ "fst/backend/docs"
    "fst/backend/internal/config"
    "fst/backend/internal/middleware"
    "fst/backend/utils"
    
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine) {
    // 初始化控制器
    authCtrl := &controllers.AuthController{}
    systemCtrl := &controllers.SystemController{}
    
    // Swagger 文档路由
    if config.GlobalConfig.EnableSwagger {
        router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }
    
    // API 路由组
    api := router.Group("/api")
    {
        // ========== V1 用户路由（公开）==========
        v1 := api.Group("/v1/user")
        {
            v1.POST("/register", authCtrl.Register)
            v1.POST("/send-register-code", authCtrl.SendRegisterCode)
            v1.POST("/resetpasswordtoemail", authCtrl.SendResetEmail)
            v1.POST("/reset-password", authCtrl.ResetPasswordConfirm)
        }
        
        // ========== 认证路由 ==========
        api.POST("/register", authCtrl.Register)
        api.POST("/login", authCtrl.Login)
        api.POST("/updateToken", authCtrl.UpdateToken)
        api.GET("/getUserRoutes", authCtrl.GetUserRoutes)
        
        // ========== 系统路由 ==========
        api.GET("/userPage", systemCtrl.GetUserPage)
        api.GET("/role/list", systemCtrl.GetRoleList)
        api.GET("/dict/list", systemCtrl.GetDictList)
        
        // ========== 兼容提示路由 ==========
        api.GET("/login", func(c *gin.Context) {
            utils.Fail(c, 405, "Please use POST method to login")
        })
        api.GET("/register", func(c *gin.Context) {
            utils.Fail(c, 405, "Please use POST method to register")
        })
        
        // ========== 受保护路由 ==========
        protected := api.Group("/")
        protected.Use(middleware.AuthMiddleware())
        {
            protected.GET("/profile", getProfileHandler)
            
            // 管理员路由
            admin := protected.Group(config.GlobalConfig.AdminPath)
            admin.Use(middleware.AdminOnly())
            {
                admin.GET("/dashboard", adminDashboardHandler)
            }
        }
    }
}
```

### 调试相关接口（管理员）

后端调试能力由 `backend/app/controllers/admin/debug_controller.go` 提供，并挂载在管理员路由组下。

- 路由前缀：`/api/v1/admin/debug`
- 典型接口：
  - `GET /goroutines/stats`
  - `POST /gc`
  - `GET /pprof/profile`
  - `GET /pprof/heap`
  - `GET /pprof/goroutine`
  - `GET /pprof/allocs`
  - `GET /pprof/block`
  - `GET /pprof/mutex`
  - `GET /pprof/trace`

> 说明：前端已取消侧边栏独立“调试”页面，调试能力统一放在“系统设置”页面中。

### 路由注册

**文件**: `backend/cmd/main.go`

```go
func main() {
    // ... 初始化代码
    
    // 创建 Gin 实例
    router := gin.Default()
    
    // 设置路由
    routes.SetupRoutes(router)
    
    // 启动服务器
    router.Run(":" + config.GlobalConfig.Port)
}
```

---

## 中间件使用

### 内置中间件

```go
// 创建带有默认中间件的 Gin 实例
// 默认包含 Logger 和 Recovery
router := gin.Default()

// 创建不带有任何中间件的 Gin 实例
router := gin.New()

// 手动添加中间件
router.Use(gin.Logger())
router.Use(gin.Recovery())
router.Use(middleware.CORSMiddleware())
router.Use(middleware.AuthMiddleware())
```

### 认证中间件

**文件**: `backend/internal/middleware/auth.go`

```go
package middleware

import (
    "net/http"
    "strings"
    "fst/backend/utils"
    "github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if !(len(parts) == 2 && parts[0] == "Bearer") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
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

// AdminOnly 管理员权限中间件
func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### CORS 中间件

**文件**: `backend/internal/middleware/cors.go`

```go
package middleware

import (
    "fst/backend/internal/config"
    "github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := config.GlobalConfig.CorsOrigins
        if origin == "" {
            origin = "*"
        }
        
        c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", 
            "Content-Type, Authorization, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", 
            "GET, POST, PUT, DELETE, OPTIONS")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### 自定义中间件示例

```go
// LoggingMiddleware 日志记录中间件
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        // 处理请求
        c.Next()
        
        // 记录日志
        duration := time.Since(start)
        status := c.Writer.Status()
        
        log.Printf("[%s] %s %d %v", 
            c.Request.Method, 
            path, 
            status, 
            duration,
        )
    }
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
    // 使用内存存储访问次数
    // 生产环境建议使用 Redis
    visits := make(map[string]int)
    
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        if visits[clientIP] >= limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        visits[clientIP]++
        
        // 定时清理
        go func() {
            time.Sleep(window)
            visits[clientIP]--
        }()
        
        c.Next()
    }
}
```

---

## 控制器编写

### 基础控制器

```go
package controllers

import (
    "fst/backend/utils"
    "github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct{}

// GetProfile 获取用户信息
func (ctrl *UserController) GetProfile(c *gin.Context) {
    // 从上下文获取用户ID
    userID, exists := c.Get("userID")
    if !exists {
        utils.Fail(c, 401, "Unauthorized")
        return
    }
    
    // 查询用户信息
    user, err := models.GetUserByID(userID.(uint64))
    if err != nil {
        utils.Fail(c, 500, "Failed to get user info")
        return
    }
    
    utils.Success(c, user)
}
```

### 请求验证

```go
// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Role     string `json:"role" binding:"oneof=user admin"`
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    // 绑定并验证请求
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Fail(c, 400, err.Error())
        return
    }
    
    // 业务逻辑
    // ...
    
    utils.Success(c, gin.H{"message": "User created"})
}
```

### 常用验证标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `required` | 必填 | `binding:"required"` |
| `min` | 最小长度 | `binding:"min=3"` |
| `max` | 最大长度 | `binding:"max=50"` |
| `email` | 邮箱格式 | `binding:"email"` |
| `url` | URL格式 | `binding:"url"` |
| `uuid` | UUID格式 | `binding:"uuid"` |
| `oneof` | 枚举值 | `binding:"oneof=a b c"` |
| `gt` | 大于 | `binding:"gt=0"` |
| `gte` | 大于等于 | `binding:"gte=0"` |
| `lt` | 小于 | `binding:"lt=100"` |
| `lte` | 小于等于 | `binding:"lte=100"` |
| `len` | 固定长度 | `binding:"len=6"` |

---

## 变更记录

### 2026-02-24
- 前端管理端侧边栏移除独立“调试”页面入口，调试能力统一放在“系统设置”页面。
- 补充管理员调试接口说明（`/api/v1/admin/debug/*`）和典型端点。
- 路由示例中“调试路由”命名调整为“兼容提示路由”，避免与管理员调试接口概念混淆。

---

## 路由分组

### 分组策略

```go
func SetupRoutes(router *gin.Engine) {
    // API 主分组
    api := router.Group("/api")
    api.Use(middleware.CORSMiddleware())
    {
        // 公开路由
        public := api.Group("/")
        {
            public.POST("/login", loginHandler)
            public.POST("/register", registerHandler)
        }
        
        // V1 版本路由
        v1 := api.Group("/v1")
        {
            // 用户模块
            user := v1.Group("/user")
            {
                user.GET("/profile", getProfile)
                user.PUT("/profile", updateProfile)
                user.POST("/avatar", uploadAvatar)
            }
            
            // 订单模块
            order := v1.Group("/order")
            order.Use(middleware.AuthMiddleware())
            {
                order.GET("/list", getOrderList)
                order.POST("/create", createOrder)
                order.GET("/:id", getOrderDetail)
            }
        }
        
        // 受保护路由
        protected := api.Group("/")
        protected.Use(middleware.AuthMiddleware())
        {
            protected.GET("/dashboard", dashboardHandler)
            
            // 管理员子分组
            admin := protected.Group("/admin")
            admin.Use(middleware.AdminOnly())
            {
                admin.GET("/users", adminGetUsers)
                admin.POST("/users", adminCreateUser)
                admin.DELETE("/users/:id", adminDeleteUser)
            }
        }
    }
}
```

---

## Swagger 文档

### 添加 Swagger 注释

```go
package controllers

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取 JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} utils.Response{data=LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
    // ... 实现
}

// GetUserProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Router /api/profile [get]
func (ctrl *UserController) GetProfile(c *gin.Context) {
    // ... 实现
}
```

### 更新 Swagger 文档

```bash
# 在项目根目录执行
go run github.com/swaggo/swag/cmd/swag init \
    -g backend/cmd/main.go \
    -o backend/docs

# 或简写
go run github.com/swaggo/swag/cmd/swag init -g backend/cmd/main.go
```

### 访问 Swagger UI

```
http://localhost:8080/swagger/index.html
```

---

## 常见模式

### RESTful API 设计

```go
// 用户资源
api.GET("/users", listUsers)          // 列表
api.POST("/users", createUser)        // 创建
api.GET("/users/:id", getUser)        // 详情
api.PUT("/users/:id", updateUser)     // 更新（全量）
api.PATCH("/users/:id", patchUser)    // 更新（部分）
api.DELETE("/users/:id", deleteUser)  // 删除
```

### 统一响应格式

```go
// utils/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(200, Response{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func Fail(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}
```

### 错误处理

```go
func (ctrl *UserController) GetUser(c *gin.Context) {
    id := c.Param("id")
    userID, err := strconv.ParseUint(id, 10, 64)
    if err != nil {
        utils.Fail(c, 400, "Invalid user ID")
        return
    }
    
    user, err := models.GetUserByID(userID)
    if err != nil {
        log.Printf("Failed to get user: %v", err)
        utils.Fail(c, 500, "Internal server error")
        return
    }
    
    if user == nil {
        utils.Fail(c, 404, "User not found")
        return
    }
    
    utils.Success(c, user)
}
```

---

## API 参考

### routes/routes.go

| 函数 | 签名 | 说明 |
|------|------|------|
| SetupRoutes | `func SetupRoutes(router *gin.Engine)` | 设置所有路由 |

### middleware/auth.go

| 函数 | 签名 | 说明 |
|------|------|------|
| AuthMiddleware | `func AuthMiddleware() gin.HandlerFunc` | JWT认证中间件 |
| AdminOnly | `func AdminOnly() gin.HandlerFunc` | 管理员权限中间件 |

### middleware/cors.go

| 函数 | 签名 | 说明 |
|------|------|------|
| CORSMiddleware | `func CORSMiddleware() gin.HandlerFunc` | 跨域中间件 |

---

> 📝 **最后更新**: 2026-02-24
> 
> 如有疑问，请参考 `backend/routes/routes.go` 源代码。

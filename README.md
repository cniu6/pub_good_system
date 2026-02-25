# FST - 全栈后台管理系统模板

FST (Full Stack Template) 是一个基于 Go (Gin) 和 Vue 3 (Naive UI) 构建的高性能、轻量级全栈后台管理系统模板。它集成了现代 Web 开发的最佳实践，旨在提供一个开箱即用的开发基础。

## 🚀 特性

- **前后端分离架构**：后端采用 Go Gin，前端采用 Vue 3 + Vite。
- **高性能后端**：使用 Go 语言，具备极高的并发处理能力和内存效率。
- **现代化前端**：基于 Vue 3、TypeScript、Naive UI、UnoCSS 和 Alova。
- **插件化设计**：后端支持插件扩展，自动发现注册，无需手动导入。
- **完善的权限系统**：内置 JWT 认证和路由级权限控制。
- **Swagger API 文档**：启动时自动更新，无需手动维护。
- **多语言支持**：内置 i18n 国际化方案。
- **一键构建**：提供跨平台构建脚本（Windows/Linux）。

## 🛠️ 技术栈

### 后端 (Backend)
- **框架**: [Gin](https://gin-gonic.dev/)
- **数据库**: [MySQL](https://www.mysql.com/) (使用 [sqlx](https://github.com/jmoiron/sqlx))
- **认证**: [JWT](https://github.com/golang-jwt/jwt)
- **文档**: [Swagger](https://github.com/swaggo/swag)
- **配置**: [godotenv](https://github.com/joho/godotenv)

### 前端 (Frontend)
- **框架**: [Vue 3](https://vuejs.org/)
- **构建工具**: [Vite](https://vitejs.dev/)
- **UI 组件库**: [Naive UI](https://www.naiveui.com/)
- **样式**: [UnoCSS](https://unocss.dev/)
- **状态管理**: [Pinia](https://pinia.vuejs.org/)
- **网络请求**: [Alova](https://alova.js.org/)
- **语言**: [TypeScript](https://www.typescriptlang.org/)

## 📂 项目结构

```text
.
├── backend/               # 后端 Go 源码
│   ├── app/
│   │   ├── controllers/   # 控制器 (public/user/admin 三层架构)
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务服务层
│   │   └── plugins/       # 插件目录 (自动发现)
│   ├── cmd/               # 程序启动入口
│   ├── docs/              # Swagger 文档 (自动生成)
│   ├── internal/          # 内部系统库
│   │   ├── config/        # 配置管理
│   │   ├── db/            # 数据库连接
│   │   └── middleware/    # 中间件 (认证、限流、日志等)
│   ├── pkg/               # 公共包
│   │   └── pluginregistry/ # 插件注册表
│   ├── routes/            # API 路由定义
│   └── utils/             # 通用工具函数
├── frontend/              # 前端 Vue 源码
│   ├── src/
│   │   ├── api/           # API 接口定义
│   │   ├── components/    # 组件
│   │   ├── views/         # 页面
│   │   └── ...
│   └── ...
├── build/                 # 构建产物
├── .env                   # 环境变量配置
└── dev.bat/ps1            # 开发启动脚本
```

## ⚙️ 快速开始

### 1. 环境准备
- [Go](https://golang.org/) 1.24+
- [Node.js](https://nodejs.org/) 18+ & [pnpm](https://pnpm.io/)
- [MySQL](https://www.mysql.com/) 8.0+

### 2. 安装依赖

```bash
# 安装 swag (Swagger 文档生成工具)
go install github.com/swaggo/swag/cmd/swag@latest

# 安装前端依赖
cd frontend && pnpm install
```

### 3. 数据库配置
1. 创建数据库 `fst_platform`。
2. 复制 `.env.example` 为 `.env`，修改数据库连接信息。
3. 运行项目时，后端会自动执行数据库迁移并创建必要的表。

### 4. 运行项目

```bash
# 方式1: 直接运行 (推荐) - Swagger 自动更新
go run ./backend/cmd/main.go

# 方式2: 使用脚本
./dev.bat          # Windows
./dev.ps1          # PowerShell

# 方式3: 热重载模式 (需安装 air)
go install github.com/cosmtrek/air@latest
air
```

## 📝 API 文档 (Swagger)

项目已集成 Swagger，**启动时自动更新文档**，无需手动维护。

### 访问文档
启动后端服务后，访问：`http://localhost:8080/swagger/index.html`

### 已包含的 API 分组

| 分组 | 端点数 | 说明 |
|------|--------|------|
| Admin-邮件模板 | 4 | 邮件模板管理 |
| Admin-用户管理 | 7 | 用户 CRUD |
| Admin-操作日志 | 3 | 日志查询和清理 |
| Public-认证 | 6 | 登录、注册、密码重置 |
| 用户中心 | 7 | 个人信息管理 |
| Plugin-* | 动态 | 插件自动注册 |

### 添加新的 API 文档

在控制器方法上添加 Swagger 注解：

```go
// GetUser 获取用户
// @Summary 获取用户详情
// @Description 根据ID获取用户信息
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
    // ...
}
```

重启服务后，Swagger 文档会自动更新！

## 🔌 插件系统

### 自动发现机制

FST 实现了**完全自动化的插件发现和注册**：

1. **自动发现** - 扫描 `backend/app/plugins/` 目录
2. **自动注册** - 插件通过 `init()` 自动注册
3. **自动文档** - Swagger 自动包含插件 API
4. **零配置** - 无需手动导入或配置

### 创建新插件

在 `backend/app/plugins/` 下创建新目录：

```go
// backend/app/plugins/myplugin/plugin.go
package myplugin

import (
    "fst/backend/app/plugins"
    "fst/backend/pkg/pluginregistry"
    "fst/backend/utils"
    "github.com/gin-gonic/gin"
)

// 自动注册 (必需)
func init() {
    pluginregistry.Register(NewPlugin())
}

type MyPlugin struct {
    plugins.BasePlugin
}

func NewPlugin() plugins.Plugin {
    return &MyPlugin{
        BasePlugin: plugins.NewBasePlugin(
            "my-plugin",      // 插件名称
            "1.0.0",          // 版本
            "我的插件描述",    // 描述
        ),
    }
}

// 实现必需接口
func (p *MyPlugin) Init() error { return nil }
func (p *MyPlugin) Migrate() error { return nil }
func (p *MyPlugin) Configure(config map[string]interface{}) error { return nil }
func (p *MyPlugin) Shutdown() error { return nil }

// 注册路由
func (p *MyPlugin) RegisterRoutes(router *gin.RouterGroup) {
    router.GET("/my-plugin/hello", p.Hello)
}

// API 处理函数 (Swagger 注解会自动识别)
// @Summary 测试接口
// @Tags Plugin-MyPlugin
// @Success 200 {object} utils.Response
// @Router /api/v1/my-plugin/hello [get]
func (p *MyPlugin) Hello(c *gin.Context) {
    utils.Success(c, gin.H{"message": "Hello from MyPlugin!"})
}
```

**就这样！重启服务后：**
- 插件自动被发现和加载
- 路由自动注册
- Swagger 文档自动更新

### 插件生命周期

```
Configure() → Init() → Migrate() → RegisterRoutes() → [运行中] → Shutdown()
```

## 🏗️ 构建部署

### 开发模式
```bash
go run ./backend/cmd/main.go
```

### 生产构建
```bash
# Windows
./build.bat

# 手动构建
cd backend
go build -o ../build/fst ./cmd/main.go
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GO_ENV` | 运行环境 (production 跳过自动更新) | development |
| `SKIP_AUTO_SWAGGER` | 跳过 Swagger 自动更新 | false |
| `ENABLE_SWAGGER` | 是否启用 Swagger 端点 | true |

## 📋 控制器架构

项目采用三层控制器架构：

```
backend/app/controllers/
├── public/          # 公共接口 (无需登录)
│   └── auth_controller.go
├── user/            # 用户接口 (需要登录)
│   └── profile_controller.go
├── admin/           # 管理接口 (需要管理员权限)
│   ├── user_controller.go
│   ├── log_controller.go
│   └── email_template_controller.go
└── system_controller.go  # 系统接口
```

### 路由前缀

- `/api/v1/public/*` - 公共接口
- `/api/v1/user/*` - 用户接口
- `/api/v1/admin/*` - 管理接口
- `/api/v1/demo/*` - 插件接口 (动态)

## 🔐 安全特性

- **JWT 认证** - 基于 JWT 的无状态认证
- **密码加密** - bcrypt 密码哈希
- **XSS 过滤** - 用户输入自动过滤
- **SQL 注入防护** - 参数化查询
- **限流中间件** - 令牌桶算法限流
- **CORS 配置** - 跨域请求控制
- **登录失败锁定** - 防暴力破解

## 📄 开源协议

本项目基于 [MIT](LICENSE) 协议开源。

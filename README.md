# FST - 全栈后台管理系统模板

FST (Full Stack Template) 是一个基于 Go (Gin) 和 Vue 3 (Naive UI) 构建的高性能、轻量级全栈后台管理系统模板。前后端分离架构，支持插件化扩展，开箱即用。

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go 1.24+ · Gin · MySQL (sqlx) · JWT · Swagger |
| 前端 | Vue 3 · TypeScript · Vite · Naive UI · UnoCSS · Pinia · Alova |
| 构建 | Windows / Linux · amd64 / arm64 |

## 项目结构

```text
fst/
├── backend/                 # 后端 Go 源码
│   ├── app/
│   │   ├── controllers/     # 控制器 (public/user/admin 三层)
│   │   ├── models/          # 数据模型
│   │   ├── services/        # 业务服务层
│   │   └── plugins/         # 插件目录 (自动发现)
│   ├── cmd/                 # 程序入口
│   ├── docs/                # Swagger 文档 (自动生成)
│   ├── internal/            # 内部系统库 (config/db/middleware)
│   ├── pkg/pluginregistry/  # 插件注册表
│   ├── routes/              # API 路由定义
│   └── utils/               # 通用工具
├── frontend/                # 前端 Vue 源码
├── doc/                     # 详细文档
├── build/                   # 构建产物
├── .env / .env.example      # 环境变量
├── dev.bat                  # 开发启动脚本
├── build.bat                # 生产构建脚本
└── test.bat                 # 集成测试脚本
```

## 快速开始

### 1. 环境准备

- [Go](https://golang.org/) 1.24+
- [Node.js](https://nodejs.org/) 18+ & [pnpm](https://pnpm.io/)
- [MySQL](https://www.mysql.com/) 8.0+

### 2. 安装依赖

```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd frontend && pnpm install
```

### 3. 数据库配置

1. 创建数据库 `fst_platform`
2. 复制 `.env.example` 为 `.env`，修改数据库连接信息
3. 启动后自动迁移建表

### 4. 启动开发

```bash
# 直接运行 (推荐，Swagger 自动更新)
go run ./backend/cmd/main.go

# 使用脚本
./dev.bat          # Windows

# 热重载 (需安装 air)
go install github.com/cosmtrek/air@latest
air
```

### 5. 生产构建

```bash
./build.bat        # 自动构建 Windows/Linux 双平台产物，输出到 build/
```

## 核心特性

### 插件系统 (全自动)

在 `backend/app/plugins/` 下创建目录，实现 Plugin 接口即可。无需手动导入配置，重启自动生效。

```go
// backend/app/plugins/myplugin/plugin.go
func init() {
    pluginregistry.Register(NewPlugin())
}
```

生命周期：`Configure() → Init() → Migrate() → RegisterRoutes() → [运行中] → Shutdown()`

### JWT 认证

双 Token 机制 (Access + Refresh)，支持用户名/邮箱登录，注册验证码，登录失败锁定。

### Swagger 文档

启动时自动检测代码变化并重新生成，自动包含插件 API。访问：`http://localhost:8080/swagger/index.html`

### 控制器三层架构

```text
controllers/
├── public/     # 无需登录 (登录/注册/密码重置)
├── user/       # 需要登录 (个人信息/修改密码)
└── admin/      # 需要管理员权限 (用户管理/日志/邮件模板)
```

### 中间件

| 中间件 | 功能 |
|--------|------|
| JWT 认证 | Token 验证、用户信息注入 |
| CORS | 跨域处理 |
| 操作日志 | 记录用户操作 |
| 请求日志 | 请求访问日志 |
| 限流 | 令牌桶算法防刷 |

### 安全特性

- bcrypt 密码哈希 / XSS 过滤 / 参数化查询防注入
- 限流中间件 / CORS 配置 / 登录失败锁定
- 极验 (Geetest) 行为验证

## API 路由

```text
/api/v1/
├── public/              # 公共 (登录/注册/密码重置/刷新Token)
├── user/                # 用户 (个人信息/修改密码/路由)
├── admin/               # 管理 (仪表盘/用户CRUD/日志/邮件模板)
└── */                   # 插件 (自动注册)
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | 数据库连接 | 127.0.0.1:3306 |
| `JWT_SECRET` | JWT 签名密钥 | your_jwt_secret |
| `ENABLE_SWAGGER` | 启用 Swagger | true |
| `GO_ENV` | 运行环境 (production 跳过 Swagger 自动更新) | development |
| `GEETEST_ENABLED` | 启用极验验证码 | true |
| `ADMIN_PATH` | 管理后台路由前缀 | /system-mgr |

完整配置参考 `.env.example` 和 [doc/配置系统.md](doc/配置系统.md)。

## 文档导航

| 文档 | 说明 |
|------|------|
| [doc/JWT认证.md](doc/JWT认证.md) | Token 生成与验证 |
| [doc/邮件系统.md](doc/邮件系统.md) | 邮件发送与模板管理 |
| [doc/插件系统.md](doc/插件系统.md) | 插件开发指南 |
| [doc/数据库模型.md](doc/数据库模型.md) | 数据模型与操作 |
| [doc/配置系统.md](doc/配置系统.md) | 配置管理 |
| [doc/API路由.md](doc/API路由.md) | 路由定义规则 |
| [doc/前端请求.md](doc/前端请求.md) | 请求封装与 API 调用 |
| [doc/架构方案.md](doc/架构方案.md) | 高扩展性 MVC 插件架构设计 |
| [doc/架构概览.md](doc/架构概览.md) | 全栈架构深度解析 |

## 更新日志

### 2026-02-24

- 管理端侧边栏"调试"入口移除，功能收敛至"系统设置"
- 系统监控新增内存详情卡片、磁盘使用率展示优化、网络卡片合并

### 2026-02-21

- 插件自动发现/注册机制 + 独立注册表 `pkg/pluginregistry`
- Swagger 启动时自动更新 + Bearer 认证支持
- 控制器三层架构重构 (public/user/admin)
- 邮件模板管理 (后端 + 前端页面)
- 请求日志中间件 + 接口限流中间件
- Plugin 接口扩展 (完整生命周期、依赖解析、优雅关闭)

### 2026-02-04

- 项目文档体系建立 (邮件系统/JWT 认证/插件系统/数据库模型)

## 后续可扩展

- 邮件发送记录管理
- 系统配置可视化
- 更多插件开发

## 开源协议

[MIT](LICENSE)

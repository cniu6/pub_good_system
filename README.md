# FST - 全栈后台管理系统模板

FST (Full Stack Template) 是一个基于 Go (Gin) 和 Vue 3 (Naive UI) 构建的高性能、轻量级全栈后台管理系统模板。前后端分离架构，支持插件化扩展，开箱即用。

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go 1.24+ · Gin · GORM（MySQL / SQLite / Postgres）· JWT · Swagger |
| 前端 | Vue 3 · TypeScript · Vite · Naive UI · UnoCSS · Pinia · Alova |
| 构建 | Windows / Linux · amd64 / arm64 |

## 项目结构

```text
fst/
├── main.go / main_embedded.go   # 入口薄壳（开发 / 嵌入前端）
├── backend/
│   ├── cmd/server/              # 进程薄壳：前端托管 + Listen
│   ├── internal/
│   │   ├── appinit/             # 启动编排
│   │   ├── migrate/             # 数据库自迁移
│   │   └── task/                # 自动任务管理器（GORM + 内存调度）
│   ├── app/                     # 业务 MVC + plugins
│   ├── pkg/                     # config / db / middleware / presence
│   ├── routes/                  # public / user / admin / ws
│   ├── tests/README.md          # 测试放置说明
│   └── 留档.md
├── frontend/                    # Vue；自有 .env / .env.dev / .env.production
├── doc/
├── tools/                       # 运维/诊断/集成脚本（非服务入口）
├── build/
├── .env.example                 # 后端唯一示例（本地复制为根 .env，勿用 backend/.env）
├── dev.bat / build.bat / test.bat
└── README.md
```

**分层**：`pkg` = 零件；`internal` = 开机骨架；`app` = 业务 API；运维脚本只放 `tools/`。

### 自动任务

- 表：`auto_job_definitions` / `auto_job_runs`
- 核心：`backend/internal/task`
- 管理 API：`/api/v1/{ADMIN_API_PATH}/auto-jobs/*`（启动时空表会导入默认任务）

### 测试放哪

| 类型 | 位置 |
|------|------|
| 单元测试 `*_test.go` | **必须**与源码同包目录（Go 规范），如 `backend/app/services/*_test.go` |
| 运维/集成脚本 | 根目录 `tools/`（见 `tools/留档.md`） |

```bash
go test ./backend/app/services/ ./backend/utils/ ./backend/pkg/... -count=1
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

### 3. 数据库与配置

1. 创建数据库 `fst_platform`（或本地用 `DB_DRIVER=sqlite`）
2. 在**仓库根目录**复制 `.env.example` → `.env`，改库连接；**不要**再写 `backend/.env`
3. 前端另配 `frontend/.env*`（Vite）；`VITE_ADMIN_BASE_PATH` 与后端 `ADMIN_PATH` 对齐
4. 启动后自动迁移建表

`.env` 查找：exe 同级 → exe 上级 `../.env` → 从 cwd 向上（跳过前端纯 VITE 与 `backend/.env`）。打包请把 `.env` 放 exe 同级或上级。

### 4. 启动开发

```bash
# 必须在仓库根目录（读根 .env）
go run .

# 使用脚本
./dev.bat          # Windows：前端 + 根目录后端

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
前端建议开启 `VITE_AUTO_REFRESH_TOKEN=Y`；登出可走 `force-logout` 清理过期会话。

### 在线 Presence

WebSocket 心跳上报在线状态；管理端「在线用户」可查看/踢下线。见 [doc/在线会话与Presence.md](doc/在线会话与Presence.md)。

### Swagger 文档

启动时（dev）可自动检测代码变化并重新生成，自动包含插件 API。访问：`http://localhost:{PORT}/swagger/index.html`。  
管理端注解路径仍为 `/api/v1/admin/*`；若配置了自定义 `ADMIN_API_PATH`，`doc.json` **运行时改写**为实际前缀。

### 管理端两套路径

| 配置 | 含义 | 默认 |
|------|------|------|
| `ADMIN_PATH` + `VITE_ADMIN_BASE_PATH` | 管理**页面**入口 | `/system-mgr` |
| `ADMIN_API_PATH`（前端运行时 app-config） | 管理 **REST** | `/admin` → `/api/v1/admin` |

详见 [doc/管理端路径与Swagger自适应.md](doc/管理端路径与Swagger自适应.md)。

### 控制器三层架构

```text
controllers/
├── public/     # 无需登录 (登录/注册/app-config/支付回调)
├── user/       # 需要登录 (资料/支付/实名/提现)
└── admin/      # 管理员 (用户/日志/设置/支付/…；API 前缀可配置)
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
├── public/                 # 公共 (登录/注册/app-config/geo/session/支付回调)
├── user/                   # 用户 (资料/支付/实名/提现)
├── system/                 # 系统状态
├── ws/presence             # 在线心跳 WebSocket
├── {ADMIN_API_PATH}/       # 管理端，默认 admin（含 online / 各类日志与模板）
└── 插件路由                # 自动注册
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_DRIVER` | mysql / postgres / sqlite | mysql |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | 数据库连接 | 127.0.0.1:3306 |
| `JWT_SECRET` | JWT 签名密钥 | 见 .env.example |
| `ENABLE_SWAGGER` | 启用 Swagger | true |
| `ENABLE_ADMIN_DEBUG` | 管理端 debug/pprof（生产强制关） | 非生产 true |
| `GO_ENV` / `APP_ENV` | 运行环境 | development |
| `GEETEST_ENABLED` | 启用极验验证码 | true |
| `ADMIN_PATH` | 管理**页面**入口 | /system-mgr |
| `ADMIN_API_PATH` | 管理 **REST** 前缀 | /admin |

完整配置参考根目录 `.env.example` 和 [doc/配置系统.md](doc/配置系统.md)。前端见 `frontend/.env.example`。

## 文档导航

| 文档 | 说明 |
|------|------|
| [doc/README.md](doc/README.md) | 知识库总索引 |
| [doc/文档索引与目录留档.md](doc/文档索引与目录留档.md) | 全仓留档清单 |
| [doc/配置系统.md](doc/配置系统.md) | 根 `.env` 加载与配置 |
| [doc/在线会话与Presence.md](doc/在线会话与Presence.md) | 在线心跳与强退 |
| [doc/管理端路径与Swagger自适应.md](doc/管理端路径与Swagger自适应.md) | 页面/API 路径分离与自适应 |
| [doc/JWT认证.md](doc/JWT认证.md) | Token 生成与验证 |
| [doc/邮件系统.md](doc/邮件系统.md) | 邮件发送与模板管理 |
| [doc/短信插件系统.md](doc/短信插件系统.md) | 短信插件 |
| [doc/支付订单系统.md](doc/支付订单系统.md) | 支付与回调 |
| [doc/实名认证接入说明.md](doc/实名认证接入说明.md) | 实名认证 |
| [doc/插件系统.md](doc/插件系统.md) | 插件开发 |
| [doc/数据库模型.md](doc/数据库模型.md) | 数据模型 |
| [doc/API路由.md](doc/API路由.md) | 路由规则 |
| [doc/前端请求.md](doc/前端请求.md) | 前端请求封装 |
| [backend/留档.md](backend/留档.md) / [frontend/留档.md](frontend/留档.md) | 目录级留档 |

## 更新日志

### 2026-07-20

- 统一根目录 `.env` 加载（exe 同级 / 上级 / cwd 向上），废弃 `backend/.env`
- Presence 在线用户、会话强退、国际手机号/geo、邮件代理与 SMTP 增强
- 管理端设置/日志/模板与 i18n、Iconify 离线、Token 自动刷新默认开启
- SQLite / GORM 适配与日志聚合等稳定性修复；文档与目录留档同步

### 2026-07-16

- 支付回调绑定校验与日志脱敏；代登录与 debug 高危接口收紧
- 全局配置并发安全（Clone/Update）
- `ADMIN_API_PATH` + app-config 注入前端管理 API 前缀；Swagger doc.json 运行时改写
- 各目录留档与 doc 知识库同步更新

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

## 本地测试与管理员/用户切换验证

### 后端基础测试

在项目根目录执行：

```powershell
cd C:\Users\Administrator\Desktop\codingfile\fst
go test ./backend/...
```

该命令会运行后端所有单元测试，其中包括针对 `authGuard` 的 JWT 与刷新 token 分离测试。

### 启动后端服务

用于手工验证管理员/用户切换时，在项目根目录执行：

```powershell
cd C:\Users\Administrator\Desktop\coding\codingfile\fst
go run ./main.go
```

默认监听在根目录 `.env` 的 `PORT`（示例见 `.env.example`），如端口占用可自行调整配置或停止旧进程。

### 管理员 / 用户双 token 行为手工验证示例

建议使用 Postman、Apifox 或浏览器 REST 插件，按以下步骤检查同一账号下管理员态和用户态是否真正隔离：

1. **管理员 guard 登录**  
    - 方法：`POST`  
    - URL：`/api/v1/public/login`  
    - Body（JSON）：
      ```json
      {
         "userName": "你的管理员用户名",
         "password": "管理员密码",
         "authGuard": "admin"
      }
      ```
    - 期望：响应中的 `data.accessToken`、`data.refreshToken`、`data.id` 均有值。

2. **管理员 token 访问 admin / user 路由**  
    - Header：`Authorization: Bearer {adminAccessToken}`  
    - `GET /api/v1/admin/dashboard` 应成功返回（管理员接口可用）。  
    - `GET /api/v1/user/profile` 应返回 401/403 或业务错误码（管理员 token 不能假装用户）。

3. **刷新管理员 guard token**  
    - 方法：`POST`  
    - URL：`/api/v1/public/refresh-token`  
    - Body：
      ```json
      {
         "refreshToken": "上一步拿到的管理员 refreshToken",
         "authGuard": "admin"
      }
      ```
    - 期望：获得新的管理员 access/refresh token，旧会话仍保持有效期内可用。

4. **管理员 login-as 自己生成用户 guard 会话**  
    - 方法：`POST`  
    - URL：`/api/v1/admin/users/{adminId}/login-as`（`adminId` 是登录响应里的 `data.id`）  
    - Header：`Authorization: Bearer {adminAccessToken}`（使用最新的管理员 access token）  
    - 期望：响应包含 `token`（用户态 accessToken）和 `refreshToken`（用户态 refreshToken），用于 user guard。

5. **用户 token 访问 user / admin 路由**  
    - Header：`Authorization: Bearer {userAccessToken}`  
    - `GET /api/v1/user/profile` 应成功（用户接口可用）。  
    - `GET /api/v1/admin/dashboard` 应被拒绝（401/403 或业务错误码），说明用户 token 不能冒充管理员。

6. **刷新用户 guard token**  
    - 方法：`POST`  
    - URL：`/api/v1/public/refresh-token`  
    - Body：
      ```json
      {
         "refreshToken": "login-as 返回的用户 refreshToken",
         "authGuard": "user"
      }
      ```
    - 期望：成功获取新的用户 access/refresh token。

7. **确认管理员会话未被挤掉**  
    - 仍然使用管理员 access token 调用：`GET /api/v1/admin/dashboard`。  
    - 期望：请求依然成功，说明 `user_sessions` 表中 admin/user 是两条独立会话，login-as 自己不会导致管理员后台掉线。

通过以上步骤，你可以在本地直观确认：管理员与用户共享同一账号时，也能保持两套 token 和会话各自独立、互不影响。

## 开源协议

[MIT](LICENSE)

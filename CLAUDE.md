# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

FST (Full Stack Template)：Go 1.24 (Gin) + Vue 3 (TypeScript / Vite / Naive UI / UnoCSS / Pinia / Alova) 前后端分离的后台管理系统模板，支持插件化扩展。Go module 名为 `fst`。

**与用户交流一律使用中文**（rules.md 要求）；解释原理/方案时尽量配 mermaid 图（语法自检、暗黑主题下清晰可见）。

## 常用命令

```bash
# 后端开发（必须在仓库根目录运行，读取根 .env）
go run .

# 前后端一起：扫插件 + 生成 Scalar + 启动前端与后端
./dev.bat            # 后端 :8080，前端 :9980，Scalar /scalar/index.html

# 前端单独
cd frontend && pnpm install && pnpm dev

# 前端 lint / 类型检查 / 构建（**改 frontend 后收工前必须 lint 通过**；详见下方「前端 lint 硬规则」）
cd frontend && pnpm lint       # eslint . && vue-tsc --noEmit —— 必跑，本地能 dev ≠ 通过
cd frontend && pnpm lint:fix  # 大批量风格问题先自动修，再跑 lint 复核
cd frontend && pnpm build      # 非极小改动建议再冒烟

# 后端测试
go test ./backend/...
go test ./backend/app/services/ ./backend/utils/ ./backend/pkg/... -count=1
go test ./backend/pkg/config/ -run DotEnv -count=1     # 运行单个测试

# 生产构建（交互选 embedded 单文件 / external 外置前端；交叉编译 Windows+Linux 到 build/）
./build.bat

# Scalar 手动生成（dev.bat 已自动做）
go run backend/app/plugins/gen_plugins.go
cd backend && swag init -g ../main.go -o docs --parseDependency --parseInternal
```

测试位置规范：单元测试 `*_test.go` **必须**与源码同包目录（Go 规范）；运维/集成/诊断脚本只放根目录 `tools/`，不进 backend。

## 架构

### 启动链与入口

```
main.go (开发, tag !embedded) / main_embedded.go (tag embedded, //go:embed dist/*)
  → backend/cmd/server.Start          # 进程薄壳：前端托管 + Listen
  → backend/internal/appinit          # Bootstrap: config → db → migrate → services → tasks
  →                                   # SetupHTTP: 全局中间件 → routes.SetupRoutes → 插件装载
  → backend/internal/migrate.RunAutoMigrate   # 启动时自动建表/补列
```

### 后端分层（业务 API 不进 internal）

- `backend/pkg/` — 可复用零件：config / db / middleware / presence / pluginregistry
- `backend/internal/` — 开机骨架：appinit（启动编排）、migrate（全部自迁移）、task（自动任务管理器，表 `auto_job_definitions` / `auto_job_runs`）
- `backend/app/` — 业务 MVC：controllers（**三层 public / user / admin**）+ models + services + plugins
- `backend/routes/` — SetupRoutes 汇总，按 public.go / user.go / admin.go / ws.go 分文件；WS 自行鉴权不挂 HTTP AuthMiddleware
- `backend/utils/` — Success/Fail 响应、JWT、邮件、手机区号等通用工具
- 数据库用 **GORM**（`db.DB *gorm.DB`），禁止绕过 models 层直接操作库；模型方法已封装软删除等约定

### 分层调用关系

```
HTTP → controllers → services → models → DB
                    ↘ utils / pkg/config
```

- **控制器**：参数绑定、权限上下文、响应封装
- **服务**：事务边界、跨模型规则（支付绑定、提现状态机等）；控制器不直接操作数据库
- **模型**：GORM CRUD 与字段映射，函数命名统一 `动词+类型+条件`（如 `GetUserByID`、`CreatePaymentOrder`），禁止业务里写死 CamelCase 列名；JSON/库列一律 snake_case

### 插件系统（全自动）

在 `backend/app/plugins/<name>/` 实现 Plugin 接口，`init()` 中 `pluginregistry.Register(...)`。根 main.go / main_embedded.go 的 `@plugins-start` ~ `@plugins-end` blank import 区域由 `gen_plugins.go` 自动扫描更新，勿手改。生命周期：`Configure → Init → Migrate → RegisterRoutes → Shutdown`。

当前业务插件：

| 插件 | 优先级 | 路由 | 说明 |
|------|--------|------|------|
| `sms` | 10 | `/api/v1/{ADMIN}/sms-send-test` | 短信发送测试 + 模板管理 |
| `pay_balance` | 50 | `/api/v1/user/payment/balance` | 余额充值支付通道 |

### 配置（关键约定）

- **只维护仓库根目录 `.env`**（参考 `.env.example`）；`backend/.env` 已废弃禁止再写。查找顺序：exe 同级 → `../.env` → 从 cwd 向上（跳过纯 VITE 文件）
- 前端另有 `frontend/.env / .env.dev / .env.production`（Vite 分层），与后端根 `.env` 分开维护；仅需人工对齐 `VITE_ADMIN_BASE_PATH` ↔ `ADMIN_PATH`
- `DB_DRIVER` 支持 mysql / sqlite / postgres（本地开发可用 sqlite，根目录 fst.db）。**数据访问统一 GORM**：换库只改 `DB_DRIVER` + DSN，详见下方「数据库（GORM）」
- 全局配置 `config.GlobalConfig` 并发安全（Clone/Update），运行时可被 system_settings 覆盖

**并发安全 API（必用）**：

| 函数 | 用途 |
|------|------|
| `CloneGlobalConfig()` | 只读快照，业务读配置优先用这个 |
| `UpdateGlobalConfig(fn)` | 写锁内更新（热更新） |
| `NormalizeAdminAPIPath(path)` | 规范化 API 前缀：`/` 开头、去尾斜杠、空则 `/admin` |
| `IsProductionMode()` | 是否生产环境 |
| `IsAdminDebugOpsEnabled()` | 是否允许 debug/pprof/重启/手动任务等 |

### 数据库（GORM，`backend/pkg/db`）

业务数据访问统一为 GORM（`Create` / `Save` / `Updates` / `Where` / `First` / `Clauses(clause.Locking...)` / `Transaction` / `clause.OnConflict`）。**不要**再引入手写方言层或 sqlx。

| 文件 | 作用 |
|------|------|
| `db.go` | `InitDB` → `gorm.Open`（mysql / glebarez-sqlite / postgres）；导出 `DB *gorm.DB`；`IsMySQL/IsSQLite/IsPostgres`；`CheckTable*` 走 Migrator |
| `gorm_helpers.go` | `ForUpdate`（SQLite 跳过行锁）、`MapGormNotFound`、`WithTx` |
| `errors.go` | `IsDuplicateKeyError`（MySQL 1062 / SQLite UNIQUE / PG 23505） |

**约定**：

- 时间字段继续 Unix `int64` / `*int64`（`create_time`/`update_time`/`delete_time`）；`system_settings` 的 `time.Time` 例外
- 软删三套并存，**不用**默认 `gorm.DeletedAt`：`delete_time IS NULL`、`verification_codes.is_deleted`、`announcements.deleted_at=0`
- **建表编排只在** `internal/migrate.RunAutoMigrate`：调 `db.DB.AutoMigrate(models.AllGormModels()...)` + 自动任务表，再跑种子与少量补丁（如实名 `cert_unique_key`）。`pkg/db` **不**负责迁移编排
- 换库：只改 `.env` 的 `DB_DRIVER` + DSN，无需改 models

SQLite 用 `github.com/glebarez/sqlite`（纯 Go，无 CGO）。Postgres 上生产前请用真实实例跑一遍 `RunAutoMigrate` + 支付/提现冒烟（integration：`FST_PG_DSN` + `-tags integration`）。

### 管理端两套路径（勿混淆）

| 配置 | 含义 | 默认 |
|------|------|------|
| `ADMIN_PATH` + 前端 `VITE_ADMIN_BASE_PATH` | 管理**页面**入口 | `/system-mgr` |
| `ADMIN_API_PATH`（前端经 app-config 运行时注入） | 管理 **REST** 前缀 | `/admin` → `/api/v1/admin` |

Scalar 注解仍写 `/api/v1/admin/*`；openapi.json 在运行时按 `ADMIN_API_PATH` 改写。

### 认证与在线

- JWT 双 Token（access + refresh）+ `authGuard` 区分 admin/user 两套会话：同一账号的管理员态与用户态在 `user_sessions` 中是独立会话，互不挤占（login-as 自己不会把管理员踢下线）
- 刷新令牌轮换带重用检测（重用即吊销该用户会话）
- Presence：WebSocket 心跳上报在线（`pkg/presence` + `routes/ws.go`），管理端 `/online/*` 查看/踢下线；公告发布经 Hub `BroadcastJSON` 推送

### 后端中间件

| 中间件 | 包 | 说明 |
|--------|-----|------|
| `AuthMiddleware` / `AuthMiddlewareForGuard` | `pkg/middleware` | JWT Bearer + X-Api-Key 验证；设置 username/userID/role/authGuard |
| `AdminOnly` | `pkg/middleware` | 检查 authGuard == "admin" && role == "admin" |
| `CorsMiddleware` | `pkg/middleware` | CORS 白名单，支持通配符 |
| `LoggerMiddleware` | `pkg/middleware` | 请求日志（可配置跳过路径） |
| `APIAccessLogMiddleware` | `pkg/middleware` | 每请求写 api_access_logs 表 |
| `DynamicGlobalRateLimitMiddleware` | `pkg/middleware` | 令牌桶，读 system_settings 运行时配置 |
| `DynamicAdminRateLimitMiddleware` | `pkg/middleware` | 管理端专用限流 |
| `RequireIdempotency` | `pkg/middleware` | X-Idempotency-Key 幂等控制 |
| `SimpleLogMiddleware` | `pkg/middleware` | 管理写操作审计日志 |
| `ScalarAdminPathRewriteMiddleware` | `pkg/middleware` | openapi.json 路径改写 |

### 后端工具函数（`backend/utils/`）

| 函数 | 说明 |
|------|------|
| `Success(c, data)` | 统一成功响应 `{code:200, message:"OK", data}` |
| `Fail(c, code, message)` | 统一失败响应；code 400–599 同时作为 HTTP 状态码 |
| `GenerateToken / GenerateRefreshToken` | JWT 签发 |
| `ParseToken / ParseTokenIgnoreExpiry` | JWT 解析（后者忽略过期，用于强退） |
| `SendEmail / SendEmailWithTemplate` | 邮件发送（支持代理） |
| `HashPassword / CheckPassword` | bcrypt 密码 |
| `NormalizePhone / FormatPhoneDisplay` | 手机号规范化/显示 |
| `MaskCertificateNo` | 证件号掩码 |
| `VerifyGeetest` | 极验验证码校验 |
| `GetClientIP / GetUserAgent / GetRequestContext` | 请求上下文提取 |

### 自动任务（`backend/internal/task/`）

- 表：`auto_job_definitions` / `auto_job_runs`
- 管理 API：`/api/v1/{ADMIN}/auto-jobs/*`
- 默认任务：`prune_auto_job_runs`(1h)、`mark_stuck_auto_jobs`(180s)、`cleanup_expired_idempotency`(1h)、`cleanup_expired_orders`(120s)、`cleanup_sessions_codes`(600s)
- 新增任务：在 `handlers.go` 加一行 map + 函数，并在 `seed.go` 加预设
- 运行要点：空跑(Quiet)不落库；running 状态看定义表而非 runs 表；id 逼近上限自动重编号；handler 与调度均有 panic 兜底

### API 路由树

```text
/scalar/*any          # 可选 EnableScalar + 管理路径改写中间件
/api/v1/
├── public/            # 无需登录：登录注册、app-config、geo、session 强退、支付回调
├── user/              # 需 user 或 admin token：资料/支付/实名/提现/公告
├── system/            # 需登录：cleanup-status
├── ws/presence        # Presence WebSocket（JWT，自行鉴权）
├── {ADMIN_API_PATH}/  # 默认 /admin；需 admin token + AdminOnly + 动态限流
│   ├── dashboard / users / online / money-logs / score-logs
│   ├── logs / api-logs / email-logs / sms-logs
│   ├── settings / email-templates / sms-templates
│   ├── payment / pay-gateways / withdraw / realname
│   ├── announcements / auto-jobs / server-management
│   └── debug/*        # 仅 IsAdminDebugOpsEnabled 时注册
└── 插件路由           # pluginregistry 自动注册
```

## 接口协议硬规则（防回归，提交前自检）

- 业务响应统一 `{code, message, data}`，只用 `utils.Success(c, data)` / `utils.Fail(c, code, message)`；**禁止**业务接口直接 `c.JSON(401/403/404/500, ...)`，**禁止** `{"error": "..."}` 结构。鉴权/权限/路由未命中同样走统一协议
- `utils.Fail` 的业务码落在 400–599 时会同时作为真实 HTTP 状态码；HTTP 200 才是业务成功
- 前端判断鉴权过期**必须以业务码 `code===401` 为主**，HTTP 状态码判断只能作标注了「兼容分支」的次级逻辑
- 自检：后端 grep 不到业务代码里的 `gin.H{"error"` 与 `c.JSON(401/403/404/500`

## 编码规范

- 数据库表名/字段名一律小写+下划线，**禁止驼峰**；Go 函数可驼峰；前后端交互的 JSON 字段用 snake_case（`page_size`、`user_id`）
- 最小化修改，尽量不动其他模块；优先复用已有封装（models 方法、utils、前端 `src/service/api/` 的 API 函数），不重复造轮子；修改文件用局部编辑而非整文件重写
- 修 Bug 流程（rules.md）：理解问题 → 分析至少两种可能原因 → 制定计划 → **动手前向用户确认** → 执行 → 自查 → 解释
- **禁止擅自加 GitHub Actions / CI 工作流**（仓库不维护 `.github/workflows`）；用户未明确要求时不要创建或恢复 CI

## 前端 lint 硬规则（防回归，改 frontend 必守）

修改 `frontend/` 下任意代码（修 bug、改组件、改 API、改样式、改 i18n）后，**收工前必须**：

```bash
cd frontend && pnpm lint
```

等价于 `eslint . && vue-tsc --noEmit`。规则细则见 `.cursor/rules/frontend-lint-required.mdc`。

硬性要求：

1. **不要只跑 `pnpm dev` 或肉眼看页面** —— 本地能跑 ≠ lint 通过
2. lint 报错必须当场修完再结束任务；不要把「只是风格问题」留给用户
3. 改动面大时先 `pnpm lint:fix`，再 `pnpm lint` 复核
4. 非极小改动（单行文案除外）建议再 `pnpm build` 冒烟

常见失败类型（修前对照）：import/组件名排序（`perfectionist`）、`defineProps` 须紧跟 import/类型定义、模板标签内文本换行、JSDoc `@param` 须覆盖函数全部参数、文件末尾多余空行、未使用变量。

## 前端管理端开发规范（Naive UI / 侧边栏 / i18n）

### 新增管理端页面的完整清单（缺一不可）

1. **视图**：`frontend/src/views/admin/<模块>/index.vue`（详情页同目录放 `detail.vue`）
2. **路由 = 侧边栏**：在 `src/router/admin.routes.ts` 的 `getAdminRoutes()` 里注册（组件一律懒加载 `() => import(...)`）。侧边栏完全由这些路由的 meta 生成（`store/router/helper.ts` 的 `createAdminMenus`），没有独立的菜单配置文件：
   - 分组目录：`component: PassThrough` + `meta.menuType: 'dir'` + `redirect: { name: 首个子页 }`，子页面放 `children`（现有分组：仪表盘 / 用户管理 / 财务中心 / 系统设置，优先挂进已有分组）
   - 普通页面 meta：`title: 'route.xxx'` + `icon: 'icon-park-outline:xxx'`
   - 详情页/不进侧边栏的页面：`hide: true` + `activeMenu: '/父菜单路径'`（维持侧边栏高亮，参考 `admin-user-detail`）
   - 新增侧边栏项前检查是否与已有入口重复；已并入「系统设置」内层 Tab 的功能（邮件/短信模板等）不再单独加侧边栏，路由保留 `hide: true` 仅供深链
3. **i18n（zh_CN.json 与 en_US.json 都要，且是两组 key）**：
   - `route.<路由name>`（如 `"admin-users": "用户列表"`）：侧边栏/面包屑/Tab 用 `$t('route.<路由name>')` 解析，**漏了会直接显示 fallback 原文**
   - `meta.title` 指向的语义 key（如 `route.userList`）：路由守卫设置浏览器标题用
   - 页面内文案放 `adminXxx` 命名空间（参考 `adminOnlineUsers`），模板里一律 `t('adminXxx.yyy')`，禁止硬编码中文（`.cursor/rules/admin-i18n-required.mdc`）
4. **API 模块**：`src/service/api/admin/<模块>.ts`
   - URL 用 `getAdminApiBase()` / `adminApiUrl()`（来自 `admin/base.ts`）在**请求时**拼接，禁止硬编码 `/api/v1/admin`、禁止模块顶层 const 固化（前缀由后端 app-config 运行时注入）
   - 定义请求/响应的 TS interface，字段 snake_case 与后端对齐
   - 在 `admin/index.ts` 的 `adminApi` 懒加载代理中注册新模块
5. **代码隔离**：管理端路由/视图/API 被打进独立 chunk（`assets/m/`，管理员登录后才动态加载）。管理端专属代码只放 `views/admin/**`、`router/admin.routes.ts`、`service/api/admin/**`，用户端代码不得 import 管理端模块

### Naive UI 使用规则

- UI 一律用 Naive UI 组件（`n-card` / `n-data-table` / `n-space` / `n-grid` / `n-alert` …），不引入其他 UI 库、不手写原生控件
- 模板中的 `n-*` 组件由 `NaiveUiResolver` 自动按需引入**无需 import**；但 `h()` 渲染函数里（如表格 `columns` 的 `render`）用到的组件必须显式 `import { NButton, NTag, NSpace } from 'naive-ui'`
- `useMessage` / `useDialog` / `useNotification` / `useModal` 及 vue / vue-router / pinia / vue-i18n 的 API（含 `useI18n`）已配置 AutoImport，无需手动 import
- 图标统一用 `icon-park-outline` 集合（已离线注册，`modules/iconify-offline.ts`）：路由 icon 写字符串 `'icon-park-outline:xxx'`，代码中渲染用 `renderIcon` / `NovaIcon`
- 列表页固定套路（模板参考 `views/admin/online-users/index.vue`）：`n-card` 内筛选 `n-space`（`n-input` 支持回车搜索 + `n-select` 筛选）+ `remote` 的 `n-data-table`；`reactive` 的 `query { page, page_size }` 与 `pagination { page, pageSize, itemCount }` 同步翻页；筛选下拉必须显式给「全部」项（`value: ''`），不留空占位
- 危险操作（删除/踢下线）必须 `dialog.warning` 二次确认，按钮文案用 `t('common.confirm')` / `t('common.cancel')`；`common.*` 命名空间已有确认/取消/刷新/重置等通用文案，优先复用
- 请求结果统一判 `res.isSuccess`（alova 封装注入），失败 `message.error(res.message || t('...'))`，加载态走 `loading` ref；优先复用 `components/common/` 现有封装（`PhoneInput`、`TableColumnSelector`、`GeetestCaptcha`、`NovaIcon`、`AnnouncementPreviewModal`、`I18nMemoEditor` 等）

### 前端启动顺序（务必遵守）

```
createApp
  → installPinia
  → settingsStore.loadConfig()   // 注入 admin_api_path
  → i18n / installRouter(mode)
  → directives / assets
  → mount
```

- **admin 模式**：`authStorage.enableSessionIsolation()`（sessionStorage 与用户端 localStorage 隔离）
- **页面入口**：`VITE_ADMIN_BASE_PATH`
- **REST 前缀**：运行时 `admin_api_path`，见 `service/api/admin/base.ts`

### 前端关键目录与模块

| 目录 | 说明 |
|------|------|
| `src/components/common/` | 通用组件：PhoneInput、NovaIcon、TableColumnSelector、GeetestCaptcha、I18nMemoEditor、AnnouncementPreviewModal 等 |
| `src/composables/` | 逻辑封装：usePresence（WS 心跳 + BroadcastChannel 选主 + 指数退避重连） |
| `src/modules/` | 初始化模块：iconify-offline（注册 icon-park-outline）、i18n |
| `src/store/auth.ts` | 认证：login/logout/refreshToken/admin-user 双 guard、login-as 窗口 |
| `src/store/settings.ts` | 运行时配置：loadConfig → fetchAppConfig → setRuntimeAdminApiPath |
| `src/store/app/index.ts` | UI 状态：主题/侧边栏宽度/语言/全屏/水印等 |
| `src/store/router/index.ts` | 路由：initAuthRoute → admin/user 模式分别加载路由 |
| `src/service/http/` | Alova 封装：request 实例、拦截器、handleServiceResult、token-refresh（单例去重） |
| `src/service/api/admin/` | 管理端 API（懒加载代理 `adminApi`，16 个子模块） |
| `src/service/api/user/` | 用户端 API（login、user-center、realname、announcement、logs） |
| `src/service/api/app-config.ts` | 公开配置获取（含 admin_api_path 运行时注入） |
| `src/layouts/` | 主布局：ProLayout + 侧边栏拖拽调宽 + 弹窗公告 + 移动端适配 |
| `locales/` | i18n：zh_CN.json / en_US.json；命名空间见下方 |

### 前端 i18n 命名空间速查

| 命名空间 | 用途 | 示例 |
|----------|------|------|
| `common.*` | 通用操作（确认/取消/刷新/重置/启用/禁用/编辑/删除等） | `t('common.confirm')` |
| `route.*` | 路由标题=侧边栏/面包屑/Tab（`route.<路由name>`） | `t('route.admin-users')` |
| `login.*` | 登录/注册页 | `t('login.userName')` |
| `adminXxx.*` | 管理端页面文案（按模块建命名空间） | `t('adminOnlineUsers.kick')` |
| `route.xxx` | meta.title 语义 key（路由守卫设浏览器标题） | `route.userList` |

### 管理端 API 懒加载代理（`admin/index.ts`）

```typescript
// 使用方式：首次调用任意方法时才 import() 对应模块
import { adminApi } from '@/service/api/admin'
const res = await adminApi.user.list({ page: 1 })

// 已注册模块：user / log / apiLog / smsLog / emailLog / emailTemplate /
//   smsTemplate / debug / settings / server / autoJob / dashboard /
//   realname / payment / finance / announcement
```

新增 API 模块必须注册进此代理，否则无法通过 `adminApi.xxx` 访问。

## 文档同步（留档约定）

各目录有 `留档.md`（backend/、frontend/、tools/、backend/internal/task/ 等），`doc/` 为知识库（索引：`doc/README.md`、`doc/文档索引与目录留档.md`）。重大功能改动后应同步更新对应留档与 doc 文档。

### 留档与 doc 索引

| 文档 | 位置 | 内容 |
|------|------|------|
| 后端总留档 | `backend/留档.md` | 目录结构、近期更新、测试与工具 |
| 业务逻辑核心 | `backend/app/业务逻辑核心.md` | 分层调用、近期业务要点 |
| 插件系统 | `backend/app/plugins/插件系统接口与管理逻辑.md` | Plugin 接口、注册/装载、gen_plugins |
| 内部系统库 | `backend/internal/内部系统库.md` | appinit/migrate/task 说明 |
| API 路由 | `backend/routes/API路由定义与分发规则.md` | 路由树、管理端前缀、Scalar |
| 通用工具 | `backend/utils/通用工具函数库.md` | 响应/JWT/邮件/手机号/代理等 |
| 全局配置 | `backend/pkg/config/全局配置管理与环境加载.md` | Config 结构、并发安全 API、.env 查找 |
| 自动任务 | `backend/internal/task/留档.md` | handler/seed/运行要点/默认任务 |
| 前端总留档 | `frontend/留档.md` | 环境变量、目录结构、两套管理路径 |
| 前端源码核心 | `frontend/src/源代码核心.md` | 启动顺序、目录索引 |
| 前端网络请求 | `frontend/src/service/网络请求与API定义.md` | Alova 封装、admin API 路径注入 |
| 前端路由 | `frontend/src/router/路由配置与权限守卫.md` | 双模式、守卫、懒加载 |
| 前端组件 | `frontend/src/components/通用UI组件与业务封装.md` | 通用组件列表 |
| 工具脚本 | `tools/留档.md` | 运维/诊断脚本与单元测试的区别 |
| 配置系统 | `doc/配置系统.md` | .env 加载与配置 |
| 在线会话 | `doc/在线会话与Presence.md` | WS 在线心跳与强退 |
| 管理端路径 | `doc/管理端路径与Scalar自适应.md` | 页面/API 路径分离与自适应 |
| JWT 认证 | `doc/JWT认证.md` | Token 生成与验证 |
| 邮件系统 | `doc/邮件系统.md` | 邮件发送与模板管理 |
| 短信插件 | `doc/短信插件系统.md` | 短信多厂商 |
| 支付订单 | `doc/支付订单系统.md` | 支付与回调 |
| 提现流程 | `doc/提现流程与余额管理.md` | 提现状态机 |
| 实名认证 | `doc/实名认证接入说明.md` | 实名审核 |
| 插件开发 | `doc/插件系统.md` | 插件开发指南 |
| 数据库模型 | `doc/数据库模型.md` | 数据模型 |
| API 路由 | `doc/API路由.md` | 路由规则 |
| 前端请求 | `doc/前端请求.md` | 前端请求封装 |

## 已知设计（勿当 Bug"修复"）

- SQLite 库中 `verification_codes` 同时存在 `email` 与 `contact` 列属预期（SQLite 不可靠改名，采用加列拷贝并保留旧列），业务只读写 `contact`
- `utils.Fail` 的业务码 400–599 同时作为 HTTP 状态码是设计意图（让网关/中间件按 c.Writer.Status() 统计 4xx/5xx 准确）
- 支付通道密钥管理端列表/详情掩码，更新时 `***` 不覆盖真密钥——不是脱敏失败
- **JWT ≠ API Key**：JWT 只服务人登录会话（Bearer + `token_type`/`auth_guard`）；API Key 只服务程序调用（`X-Api-Key`）。二者用途不同，互不替代。详见 `backend/留档.md`「鉴权两套」
- JWT 强制 `token_type` + `auth_guard`：升级后旧 Token 失效需重新登录——预期，不是鉴权回归
- API Key 库内存明文 + 末4位 hint；管理端列表/详情只下发掩码，用户中心可随时查看明文；启动时会把旧 SHA256 哈希密钥自动重置为新明文（旧 Key 失效需换新——预期，与登录 JWT 无关）
- 管理端数据库控制台（`/db/*`）是正式管理功能，不要用 debug 运维开关关掉；只读 SQL 不得放行可写 `WITH`/`PRAGMA`

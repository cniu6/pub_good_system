# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

FST (Full Stack Template)：Go 1.24 (Gin) + Vue 3 (TypeScript / Vite / Naive UI / UnoCSS / Pinia / Alova) 前后端分离的后台管理系统模板，支持插件化扩展。Go module 名为 `fst`。

**与用户交流一律使用中文**（rules.md 要求）；解释原理/方案时尽量配 mermaid 图（语法自检、暗黑主题下清晰可见）。

## 常用命令

```bash
# 后端开发（必须在仓库根目录运行，读取根 .env）
go run .

# 前后端一起：扫插件 + 生成 Swagger + 启动前端与后端
./dev.bat            # 后端 :8080，前端 :9980，Swagger /swagger/index.html

# 前端单独
cd frontend && pnpm install && pnpm dev

# 前端 lint / 类型检查 / 构建（改完代码跑 lint:fix；非小改动跑 build 验证）
cd frontend && pnpm lint       # eslint + vue-tsc --noEmit
cd frontend && pnpm lint:fix
cd frontend && pnpm build

# 后端测试
go test ./backend/...
go test ./backend/app/services/ ./backend/utils/ ./backend/pkg/... -count=1
go test ./backend/pkg/config/ -run DotEnv -count=1     # 运行单个测试

# 生产构建（交互选 embedded 单文件 / external 外置前端；交叉编译 Windows+Linux 到 build/）
./build.bat

# Swagger 手动生成（dev.bat 已自动做）
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
- 数据库用 sqlx，禁止绕过 models 层手拼 SQL；模型方法已封装软删除等约定

### 插件系统（全自动）

在 `backend/app/plugins/<name>/` 实现 Plugin 接口，`init()` 中 `pluginregistry.Register(...)`。根 main.go / main_embedded.go 的 `@plugins-start` ~ `@plugins-end` blank import 区域由 `gen_plugins.go` 自动扫描更新，勿手改。生命周期：`Configure → Init → Migrate → RegisterRoutes → Shutdown`；demo 插件需 `-tags demo`。

### 配置（关键约定）

- **只维护仓库根目录 `.env`**（参考 `.env.example`）；`backend/.env` 已废弃禁止再写。查找顺序：exe 同级 → `../.env` → 从 cwd 向上（跳过纯 VITE 文件）
- 前端另有 `frontend/.env / .env.dev / .env.production`（Vite 分层），与后端根 `.env` 分开维护；仅需人工对齐 `VITE_ADMIN_BASE_PATH` ↔ `ADMIN_PATH`
- `DB_DRIVER` 支持 mysql / postgres / sqlite（本地开发可用 sqlite，根目录 fst.db）
- 全局配置 `config.GlobalConfig` 并发安全（Clone/Update），运行时可被 system_settings 覆盖

### 管理端两套路径（勿混淆）

| 配置 | 含义 | 默认 |
|------|------|------|
| `ADMIN_PATH` + 前端 `VITE_ADMIN_BASE_PATH` | 管理**页面**入口 | `/system-mgr` |
| `ADMIN_API_PATH`（前端经 app-config 运行时注入） | 管理 **REST** 前缀 | `/admin` → `/api/v1/admin` |

Swagger 注解仍写 `/api/v1/admin/*`；doc.json 在运行时按 `ADMIN_API_PATH` 改写。

### 认证与在线

- JWT 双 Token（access + refresh）+ `authGuard` 区分 admin/user 两套会话：同一账号的管理员态与用户态在 `user_sessions` 中是独立会话，互不挤占（login-as 自己不会把管理员踢下线）
- 刷新令牌轮换带重用检测（重用即吊销该用户会话）
- Presence：WebSocket 心跳上报在线（`pkg/presence` + `routes/ws.go`），管理端 `/online/*` 查看/踢下线；公告发布经 Hub `BroadcastJSON` 推送

## 接口协议硬规则（防回归，提交前自检）

- 业务响应统一 `{code, message, data}`，只用 `utils.Success(c, data)` / `utils.Fail(c, code, message)`；**禁止**业务接口直接 `c.JSON(401/403/404/500, ...)`，**禁止** `{"error": "..."}` 结构。鉴权/权限/路由未命中同样走统一协议
- `utils.Fail` 的业务码落在 400–599 时会同时作为真实 HTTP 状态码；HTTP 200 才是业务成功
- 前端判断鉴权过期**必须以业务码 `code===401` 为主**，HTTP 状态码判断只能作标注了「兼容分支」的次级逻辑
- 自检：后端 grep 不到业务代码里的 `gin.H{"error"` 与 `c.JSON(401/403/404/500`

## 编码规范

- 数据库表名/字段名一律小写+下划线，**禁止驼峰**；Go 函数可驼峰；前后端交互的 JSON 字段用 snake_case（`page_size`、`user_id`）
- 最小化修改，尽量不动其他模块；优先复用已有封装（models 方法、utils、前端 `src/service/api/` 的 API 函数），不重复造轮子；修改文件用局部编辑而非整文件重写
- 修 Bug 流程（rules.md）：理解问题 → 分析至少两种可能原因 → 制定计划 → **动手前向用户确认** → 执行 → 自查 → 解释

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
- 请求结果统一判 `res.isSuccess`（alova 封装注入），失败 `message.error(res.message || t('...'))`，加载态走 `loading` ref；优先复用 `components/common/` 现有封装（`PhoneInput`、`TableColumnSelector`、`GeetestCaptcha`、`NovaIcon`、`AnnouncementPreviewModal` 等）

## 文档同步（留档约定）

各目录有 `留档.md`（backend/、frontend/、tools/、backend/internal/task/ 等），`doc/` 为知识库（索引：`doc/README.md`、`doc/文档索引与目录留档.md`）。重大功能改动后应同步更新对应留档与 doc 文档。

## 已知设计（勿当 Bug"修复"）

- SQLite 库中 `verification_codes` 同时存在 `email` 与 `contact` 列属预期（SQLite 不可靠改名，采用加列拷贝并保留旧列），业务只读写 `contact`
- `frontend/agents.md` 的 Stack 描述（shadcn-vue/tailwind 等）已过时，以 package.json 为准（Naive UI + UnoCSS + Alova）

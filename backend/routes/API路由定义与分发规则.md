# API 路由定义与分发规则 (Routes)

> 路径：`backend/routes/`（`routes.go` 汇总，`public.go` / `user.go` / `admin.go` / `ws.go` 分文件）  
> **最后更新**：2026-07-20

## 简介

`SetupRoutes` 为全站 API 注册入口：公共、用户、系统、管理端、Presence WS；插件路由在 `appinit.SetupHTTP` 里挂载。

## 路由树（当前）

```text
/swagger/*any          # 可选 EnableSwagger + 管理路径改写中间件
/api/v1/
├── public/            # 无需登录：登录注册、app-config、geo、session 强退、支付回调
├── user/              # 需 user 或 admin token：资料/支付/实名/提现
├── system/            # 需登录：cleanup-status 等
├── ws/presence        # Presence WebSocket（JWT）
├── {ADMIN_API_PATH}/  # 默认 /admin；需 admin token + AdminOnly + 动态限流
│   ├── dashboard
│   ├── users / online / money-logs / score-logs
│   ├── logs / api-logs
│   ├── settings / email-templates / email-logs
│   ├── sms-templates / sms-logs
│   ├── payment / payment/gateways / withdraw / realname
│   ├── db/*              # 数据库控制台：表/数据/结构/DDL；生产环境强制只读
│   ├── generate-nos
│   └── debug/*        # 仅 IsAdminDebugOpsEnabled 时注册
└── 插件路由           # pluginregistry 自动注册
```

公开补充：`/public/geo/*`（区号/地理探测）、`/public/session/force-logout`（过期 token 亦可吊销会话）。  
在线管理：`/online/stats`、`/online/sessions`。详见 [doc/在线会话与Presence.md](../../doc/在线会话与Presence.md)。

## 管理端前缀

```go
adminAPIPath := config.NormalizeAdminAPIPath(cfg.AdminAPIPath) // 默认 /admin
adminGroup := v1.Group(adminAPIPath)
```

- **页面路径** `ADMIN_PATH` 不在此注册，由前端 history/hash 入口处理。
- 改 `ADMIN_API_PATH` 后：路由、app-config、Swagger doc、前端 `getAdminApiBase()` 一致。

## Swagger

```go
router.GET("/swagger/*any",
  middleware.SwaggerAdminPathRewriteMiddleware(),
  ginSwagger.WrapHandler(swaggerFiles.Handler),
)
```

注解路径保持 `/api/v1/admin/...`；非默认前缀时 doc.json 运行时改写。

## 规范

- 新接口按 public / user / admin 分层注册。
- 管理写操作可挂 `SimpleLogMiddleware` 做操作审计。
- 支付回调必须在 public，且服务层做签名与订单绑定校验。

## 数据库控制台

`/db/*` 仅允许管理员访问，提供表列表、分页数据预览、字段/索引/联合索引/外键元数据和 DDL 查看。

- 生产环境固定只读，禁用写 SQL、行级 CRUD 与 SQLite 备份下载。
- 非生产环境由 `ENABLE_ADMIN_DB_WRITE` 控制受限写能力；仅允许单条 `INSERT` 或带 `WHERE` 的 `UPDATE`/`DELETE`，写入使用事务与影响行数上限。
- 在线字段、索引、表结构变更不开放；结构变更仍由 `internal/migrate` 统一管理。

# 网络请求与 API 定义 (Service)

> 路径：`frontend/src/service/`  
> **最后更新**：2026-07-16

## 简介

基于 Alova 的请求封装：Token 注入、刷新、业务错误处理。API 按 `public` / `user` / `admin` 分目录。

## 管理端 API 路径（运行时注入）

**文件**：`service/api/admin/base.ts`

| 函数 | 说明 |
|------|------|
| `normalizeAdminApiPath` | 规范化前缀（`/` 开头、去尾斜杠、默认 `/admin`） |
| `setRuntimeAdminApiPath` | bootstrap 从 app-config 注入 |
| `getAdminApiPath` | 当前生效前缀 |
| `getAdminApiBase` | `/api/v1` + 前缀，如 `/api/v1/admin` |
| `adminApiUrl(sub)` | 拼接子路径 |

**注入时机**：`bootstrap` → `settingsStore.loadConfig()` → `fetchAppConfig()` → `setRuntimeAdminApiPath(data.admin_api_path)`。

**回退**：app-config 失败时用 `VITE_ADMIN_API_PATH`。

**注意**：各 admin API 文件用 `function baseUrl() { return \`${getAdminApiBase()}/...\` }`，**禁止**模块顶层 `const BASE_URL = getAdminApiBase()...` 固化（否则注入晚于 import 会失效）。

## 页面入口 vs API 前缀

| 变量 | 用途 |
|------|------|
| `VITE_ADMIN_BASE_PATH` / 后端 `ADMIN_PATH` | 管理端**页面** URL，如 `/system-mgr` |
| `admin_api_path` / `ADMIN_API_PATH` | 管理端 **REST** 前缀 |
| `VITE_ADMIN_API_PATH` | 仅回退 |

## 目录

```text
service/
├── http/                 # Alova 实例、拦截器、错误处理
└── api/
    ├── app-config.ts     # 公开配置（含 admin_api_path）
    ├── admin/            # 管理端（user/settings/payment/logs/...）
    │   └── base.ts
    ├── user/             # 用户端
    └── system.ts 等
```

## 管理端模块（请求时拼 base）

- `user.ts` / `dashboard.ts` / `settings.ts` / `payment.ts` / `paygateway.ts`
- `log.ts` / `api-log.ts` / `email-log.ts` / `sms-log.ts` / `email-template.ts`
- `finance.ts` / `realname.ts` / `debug.ts` / `server.ts`

## 认证存储

管理端 bootstrap 启用 `authStorage` session 隔离，避免与用户端 localStorage token 互相覆盖。设置页等读 token 用 `authStorage.get('accessToken')`，不要混用裸 `local.get`。

## 规范

- 新接口放在对应 `api/` 子目录，泛型标注返回类型。
- 管理端一律 `getAdminApiBase()`，不要写死 `/api/v1/admin`。

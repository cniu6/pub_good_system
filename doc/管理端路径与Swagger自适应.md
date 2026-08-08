# 管理端路径与 Scalar 自适应

> **最后更新**：2026-07-16  
> **关联**：`ADMIN_PATH` / `ADMIN_API_PATH` / `VITE_ADMIN_*` / app-config / scalar 中间件

---

## 一句话

**页面入口**和 **REST 前缀**是两套配置，可独立修改。API 前缀改后端后，前端与 Scalar 自动对齐。

---

## 对照表

| 配置项 | 位置 | 作用 | 改后是否需重建前端 |
|--------|------|------|-------------------|
| `ADMIN_PATH` | 根 `.env` | 管理后台**页面**隐藏路径 | 是（还要改 `VITE_ADMIN_BASE_PATH`） |
| `VITE_ADMIN_BASE_PATH` | frontend `.env.*` | 与上一致，Vite 入口/base | 是 |
| `ADMIN_API_PATH` | 根 `.env` | `/api/v1` 下管理 REST 前缀 | **否**（运行时注入） |
| `VITE_ADMIN_API_PATH` | frontend `.env.*` | app-config 失败时的回退 | 建议保持一致，非必须每次重建 |

完整 URL 示例：

- 页面：`https://站点/system-mgr#/dashboard`
- API：`https://站点/api/v1/admin/users`（若 `ADMIN_API_PATH=/admin`）

---

## 后端行为

1. **路由**：`v1.Group(NormalizeAdminAPIPath(AdminAPIPath))`  
2. **app-config**：`admin_api_path` 字段  
3. **Scalar**：`ScalarAdminPathRewriteMiddleware` 改写 `openapi.json` 里 `/api/v1/admin` 前缀  
4. **限流 pprof 前缀**：随 `ADMIN_API_PATH` 变化  

控制器 Scalar 注解可继续写：

```go
// @Router /api/v1/admin/users [get]
```

非默认前缀时由中间件改写展示路径。

---

## 前端行为

1. `bootstrap` → `settingsStore.loadConfig()`  
2. `GET /api/v1/public/app-config`  
3. `setRuntimeAdminApiPath(admin_api_path)`  
4. 所有管理请求 `getAdminApiBase()` → `/api/v1` + 前缀  

实现文件：

- `frontend/src/service/api/admin/base.ts`
- `frontend/src/store/settings.ts`
- `frontend/src/service/api/app-config.ts`

---

## 操作步骤

### 只换 API 前缀（例如改成 `/mgr-api`）

1. 根 `.env`：`ADMIN_API_PATH=/mgr-api`  
2. 重启后端  
3. 验证：`/api/v1/public/app-config` 中 `admin_api_path`  
4. 验证：管理端请求是否变为 `/api/v1/mgr-api/...`  
5. 验证：`/scalar/index.html` 文档路径是否已改写  

### 只换页面入口（例如改成 `/console`）

1. 根 `.env`：`ADMIN_PATH=/console`  
2. frontend：`VITE_ADMIN_BASE_PATH=/console`  
3. 重建前端并部署（embedded 则重新 `build.bat`）  

---

## 安全提示

- 页面路径用于“隐藏后台入口”，不要和公开文档写死的路径混用。  
- API 前缀不是安全边界；权限仍靠 JWT + AdminOnly。  
- 生产关闭 Scalar：`ENABLE_SWAGGER=false`。  
- 高危 debug：`ENABLE_ADMIN_DEBUG`，生产永远关。

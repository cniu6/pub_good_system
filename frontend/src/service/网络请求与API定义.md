# 网络请求与API定义 (Service) 深度流证

## 简介
Service 层是前端与后端通信的唯一桥梁，基于 Alova.js 构建，支持 Token 自动刷新及请求头动态注入。

## 核心组件详细流证

### 1. Alova 实例 (createAlovaInstance)
配置网络请求的基础行为。
- **状态钩子**: 使用 `VueHook`，适配 Vue 的响应式数据。
- **拦截器**:
  - **beforeRequest**:
    - 自动从本地缓存获取 `accessToken` 并注入 `Authorization` 头。
    - **极验增强**: 自动调用 `geetestManager` 注入合法的验证头信息。
    - **FormPost 支持**: 识别 `isFormPost` 标记并自动转换 `Content-Type` 及数据格式。
  - **responded**:
    - **onSuccess**:
      - 200 状态码下：
        - 处理 `isBlob` 下载场景。
        - 校验后端自定义状态码（默认 200 为成功）。
        - 统一调用 `handleServiceResult` 包装返回。
      - 业务失败：调用 `handleBusinessError` 并触发全局提示。
    - **Token 自动刷新 (handleRefreshToken)**:
      - 拦截器识别 401 状态码 -> 挂起当前请求 -> 调用 `/api/v1/updateToken` -> 更新本地 `accessToken` 和 `refreshToken` -> 重试之前挂起的请求。

### 2. 系统 API (system.ts)
- `fetchAllRoutes()`: 获取全站路由。返回 `AppRoute.RowRoute[]`。
- `fetchUserPage()`: 获取用户分页列表。返回 `Entity.User[]`。
- `fetchRoleList()`: 获取角色定义列表。返回 `Entity.Role[]`。

### 3. 认证 API (login.ts)
- `fetchLogin(data)`: 用户登录。提交 `userName`, `password`。
- `fetchUpdateToken(data)`: 刷新令牌。提交 `refreshToken`。
- `fetchUserRoutes(params)`: 获取指定用户的路由配置。
- `fetchSendRegisterCode(data)`: 发送注册验证码。提交 `email`, `lang`。
- `fetchRegister(data)`: 用户注册。提交注册表单数据。
- `fetchSendResetEmail(data)`: 发送重置密码邮件。提交 `email`, `lang`。
- `fetchResetPasswordConfirm(data)`: 确认重置密码。提交 `email`, `code`, `newPassword`。

### 4. 列表与数据 API (list.ts)
- 提供各类业务数据的增删改查接口（基于 `Entity` 类型定义）。

## 异常处理机制 (handle.ts)
- **handleResponseError**: 处理网络协议层错误（如 403, 404, 500）。
- **handleBusinessError**: 处理后端逻辑错误（如用户名已存在、余额不足）。
- **showError**: 全局统一的错误消息展示逻辑，自动过滤 `ERROR_NO_TIP_STATUS`。

## 开发规范
- 接口必须通过 `request.Get / Post` 等泛型方法定义返回值类型。
- 复杂的请求逻辑必须在 `api/` 目录下按业务模块拆分文件。

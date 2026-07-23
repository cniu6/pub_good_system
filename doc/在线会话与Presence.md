# 在线会话与 Presence（WebSocket）

> **最后更新**：2026-07-24  
> 关联：`backend/pkg/presence/` · `backend/routes/ws.go` · `frontend/src/composables/usePresence.ts`

## 作用

管理端可查看当前在线用户会话，并支持踢下线。前端登录后通过 WebSocket 心跳上报在线状态；同一**登录会话**的多标签由 `BroadcastChannel` 选出一个标签建立连接。

## 后端

| 路径 | 说明 |
|------|------|
| `POST /api/v1/system/ws-ticket` | 鉴权后签发短时（60 秒）、一次性 WebSocket ticket；单用户/guard 限流 12 次/分钟 |
| `GET /api/v1/ws/presence?ticket=...` | Presence WebSocket；ticket 消费一次后失效，CORS 与 WS 校验见中间件测试 |
| `GET /api/v1/{ADMIN}/online/stats` | 在线统计 |
| `GET /api/v1/{ADMIN}/online/sessions` | 会话列表 |
| `DELETE /api/v1/{ADMIN}/online/sessions/:id` | 吊销会话 |

核心包：`pkg/presence`（Hub、连接处理、限流）。会话落库/吊销与 `user_sessions` 联动：心跳会复核会话是否仍有效；业务广播前也会复核，已撤销/过期会话不能继续接收公告。

## 前端

| 文件 | 说明 |
|------|------|
| `composables/usePresence.ts` | 布局挂载后连 WS、心跳 |
| `utils/browserId.ts` | 浏览器级 ID，减少多 Tab 重复在线 |
| `views/admin/online-users/` | 管理端「在线用户」页（用户管理下） |
| `service/api/admin/online.ts` | 管理 API |

心跳间隔可由系统设置 `online_report_interval_seconds` 注入（默认 **30 秒**；前端 `usePresence` 默认与此对齐）。管理端「在线用户」页保存后会热更新当前会话 Presence。

列表按 **用户 + 登录端** 归并为一行，多设备挂在 `devices` 数组内展示；踢下线仍按单条会话 ID。

## 多标签连接规则（必要 / 不必要）

### 必要保留

- **选主按会话指纹分组**：指纹由当前 access token 派生。同一登录会话的多个标签只保留一个 leader；管理员本人和 login-as 的独立会话必须各自建立 Presence，不能互相抢主。
- **非 leader 不建连**：通过 `BroadcastChannel` 接收 leader 转发的公告和未读数事件。leader 断线时由自身按退避重连；leader 关闭、登出或心跳超时后，其它标签随机延迟接棒。
- **`browser_id` 只做后端会话收口，不参与选主**：同浏览器同用户/guard 的新会话会收口旧会话；不同用户、不同 guard 或已有 browser_id 的其它设备必须保留。
- **重连退避只处理临时失败**：WebSocket 意外关闭、网络异常、5xx/限流可指数退避，上限 30 秒；`ws-ticket` 返回 401 表示会话已无效，必须停止重连并本地退出。

### 不要做

- **不要按账号 ID 或 guard 粗粒度选主**：会把管理员本人和 login-as 的独立会话误判为同一组，导致其中一个账号没有在线状态。
- **不要把 `ws-ticket` 当轮询接口**：正常心跳在已建立的 WebSocket 内发送；只有首次连接或断线重连才需要新 ticket。
- **不要让所有账号无差别共用一条 WS**：WebSocket ticket 绑定具体用户会话；不同账号、管理端和用户端强行共用会造成实时消息丢失或串页。
- **不要对 401 做指数重试**：这只会产生 `ws-ticket` 日志风暴；应让认证层统一处理会话失效。

## 相关：过期 Token 强退

- 公开：`POST /api/v1/public/session/force-logout`  
  允许携带**已过期** access token（`ParseTokenForGuardIgnoreExpiry`），按 token hash 吊销会话。  
- 前端登出时会尽量调用，避免过期会话残留。

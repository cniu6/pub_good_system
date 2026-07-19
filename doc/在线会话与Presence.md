# 在线会话与 Presence（WebSocket）

> **最后更新**：2026-07-20  
> 关联：`backend/pkg/presence/` · `backend/routes/ws.go` · `frontend/src/composables/usePresence.ts`

## 作用

管理端可查看当前在线用户会话，并支持踢下线。前端登录后通过 WebSocket 心跳上报在线状态；同浏览器用 `browser_id` + `BroadcastChannel` 尽量合并为一条会话。

## 后端

| 路径 | 说明 |
|------|------|
| `GET /api/v1/ws/presence` | Presence WebSocket（需有效 JWT；CORS 与 WS 校验见中间件测试） |
| `GET /api/v1/{ADMIN}/online/stats` | 在线统计 |
| `GET /api/v1/{ADMIN}/online/sessions` | 会话列表 |
| `DELETE /api/v1/{ADMIN}/online/sessions/:id` | 吊销会话 |

核心包：`pkg/presence`（Hub、连接处理、简单限流）。会话落库/吊销与 `user_sessions` 联动。

## 前端

| 文件 | 说明 |
|------|------|
| `composables/usePresence.ts` | 布局挂载后连 WS、心跳 |
| `utils/browserId.ts` | 浏览器级 ID，减少多 Tab 重复在线 |
| `views/admin/online-users/` | 管理端「在线用户」页（用户管理下） |
| `service/api/admin/online.ts` | 管理 API |

心跳间隔可由系统设置注入（见 app-config / settings）。

## 相关：过期 Token 强退

- 公开：`POST /api/v1/public/session/force-logout`  
  允许携带**已过期** access token（`ParseTokenForGuardIgnoreExpiry`），按 token hash 吊销会话。  
- 前端登出时会尽量调用，避免过期会话残留。

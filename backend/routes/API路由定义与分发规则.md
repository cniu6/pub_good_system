# API路由定义与分发规则 (Routes)

## 简介
定义全站 API 端点及其对应的处理函数与权限。

## 功能字段与函数
- `SetupRoutes`: 路由注册总入口。
- 包含 `/api/register`, `/api/login` 等公共接口。
- 包含 `/api/profile`, `/api/admin/*` 等受保护接口。
- 包含 `/api/v1/system/cleanup-status` 系统状态接口（需登录），用于查询验证码清理任务状态。

## 规范
- 路由分组必须清晰。
- 管理端路径支持通过配置动态修改。

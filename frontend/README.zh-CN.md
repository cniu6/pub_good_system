<div align="center">
    <h1>F.st Admin</h1>
    <p>基于 Vue3 · Vite · TypeScript · Naive UI 的后台管理前端</p>
</div>

<div align='center'>

  [English](./README.md) | 中文
</div>

## 介绍

F.st 前端是 [F.st (Full Stack Template)](../README.md) 项目的管理后台与用户中心界面，基于 Vue3、Vite、TypeScript、Naive UI 构建，配合 Go(Gin) 后端一起使用，前后端分离，也支持打包为单二进制同源部署。

## 特性

- 基于 Vue3、Vite、TypeScript、NaiveUI、UnoCSS 等主流技术栈开发
- 基于 [alova](https://alova.js.org/) 封装和配置，提供统一的响应处理和多场景能力
- 完善的前后端权限管理方案（用户端 / 管理端分离鉴权）
- 支持本地静态路由和后台返回动态路由，路由简单易配置
- 对日常使用频率较高的组件二次封装，满足基础工作需求
- 黑暗主题适配，界面样式保持 Naive 风格
- 多语言（i18n）支持（中文 / 英文）

## 安装使用

本地开发环境建议使用 pnpm 10.x、Node.js 21.x

```bash
# 安装依赖
pnpm i

# 开发
pnpm dev

# 构建产物
pnpm build
```

环境变量说明见 `.env.example`；仅本地开发用的敏感配置放在 `.env`（不提交版本库）。

## 目录结构

详见项目根目录 [README.md](../README.md) 与 [frontend/留档.md](./留档.md)。

## 协议

[MIT](LICENSE)

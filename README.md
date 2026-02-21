# FST - 全栈后台管理系统模板

FST (Full Stack Template) 是一个基于 Go (Gin) 和 Vue 3 (Naive UI) 构建的高性能、轻量级全栈后台管理系统模板。它集成了现代 Web 开发的最佳实践，旨在提供一个开箱即用的开发基础。

## 🚀 特性

- **前后端分离架构**：后端采用 Go Gin，前端采用 Vue 3 + Vite。
- **高性能后端**：使用 Go 语言，具备极高的并发处理能力和内存效率。
- **现代化前端**：基于 Vue 3、TypeScript、Naive UI、UnoCSS 和 Alova。
- **插件化设计**：后端支持插件扩展，方便集成第三方功能。
- **完善的权限系统**：内置 JWT 认证和路由级权限控制。
- **Swagger API 文档**：内置集成 Swagger，支持通过环境变量开启/关闭，自动生成接口文档。
- **多语言支持**：内置 i18n 国际化方案。
- **一键构建**：提供跨平台构建脚本（Windows/Linux）。

## 🛠️ 技术栈

### 后端 (Backend)
- **框架**: [Gin](https://gin-gonic.dev/)
- **数据库**: [MySQL](https://www.mysql.com/) (使用 [sqlx](https://github.com/jmoiron/sqlx))
- **认证**: [JWT](https://github.com/golang-jwt/jwt)
- **文档**: [Swagger](https://github.com/swaggo/swag)
- **配置**: [godotenv](https://github.com/joho/godotenv)

### 前端 (Frontend)
- **框架**: [Vue 3](https://vuejs.org/)
- **构建工具**: [Vite](https://vitejs.dev/)
- **UI 组件库**: [Naive UI](https://www.naiveui.com/)
- **样式**: [UnoCSS](https://unocss.dev/)
- **状态管理**: [Pinia](https://pinia.vuejs.org/)
- **网络请求**: [Alova](https://alova.js.org/)
- **语言**: [TypeScript](https://www.typescriptlang.org/)

## 📂 项目结构

```text
.
├── backend/            # 后端 Go 源码
│   ├── app/            # 业务逻辑核心
│   ├── cmd/            # 程序启动入口与静态资源托管
│   ├── internal/       # 内部系统库（配置、数据库、中间件）
│   ├── routes/         # API 路由定义
│   └── utils/          # 通用工具函数
├── frontend/           # 前端 Vue 源码
│   ├── src/            # 前端核心代码
│   ├── public/         # 静态资源
│   └── ...             # 前端配置文件
├── build/              # 跨平台构建产物
├── .env                # 环境变量配置
├── dev.bat/ps1         # 本地开发启动脚本
└── build.bat/ps1       # 项目构建脚本
```

## ⚙️ 快速开始

### 1. 环境准备
- [Go](https://golang.org/) 1.24+
- [Node.js](https://nodejs.org/) 18+ & [pnpm](https://pnpm.io/)
- [MySQL](https://www.mysql.com/) 8.0+

### 2. 数据库配置
1. 创建数据库 `fst_platform`。
2. 修改根目录下的 `.env` 文件，配置你的数据库连接信息。
3. 运行项目时，后端会自动执行数据库迁移并创建必要的表。

### 3. 运行项目
#### 自动脚本 (推荐)
```bash
# Windows
./dev.bat
```

#### 手动运行
**后端:**
```bash
go run backend/cmd/main.go
```

**前端:**
```bash
cd frontend
pnpm install
pnpm dev
```

## 📝 API 文档 (Swagger)

项目已集成 Swagger 用于自动生成和展示 API 文档。

### 开启/关闭
在根目录的 `.env` 文件中，可以通过 `ENABLE_SWAGGER` 变量控制：
```env
ENABLE_SWAGGER=true  # 开启 (默认)
ENABLE_SWAGGER=false # 关闭
```

### 查看文档
1. 启动后端服务。
2. 在浏览器访问：`http://localhost:8080/swagger/index.html`。

### 更新文档
如果你在后端代码中添加或修改了 Swagger 注释（如在 `auth_controller.go` 中），请运行以下命令重新生成文档：
```bash
# 在项目根目录下执行
go run github.com/swaggo/swag/cmd/swag init -g backend/cmd/main.go -o backend/docs
```

## 🏗️ 构建

使用提供的构建脚本可以一键生成跨平台的二进制文件：

```bash
# Windows
./build.bat
```

构建产物将存放在 `build/` 目录下，按平台（Windows/Linux）和架构（amd64/arm64）分类。

## 📄 开源协议

本项目基于 [MIT](LICENSE) 协议开源。

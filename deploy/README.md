# FST Docker 部署说明

## 前置

- 已安装 Docker / Docker Compose
- 在**仓库根目录**操作（构建上下文需要 `go.mod`、`frontend/`、根 `.env.example`）

## 首次启动（MySQL，推荐）

```bash
# 1) 准备环境变量（勿把真实密码提交进 Git）
copy deploy\.env.example deploy\.env   # Windows
# cp deploy/.env.example deploy/.env   # Linux/macOS

# 2) 编辑 deploy/.env：至少改 JWT_SECRET、DB_PASSWORD、MYSQL_ROOT_PASSWORD

# 3) 构建并启动（embedded 单二进制 + MySQL）
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build
```

浏览器访问：`http://localhost:8080`  
健康检查：`http://localhost:8080/health`  
就绪检查：`http://localhost:8080/ready`  
简易指标：`http://localhost:8080/metrics`

## 可选：PostgreSQL

> 注意：PostgreSQL 适配层尚未用真机 CI 认证为生产可用；上生产前请先跑通  
> `go test -tags integration ./backend/pkg/db/ -run Postgres`（需设置 `FST_PG_DSN`）。

```bash
# 修改 deploy/.env：
#   DB_DRIVER=postgres
#   DB_HOST=postgres
#   DB_PORT=5432
#   DB_SSLMODE=disable

docker compose -f deploy/docker-compose.yml --env-file deploy/.env --profile postgres up -d --build
```

启用 `postgres` profile 后会启动 Postgres 容器；请同时把 `app` 的 `DB_*` 指到 `postgres`。若只要 Postgres、不要 MySQL，可在 compose 里去掉 mysql 依赖或给 mysql 也加 profile。

## 常用命令

```bash
# 查看日志
docker compose -f deploy/docker-compose.yml logs -f app

# 停止
docker compose -f deploy/docker-compose.yml --env-file deploy/.env down

# 仅重建应用镜像
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build app
```

## 镜像说明

`deploy/Dockerfile` 为三阶段构建：

1. Node 20 + pnpm 构建前端到根目录 `dist/`
2. Go 1.24 `go build -tags embedded` 打入单文件
3. Alpine 运行时，健康检查打 `/health`

容器内默认用户 `fst`（uid 10001），监听 `8080`。配置通过 `env_file` / `environment` 注入，**不要把含真实口令的 `.env` 打进镜像层**。

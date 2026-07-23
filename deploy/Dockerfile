# FST 多阶段构建：前端 dist → Go embedded 单二进制
# 构建上下文必须是仓库根目录：
#   docker build -f deploy/Dockerfile -t fst-app .

# ---------- 阶段 1：前端（产物落到 /src/dist，与 go:embed 一致）----------
FROM node:20-alpine AS frontend
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml ./
COPY frontend/ ./
ENV VITE_BUILD_MODE=embedded
RUN pnpm install --frozen-lockfile && pnpm build

# ---------- 阶段 2：Go embedded ----------
FROM golang:1.24-alpine AS backend
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Vite outDir = 仓库根 dist → 容器内 /src/dist
COPY --from=frontend /src/dist ./dist
RUN test -f dist/index.html
ENV CGO_ENABLED=0
RUN go build -tags embedded -ldflags "-X main.BuildMode=embedded -s -w" -o /out/fst .

# ---------- 阶段 3：运行时 ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
 && adduser -D -H -u 10001 fst
WORKDIR /app
COPY --from=backend /out/fst /app/fst
COPY .env.example /app/.env.example
USER fst
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["/app/fst"]

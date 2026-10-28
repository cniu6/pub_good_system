# Go 依赖漏洞定点修复报告

> 扫描日期：2026-10-29
> 范围：`pub_good_system`（module `fst`）
> 方法：`govulncheck` 可达性分析 + OSV / NVD / GitHub Advisory 版本核对 + 自有高危面代码审计
> 未执行任何公开 PoC，也未连接生产/共享数据库。

## 一、结论

- 修复前 `govulncheck ./...`：**16 条可达漏洞**（3 个模块 + Go 标准库）。
- 修复后 `govulncheck ./...`：**0 条可达漏洞**（`Your code is affected by 0 vulnerabilities`）。
- 依赖升级：`golang.org/x/crypto`、`golang.org/x/net`、`golang.org/x/text`、`quic-go`、`pgx/v5`。
- 工具链：`go 1.26.0` + `toolchain go1.27.3`；Docker 构建镜像固定为 `golang:1.27.3-alpine`。
- 自有代码：JWT 密钥分离、CORS 凭证边界、管理端高危路由与三处 `exec`/`fmt.Sprintf` sink 已完成复核；新增 JWT/CORS/管理端鉴权回归测试。
- 未发现需要修改业务契约的漏洞；未改动路由路径、JWT Claims 字段名和响应结构。

## 二、依赖变更

| 模块 | 修复前 | 修复后 | 说明 |
|---|---|---|---|
| `golang.org/x/crypto` | v0.40.0 | v0.58.0 | 覆盖 CVE-2026-46595（x/crypto/ssh）修复线 v0.52.0；同时满足 x/net v0.62.0 的最低要求 |
| `golang.org/x/net` | v0.42.0 | v0.62.0 | 覆盖 GO-2026-6617 / 6612 / 6611 / 6610 / 6603（修复线 v0.60.0）及 GO-2026-6866 / 6860（修复线 v0.62.0）；亦高于 CVE-2026-42506 的 v0.55.0 |
| `golang.org/x/text` | v0.27.0 | v0.43.0 | 覆盖 GO-2026-5970（v0.39.0）与 GO-2026-6629（v0.41.0） |
| `github.com/quic-go/quic-go` | v0.54.0 | v0.59.1 | 覆盖 GO-2026-5676 / CVE-2026-40898（v0.59.1）与 GO-2025-4233（v0.57.0）；`qpack` 同步升至 v0.6.0 |
| `github.com/jackc/pgx/v5` | v5.7.6 | v5.9.2 | 覆盖 GO-2026-5004（SQL 注入，修复线 v5.9.2） |
| Go 工具链 | `go 1.25` | `go 1.26.0` + `toolchain go1.27.3` | x/net v0.62.0 的 `GoVersion` 为 1.26.0，无法保留 1.25；标准库修复线为 go1.27.3（crypto/x509）与 go1.27.2（net/http 等） |

保持不变：

- `github.com/golang-jwt/jwt/v5 v5.3.1`：高于 CVE-2025-30204 的修复线 v5.2.2，无新增公告。
- `github.com/gin-gonic/gin v1.11.0`：`govulncheck` 未报告该项目可达的 Gin 公告；未把 `go.senan.xyz/gonic` 的 CVE 混入。

> 口径修正：任务稿称数据层为 Sqlx；实际 `go.mod` 使用 `gorm.io/gorm v1.30.0`，无 `jbs`/`sqlx` 依赖。本报告按实际依赖扫描。

## 三、govulncheck 修复前后

修复前可达条目（16）：

| 编号 | 模块/包 | 修复线 | 处置 |
|---|---|---|---|
| GO-2026-6809 | `crypto/x509`（标准库） | go1.27.3 | 工具链升级 |
| GO-2026-6629 | `x/text/secure/precis` | v0.41.0 | 升级 x/text |
| GO-2026-6617 | `x/net/http2` HPACK 竞态 | v0.60.0 | 升级 x/net |
| GO-2026-6613 | `net/http` CONNECT 反同步 | go1.27.2 | 工具链升级 |
| GO-2026-6612 | `x/net/http2` 流控重复退款 | v0.60.0 | 升级 x/net |
| GO-2026-6611 | `x/net/http2` 初始窗口 CPU 消耗 | v0.60.0 | 升级 x/net |
| GO-2026-6610 | `x/net/http2` 畸形 framing 头 | v0.60.0 | 升级 x/net |
| GO-2026-6609 | `net/http` Range 头无上限 | go1.27.2 | 工具链升级 |
| GO-2026-6608 | `net/textproto` MIME 头内存限制绕过 | go1.27.2 | 工具链升级 |
| GO-2026-6607 | `crypto/tls` ECH 扩展 | go1.27.2 | 工具链升级 |
| GO-2026-6605 | `net/http` CONNECT 拒绝后反同步 | go1.27.2 | 工具链升级 |
| GO-2026-6603 | `x/net/http2` Trailer 内存耗尽 | v0.60.0 | 升级 x/net |
| GO-2026-5970 | `x/text` 非法输入死循环 | v0.39.0 | 升级 x/text |
| GO-2026-5676 | `quic-go` QPACK Trailer 内存耗尽 | v0.59.1 | 升级 quic-go |
| GO-2026-5004 | `pgx/v5` 美元引号占位符 SQL 注入 | v5.9.2 | 升级 pgx |
| GO-2025-4233 | `quic-go` QPACK Header DoS | v0.57.0 | 升级 quic-go |

修复后结果：

```text
=== Symbol Results ===
No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.
```

剩余 1 条模块级公告未形成可达调用链；`govulncheck` 退出码为 0。明细：

| 编号 | 模块/包 | 当前版本 | 修复版本 | 可达性 | 说明 |
|---|---|---|---|---|---|
| [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) | `golang.org/x/crypto/openpgp` | v0.58.0 | N/A（无修复） | not reachable | 项目仅导入 `x/crypto/bcrypt`，`go list -deps ./...` 无 `openpgp`；该子包已不再维护 |

## 四、点名 CVE 核验

| CVE | 组件 | 受影响 | 修复版本 | 当前版本 | 结论 |
|---|---|---|---|---|---|
| [CVE-2026-46595](https://nvd.nist.gov/vuln/detail/CVE-2026-46595) / [GHSA-x527-x647-q7gg](https://github.com/advisories/GHSA-x527-x647-q7gg) | `golang.org/x/crypto/ssh` | `< 0.52.0` | 0.52.0 | 0.58.0 | 本项目仅使用 `x/crypto/bcrypt`，`go list -deps` 无 `x/crypto/ssh`，不可达；版本仍已高于修复线 |
| [CVE-2026-42506](https://nvd.nist.gov/vuln/detail/CVE-2026-42506) | `golang.org/x/net/html` | `< 0.55.0` | 0.55.0 | 0.62.0 | `x/net/html` 经 `mimetype→charset` 进入构建，但项目不使用 `html.Render` 渲染不可信 HTML；版本已高于修复线 |
| [CVE-2026-40898](https://nvd.nist.gov/vuln/detail/CVE-2026-40898) / [GHSA-vvgj-x9jq-8cj9](https://github.com/advisories/GHSA-vvgj-x9jq-8cj9) | `quic-go` | `< 0.59.1` | 0.59.1 | 0.59.1 | `gin` 的 `http3` 包引入 quic-go；项目未调用 `RunQUIC`，但仍已升级到修复版本 |
| [CVE-2025-30204](https://nvd.nist.gov/vuln/detail/CVE-2025-30204) | `golang-jwt/jwt/v5` | `< 5.2.2` | 5.2.2 | 5.3.1 | 不受影响，保持不动 |

## 五、自有代码审计

### 5.1 JWT 验证入口

全仓 `ParseWithClaims` 仅出现在：

- `backend/utils/jwt.go:115` `ParseTokenForGuard`
- `backend/utils/jwt.go:152` `ParseTokenForGuardIgnoreExpiry`
- `backend/utils/jwt.go:218` `ParseRefreshTokenForGuard`

三处均使用 `jwtSigningKeyByGuard(expectedGuard)`；keyfunc 只接受 `HS256`/HMAC。access/refresh 分别校验 `token_type`；所有路径校验 `auth_guard == expectedGuard`；admin guard 额外要求 `role == admin`。

本次加固：

- `backend/pkg/config/config.go`：生产环境 `JWT_ADMIN_SECRET` 为空时拒绝启动；显式设置 `ALLOW_ADMIN_JWT_SECRET_FALLBACK=true` 才允许回退；**即使允许回退，只要 `JWT_ADMIN_SECRET` 与 `JWT_SECRET` 相同也拒绝启动**。
- `backend/utils/jwt.go`：`getJWTSecretByGuard` 增加第二道防线——生产环境下 admin secret 为空且未显式允许回退时直接返回错误，覆盖不经过启动校验的入口（测试、CLI）。
- 新增 `TestAdminJWTSecretSeparatedInProduction`、`TestAdminJWTSecretFallbackRequiresExplicitOptIn`、`TestAdminRoutes_UserTokenRejected`、`TestAdminRoutes_AdminTokenAccepted`。

### 5.2 CORS

`backend/pkg/middleware/cors.go`：精确 Origin 或显式 `*.example.com` 才回显 `Access-Control-Allow-Origin` 并带 `Vary: Origin`；`CORS_ORIGINS=*` 只返回 `*`，不设置 `Access-Control-Allow-Credentials`；空配置拒绝。

新增 `backend/pkg/middleware/cors_credentials_test.go` 覆盖：

- 白名单 Origin 回显 + `Allow-Credentials: true`
- `https://evil.example` 不带 Cookie 也拿不到任何 CORS 许可
- `*` 不与凭证同现
- `*.example.com` 不匹配 `notexample.com`

### 5.3 `exec.Command` 与 SQL 拼接 sink

| 位置 | 数据来源 | 判定 | 处置 |
|---|---|---|---|
| `backend/app/plugins/swagger_auto.go:158` | `exec.Command("go", "run", <固定相对路径>)` | 文件带 `//go:build !prod`，仅 `APP_MODE=dev` 且 `GO_ENV!=production` 时后台运行；参数不含请求输入 | 保留，报告留痕 |
| `backend/app/plugins/swagger_auto.go:185` | `exec.Command(swagPath, 固定 args...)`；`swagPath` 来自 `exec.LookPath` | 同上，仅开发期 | 保留，报告留痕 |
| `backend/internal/migrate/migrate_pay_gateway_ext_config.go:62` | 表名/列名均为函数内硬编码 + 按数据库方言引用 | 无外部输入，不可注入 | 保留，报告留痕 |
| `backend/app/services/user_service.go:169` | `models.BuildUserSelectColumns("users")` 固定列清单 + `models.RealPaidOrderFilterSQL` 常量；`id` 使用 `?` 绑定 | 无外部拼接 | 保留，报告留痕 |

### 5.4 管理端高危面

| 面 | 路由 | 鉴权/开关 | 结论 |
|---|---|---|---|
| 全部 admin REST | `/api/v1{ADMIN_API_PATH}/*` | `AuthMiddlewareForGuard("admin")` → `AdminOnly()` → `DynamicAdminRateLimitMiddleware()` | 统一鉴权，子路由继承 |
| 调试终端 / pprof / GC / 重启 | `/debug/*` | `IsAdminDebugOpsEnabled()` 注册时短路；生产恒 false | 生产不可达 |
| 数据库控制台 | `/db/*` | 继承 admin 双中间件；写操作 `requireDBConsoleWrite` + `IsAdminDBWriteEnabled()`，生产恒 false | 生产只读；只读 SQL 仅 `SELECT/SHOW/EXPLAIN` |
| 插件自动注册 | `backend/app/plugins` | 插件在 `init()` 编译期注册，`gen_plugins.go` 是 `//go:build ignore` 的代码生成器，不扫描运行时目录 | 无运行期任意加载面 |

## 六、验证记录

在 `pub_good_system` 下执行：

| 命令 | 结果 |
|---|---|
| `go mod tidy` ×2 | 第二次无 diff（`TIDY_STABLE=yes`） |
| `go build ./...` | 通过 |
| `go build -tags embedded ./...` | 通过 |
| `go vet ./backend/...` | 通过 |
| `go test ./backend/... -count=1` | 通过 |
| `go test ./... -count=1` | 仅 `fst/tools/payment_test [setup failed]`：`tools/payment_test/payment_test.go:9` 导入 `fst/backend/internal/migrate` 违反 internal 规则，属改动前既存问题，与本次变更无关 |
| `go vet ./...` | 同上，仅该 tools 包报错 |
| `govulncheck ./...` | 0 条可达漏洞（修复前 16 条） |
| `git diff --check` | 通过（仅 CRLF 提示） |
| `docker build -f deploy/Dockerfile -t fst:security-check .` | 未执行：本机未安装 Docker（`docker` 不是可识别命令）；Dockerfile 已固定 `golang:1.27.3-alpine`，该 tag 经 Docker Hub 检索存在 |

## 七、剩余风险

1. **Docker 构建未实测**：当前主机无 Docker，无法完成镜像构建验证；需在具备 Docker 的环境补跑。
2. **`tools/payment_test` 编译失败为既存问题**：不属于本次依赖/安全改动范围，未修改。
3. **生产部署需配置 `JWT_ADMIN_SECRET`**：升级后若生产 `.env` 仍为空，会按设计拒绝启动；可临时设 `ALLOW_ADMIN_JWT_SECRET_FALLBACK=true`，但不建议长期使用。
4. `golang.org/x/net` v0.62.0 要求 `go 1.26.0`，因此 `go.mod` 的 `go` 指令从 1.25 升到 1.26.0；这是采用修复版本的必要条件，不是无关的大版本升级。
5. **GO-2026-5932（x/crypto/openpgp，无修复版本）**：当前版本仍被 `govulncheck` 标为模块级公告，但项目未导入 `openpgp` 子包，无可达符号；如后续引入 openpgp，需单独评估或改为维护中的替代实现。

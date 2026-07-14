# Embedded 打包问题总结与修复留档

> 文件位置：`golang/留档-embedded打包问题总结.md`
> 
> 最后更新：2026-04-26
> 
> 适用范围：`build.bat`、`backend/cmd/main_embedded.go`、前端 embedded 同源 API、静态资源嵌入、HTTP 超时

---

## 1. 问题背景

本项目支持两种生产构建模式：

- `Embedded - Single binary`
- `External - Separate frontend assets`

按设计，`Embedded` 模式应该做到：

- 前端先构建到 `golang/dist`
- 再复制到 `backend/cmd/dist`
- Go 通过 `go:embed` 把前端打进二进制
- 启动后同一个程序同时提供：
  - 前端页面
  - 后端 API
- 前端在 embedded 模式下直接请求同源 `/api/...`

但原始源码存在多处缺陷，导致 embedded 模式实际上不稳定。

---

## 2. 问题总览

### 2.1 主要问题

- 前端资源复制路径写错，导致资源没真正嵌进去
- embedded 托管逻辑错误依赖 `APP_MODE`
- 根路径 `/` 因静态托管实现触发 301 循环
- embedded 模式下 API 地址仍固化为 `http://localhost:8046`
- 大 JS 资源下载约 10 秒时被写超时切断
- `build.bat` 的 `pause` 导致 IDE 中明明成功却显示失败
- 默认构建 arm 产物，不符合当前需求

### 2.2 影响

- 页面 404 / 假成功
- 首页无法打开
- 一体化部署后仍连错 API 地址
- 静态资源大文件加载失败
- IDE 中构建状态误判

---

## 3. 问题一：前端资源复制路径错误

### 3.1 现象

打包时出现：

```text
Copying frontend assets...
Invalid path
```

二进制能编出来，但 embedded 页面不能正常访问。

### 3.2 根因

前端 Vite 的真实输出目录是：

```text
golang/dist
```

但原脚本复制的是：

```bat
frontend\dist
```

因此脚本复制源路径错误。

### 3.3 怎么修

统一把复制来源改为项目根目录 `dist`。

### 3.4 关键代码

**文件：`build.bat`**

```bat
if exist dist rmdir /s /q dist
```

```bat
xcopy /s /e /q /y /i dist\* backend\cmd\dist\ >nul
```

```bat
if "%BMODE%" == "external" (
    echo     - Copying external assets...
    mkdir %OUTDIR%\dist >nul 2>&1
    xcopy /s /e /q /y dist\* %OUTDIR%\dist\ >nul
)
```

### 3.5 如何验证

- 执行 `build.bat 1`
- 启动 embedded 产物
- 确认访问 `/` 能返回前端页面
- 确认不再出现 `Invalid path`

---

## 4. 问题二：embedded 托管逻辑错误依赖 `APP_MODE`

### 4.1 现象

已经打成 embedded，但前端仍不托管，访问页面会失败。

### 4.2 根因

原逻辑要求：

- `AppMode == "integrated"`
- 且 `BuildMode != "none"`

这会把“是否托管前端”错误绑定到运行时环境变量，而不是构建模式。

### 4.3 怎么修

只要构建模式不是 `none`，就直接托管前端，不再依赖 `APP_MODE`。

### 4.4 关键代码

**文件：`backend/cmd/main_embedded.go`**

```go
// 前端资源处理
// 只要构建模式不是 none，就提供前端托管能力
if BuildMode != "none" {
```

```go
} else {
	log.Printf("[Mode] Backend only: BuildMode=%s, frontend not served", BuildMode)
}
```

### 4.5 如何验证

- `.env` 中不设置 `APP_MODE=integrated`
- 仍然用 embedded 包启动
- 前端照常可访问

---

## 5. 问题三：根路径 `/` 301 循环

### 5.1 现象

日志反复出现：

```text
GET "/" -> 301
```

### 5.2 根因

原实现对 embedded 资源使用 `http.FileServer`，在嵌入式文件系统和根路径场景下，容易触发目录重定向，从而产生 301 循环。

### 5.3 怎么修

不再使用 `http.FileServer` 作为首页兜底。

改为：

- 静态文件用 `fs.ReadFile` 直接返回
- 根路径和 SPA 路径统一返回 `index.html`

### 5.4 关键代码

**文件：`backend/cmd/main_embedded.go`**

```go
if rawFS != nil {
	_ = publicFS
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			utils.Fail(c, 404, "API not found")
			return
		}
		relPath := strings.TrimPrefix(path, "/")
		if relPath != "" && relPath != "index.html" {
			if data, err := fs.ReadFile(rawFS, relPath); err == nil {
				contentType := mime.TypeByExtension(filepath.Ext(relPath))
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				c.Data(http.StatusOK, contentType, data)
				return
			}
		}
		indexData, err := fs.ReadFile(rawFS, "index.html")
		if err != nil {
			log.Printf("[Embedded] index.html 读取失败: %v", err)
			c.Status(http.StatusNotFound)
			return
		}
		log.Printf("[Embedded] 返回 index.html for path: %s", path)
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})
}
```

### 5.5 如何验证

访问：

- `/`
- `/admin`

预期：

- 返回 `200`
- 不再出现连续 `301`

---

## 6. 问题四：embedded 模式下 API 地址仍写死为 localhost

### 6.1 现象

即使前后端已经打包到一起，前端仍请求：

```text
http://localhost:8046/api/v1/public/app-config
```

而不是：

```text
/api/v1/public/app-config
```

### 6.2 根因

前端生产配置始终读取 `VITE_API_URL`。

**文件：`frontend/.env.production`**

```env
VITE_API_URL=http://localhost:8046
```

所以 embedded 构建后，API 地址被固化成绝对路径。

### 6.3 怎么修

增加一个构建期变量 `VITE_BUILD_MODE`：

- `embedded` 模式：生产 API 地址置空 `''`
- `external` 模式：继续使用 `VITE_API_URL`

同时在 `build.bat` 中构建前端时注入当前模式。

### 6.4 关键代码

**文件：`frontend/service.config.ts`**

```ts
export function getServiceConfig(env: Record<string, string>): Record<ServiceEnvType, Record<string, string>> {
  const apiUrl = env.VITE_API_URL || 'http://localhost:8085'
  const buildMode = env.VITE_BUILD_MODE || ''
  const productionApiUrl = buildMode === 'embedded' ? '' : apiUrl

  return {
    dev: {
      url: apiUrl,
    },
    production: {
      url: productionApiUrl,
    },
  }
}
```

**文件：`build.bat`**

```bat
cd frontend
set VITE_BUILD_MODE=%BMODE%
call pnpm build
set VITE_BUILD_MODE=
```

**文件：`frontend/.env.production`**

```env
# 生产环境后端 API 地址
# 仅 external 模式或独立前端构建时使用；embedded 模式会自动走同源相对路径 /api
VITE_API_URL=http://localhost:8046
```

### 6.5 如何验证

- 执行 `build.bat 1`
- 启动 embedded 包
- 打开浏览器网络面板
- 查看 `app-config` 请求地址

预期应为：

```text
/api/v1/public/app-config
```

---

## 7. 问题五：大资源在 10 秒左右超时

### 7.1 现象

大 JS 请求在约 10 秒时失败，日志出现：

```text
i/o timeout
broken pipe
```

### 7.2 根因

HTTP 写超时默认值过小：

**文件：`backend/utils/http_server.go`**

```go
writeTimeout := 10 * time.Second
```

同时配置默认值也只有 `10` 秒。

### 7.3 怎么修

把默认写超时提高到 `60` 秒，并同步代码与 env 示例。

### 7.4 关键代码

**文件：`backend/utils/http_server.go`**

```go
func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	readHeaderTimeout := 5 * time.Second
	readTimeout := 5 * time.Second
	writeTimeout := 60 * time.Second
	idleTimeout := 120 * time.Second
	maxHeaderBytes := 1 << 20
```

**文件：`backend/pkg/config/config.go`**

```go
HTTPWriteTimeoutSeconds:      getEnvAsPositiveInt("HTTP_WRITE_TIMEOUT_SECONDS", 60),
```

**文件：`.env.example`**

```env
# 响应写入超时（单位：秒）
HTTP_WRITE_TIMEOUT_SECONDS=60
```

### 7.5 如何验证

- 重新部署二进制
- 访问 embedded 页面
- 检查大文件是否仍在 10 秒附近失败
- 检查正式 `.env` 中是否也同步配置：

```env
HTTP_WRITE_TIMEOUT_SECONDS=60
```

---

## 8. 问题六：`build.bat` 成功后在 IDE 里被误判为失败

### 8.1 现象

日志明明已经：

```text
[4/4] Build complete
```

但 IDE 仍显示 `Exit code: 1`。

### 8.2 根因

脚本里有多个 `pause`，IDE 中如果终止等待按键的批处理，会把结果记成失败。

### 8.3 怎么修

删除成功路径和失败路径中的 `pause`，让脚本直接按实际结果退出。

### 8.4 关键代码

**文件：`build.bat`**

成功路径：

```bat
echo [4/4] Build complete!
echo ============================================================
echo   Output: ./build/
echo   Mode:   [%BMODE%]
echo ============================================================
echo.
exit /b 0
```

失败路径：

```bat
:build_fail
echo.
echo [ERROR] Build aborted!
exit /b 1
```

### 8.5 如何验证

在 IDE 中执行：

```bat
.\build.bat 1
```

预期：

- 成功时返回 `Exit code: 0`
- 不再停在按键等待

---

## 9. 问题七：默认 arm 构建不符合当前需求

### 9.1 现象

原脚本默认构建：

- Windows amd64
- Windows arm64
- Linux amd64
- Linux arm64

当前需求只需要 amd64。

### 9.2 怎么修

暂时移除 arm 构建目标，仅保留 amd64。

### 9.3 关键代码

**文件：`build.bat`**

```bat
call :build_target windows amd64 .exe "Windows x64"
if %ERRORLEVEL% neq 0 goto :build_fail
call :build_target linux amd64 "" "Linux x64"
if %ERRORLEVEL% neq 0 goto :build_fail
```

### 9.4 如何验证

构建完成后只应生成：

```text
build/windows_amd64/fst.exe
build/linux_amd64/fst
```

---

## 10. 本次修复涉及文件

- `build.bat`
- `backend/cmd/main_embedded.go`
- `backend/utils/http_server.go`
- `backend/pkg/config/config.go`
- `.env.example`
- `frontend/service.config.ts`
- `frontend/.env.production`

---

## 11. 推荐验证流程

### 11.1 构建

```bat
.\build.bat 1
```

确认：

- 构建成功返回 `0`
- 输出仅包含 amd64 产物

### 11.2 运行

```bat
build\windows_amd64\fst.exe
```

确认日志包含：

```text
[Mode] Embedded: Serving from binary internal FS
```

### 11.3 页面验证

访问：

- `/`
- `/admin`

确认：

- 返回 `200`
- 不再出现 301 循环

### 11.4 API 验证

在浏览器网络面板中确认：

```text
/api/v1/public/app-config
```

而不是：

```text
http://localhost:8046/api/v1/public/app-config
```

### 11.5 大资源验证

确认下列资源不再在 10 秒左右失败：

- `vendor-naive-*.js`
- `vendor-echarts-*.js`
- `main-*.js`

---

## 12. 结论

本次 embedded 模式问题不是单点故障，而是打包脚本、前端服务地址、后端托管逻辑、静态资源处理和 HTTP 超时多个点叠加造成的系统性问题。

修复完成后，embedded 模式已经具备以下能力：

- 正确复制并嵌入前端资源
- 正常托管首页和管理页
- 生产环境自动走同源 API
- 大资源下载不再容易在 10 秒超时
- 构建脚本在 IDE 中返回正确退出码
- 构建目标已收敛到当前需要的 amd64

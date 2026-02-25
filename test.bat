@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   FST 项目集成测试脚本
echo ========================================
echo.

:: 设置颜色
set "GREEN=[92m"
set "RED=[91m"
set "YELLOW=[93m"
set "RESET=[0m"

:: 错误计数
set ERROR_COUNT=0

:: 1. 检查 Go 环境
echo [1/6] 检查 Go 环境...
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo %RED%✗ Go 未安装%RESET%
    set /a ERROR_COUNT+=1
) else (
    for /f "tokens=3" %%v in ('go version 2^>^&1') do set GO_VERSION=%%v
    echo %GREEN%✓ Go 已安装: !GO_VERSION!%RESET%
)

:: 2. 检查 Node.js 环境
echo.
echo [2/6] 检查 Node.js 环境...
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo %RED%✗ Node.js 未安装%RESET%
    set /a ERROR_COUNT+=1
) else (
    for /f "tokens=1" %%v in ('node -v 2^>^&1') do set NODE_VERSION=%%v
    echo %GREEN%✓ Node.js 已安装: !NODE_VERSION!%RESET%
)

:: 3. 检查依赖
echo.
echo [3/6] 检查项目依赖...
cd /d "%~dp0"

:: Go 依赖
if not exist go.mod (
    echo %RED%✗ go.mod 不存在%RESET%
    set /a ERROR_COUNT+=1
) else (
    echo 检查 Go 模块...
    go mod verify >nul 2>&1
    if %errorlevel% equ 0 (
        echo %GREEN%✓ Go 依赖完整%RESET%
    ) else (
        echo %YELLOW%! Go 依赖可能需要更新，正在执行 go mod tidy...%RESET%
        go mod tidy
    )
)

:: 前端依赖
if exist frontend\package.json (
    if not exist frontend\node_modules (
        echo %YELLOW%! 前端依赖未安装，正在安装...%RESET%
        cd frontend
        call pnpm install
        cd ..
    ) else (
        echo %GREEN%✓ 前端依赖已安装%RESET%
    )
)

:: 4. 编译后端
echo.
echo [4/6] 编译后端...
go build -o build\test_backend.exe .\backend\cmd\main.go 2>&1
if %errorlevel% neq 0 (
    echo %RED%✗ 后端编译失败%RESET%
    set /a ERROR_COUNT+=1
    goto :end
) else (
    echo %GREEN%✓ 后端编译成功%RESET%
)

:: 5. 编译前端
echo.
echo [5/6] 编译前端...
if exist frontend\package.json (
    cd frontend
    call pnpm build >nul 2>&1
    if %errorlevel% neq 0 (
        echo %RED%✗ 前端编译失败%RESET%
        set /a ERROR_COUNT+=1
    ) else (
        echo %GREEN%✓ 前端编译成功%RESET%
    )
    cd ..
) else (
    echo %YELLOW%! 前端目录不存在，跳过%RESET%
)

:: 6. 检查配置文件
echo.
echo [6/6] 检查配置文件...
if exist .env (
    echo %GREEN%✓ .env 配置文件存在%RESET%
) else (
    echo %YELLOW%! .env 不存在，请复制 .env.example 并配置%RESET%
)

if exist .env.example (
    echo %GREEN%✓ .env.example 存在%RESET%
) else (
    echo %RED%✗ .env.example 不存在%RESET%
    set /a ERROR_COUNT+=1
)

:end
echo.
echo ========================================
if %ERROR_COUNT% equ 0 (
    echo %GREEN%✓ 所有测试通过！%RESET%
    echo.
    echo 后端测试程序: build\test_backend.exe
    echo 运行开发环境: dev.bat
) else (
    echo %RED%✗ 发现 %ERROR_COUNT% 个错误%RESET%
)
echo ========================================

:: 清理测试文件
if exist build\test_backend.exe del build\test_backend.exe >nul 2>&1

exit /b %ERROR_COUNT%

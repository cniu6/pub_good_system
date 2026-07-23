@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion
cd /d "%~dp0"

echo ========================================
echo   FST 测试（go test / go vet / pnpm lint）
echo ========================================
echo.

set "GREEN=[92m"
set "RED=[91m"
set "YELLOW=[93m"
set "RESET=[0m"
set ERROR_COUNT=0

:: 1. Go 单元测试
echo [1/4] go test ./backend/... -count=1
go test ./backend/... -count=1
if %errorlevel% neq 0 (
    echo %RED%✗ go test 失败%RESET%
    set /a ERROR_COUNT+=1
) else (
    echo %GREEN%✓ go test 通过%RESET%
)

:: 2. Go 静态检查
echo.
echo [2/4] go vet ./backend/...
go vet ./backend/...
if %errorlevel% neq 0 (
    echo %RED%✗ go vet 失败%RESET%
    set /a ERROR_COUNT+=1
) else (
    echo %GREEN%✓ go vet 通过%RESET%
)

:: 3. 前端 lint（目录存在时）
echo.
echo [3/4] 前端 pnpm lint
if exist frontend\package.json (
    pushd frontend
    if not exist node_modules (
        echo %YELLOW%! 安装前端依赖...%RESET%
        call pnpm install
        if !errorlevel! neq 0 (
            echo %RED%✗ pnpm install 失败%RESET%
            set /a ERROR_COUNT+=1
            popd
            goto :optional_build
        )
    )
    call pnpm lint
    if !errorlevel! neq 0 (
        echo %RED%✗ pnpm lint 失败%RESET%
        set /a ERROR_COUNT+=1
    ) else (
        echo %GREEN%✓ pnpm lint 通过%RESET%
    )
    popd
) else (
    echo %YELLOW%! 无 frontend 目录，跳过 lint%RESET%
)

:optional_build
:: 4. 可选：编译冒烟（传 --build 才跑）
echo.
if /i "%~1"=="--build" (
    echo [4/4] 可选编译冒烟（--build）
    go build -o build\test_backend.exe .
    if %errorlevel% neq 0 (
        echo %RED%✗ 后端编译失败%RESET%
        set /a ERROR_COUNT+=1
    ) else (
        echo %GREEN%✓ 后端编译成功%RESET%
        if exist build\test_backend.exe del /q build\test_backend.exe >nul 2>&1
    )
    if exist frontend\package.json (
        pushd frontend
        call pnpm build
        if !errorlevel! neq 0 (
            echo %RED%✗ 前端编译失败%RESET%
            set /a ERROR_COUNT+=1
        ) else (
            echo %GREEN%✓ 前端编译成功%RESET%
        )
        popd
    )
) else (
    echo [4/4] 跳过编译冒烟（需要时请：test.bat --build）
)

echo.
echo ========================================
if %ERROR_COUNT% equ 0 (
    echo %GREEN%✓ 全部通过%RESET%
) else (
    echo %RED%✗ 发现 %ERROR_COUNT% 个失败%RESET%
)
echo ========================================
exit /b %ERROR_COUNT%

@echo off
chcp 65001 >nul
echo.
echo ╔════════════════════════════════════════════════════════════╗
echo ║           FST 开发模式 - 自动更新 Swagger                  ║
echo ╚════════════════════════════════════════════════════════════╝
echo.

cd /d "%~dp0"

REM 设置开发环境变量
set GO_ENV=development

echo [1/2] 初始化: 扫描插件 + 更新 Swagger...
echo.

REM 扫描插件
go run backend/app/plugins/gen_plugins.go

REM 更新 Swagger
cd backend
swag init -g cmd/main.go -o docs --parseDependency --parseInternal 2>nul
if %ERRORLEVEL% neq 0 (
    echo [警告] Swagger 生成失败，请检查 swag 是否安装
    echo 安装命令: go install github.com/swaggo/swag/cmd/swag@latest
)
cd ..

echo.
echo [2/2] 启动服务...
echo ════════════════════════════════════════════════════════════
echo.
echo  服务地址: http://localhost:8080
echo  Swagger:  http://localhost:8080/swagger/index.html
echo.
echo  按 Ctrl+C 停止服务
echo ════════════════════════════════════════════════════════════
echo.

REM 启动服务
go run ./backend/cmd/main.go

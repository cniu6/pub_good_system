@echo off
chcp 65001 >nul
echo.
echo ============================================
echo  FST Swagger 文档生成工具
echo ============================================
echo.

cd /d "%~dp0"

echo [1/3] 扫描并注册插件...
echo.
go run backend/app/plugins/gen_plugins.go
if %ERRORLEVEL% neq 0 (
    echo [警告] 插件扫描失败，继续执行...
)
echo.

echo [2/3] 生成 Swagger 文档...
echo.
cd backend
swag init -g cmd/main.go -o docs --parseDependency --parseInternal

if %ERRORLEVEL% neq 0 (
    echo.
    echo [错误] Swagger 文档生成失败！
    echo 请确保已安装 swag: go install github.com/swaggo/swag/cmd/swag@latest
    cd ..
    pause
    exit /b 1
)

echo.
echo [3/3] 编译验证...
go build -o ../build/fst.exe ./cmd/main.go

if %ERRORLEVEL% neq 0 (
    echo.
    echo [错误] 编译失败！
    cd ..
    pause
    exit /b 1
)

cd ..

echo.
echo ============================================
echo  完成！
echo ============================================
echo.
echo  文档位置: backend/docs/
echo  访问地址: http://localhost:8080/swagger/index.html
echo.
echo ============================================
echo  添加新插件方法:
echo ============================================
echo.
echo  1. 在 backend/app/plugins/ 下创建插件目录
echo  2. 创建 plugin.go 文件，实现 Plugin 接口
echo  3. 在 init() 中调用 plugins.RegisterPlugin(NewPlugin())
echo  4. 重新运行此脚本 (gen-swagger.bat)
echo.
echo  无需手动修改任何导入！
echo.
pause

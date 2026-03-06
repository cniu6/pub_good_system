@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo.
echo ============================================================
echo   FST - Dev Mode (Auto Swagger)
echo ============================================================
echo.

set GO_ENV=development

echo [1/3] Scanning plugins + Generating Swagger...
echo.

go run backend/app/plugins/gen_plugins.go

cd backend
swag init -g cmd/main.go -o docs --parseDependency --parseInternal --quiet 2>nul
if %ERRORLEVEL% neq 0 (
    echo [WARN] Swagger generation failed. Install swag:
    echo   go install github.com/swaggo/swag/cmd/swag@latest
)
cd ..

echo.
echo [2/3] Starting frontend dev server...
start "FST Frontend" cmd /c "cd /d "%~dp0frontend" && pnpm dev"

echo.
echo [3/3] Starting backend server...
echo ============================================================
echo.
echo   Backend:  http://localhost:8080
echo   Frontend: http://localhost:5173
echo   Swagger:  http://localhost:8080/swagger/index.html
echo.
echo   Press Ctrl+C to stop backend
echo   Close "FST Frontend" window to stop frontend
echo ============================================================
echo.

go run ./backend/cmd/main.go

@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion
cd /d "%~dp0"

echo ============================================================
echo   [FST] Production Build Tool
echo ============================================================
echo.
echo   [1] Embedded - Single binary
echo   [2] External - Separate frontend assets
echo.

if "%~1" neq "" (
    set CHOICE=%~1
) else (
    set /p CHOICE="Select [1 or 2] (default 1): "
)
if "%CHOICE%" == "" set CHOICE=1

set BMODE=embedded
if "%CHOICE%" == "2" set BMODE=external

echo.
echo [1/4] Cleaning old build artifacts...

if exist build rmdir /s /q build
if exist frontend\dist rmdir /s /q frontend\dist
if exist backend\cmd\dist rmdir /s /q backend\cmd\dist

mkdir build >nul 2>&1
mkdir backend\cmd\dist >nul 2>&1
echo building... > backend\cmd\dist\index.html

echo.
echo [2/4] Building frontend (pnpm build)...
echo.

cd frontend
call pnpm build
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Frontend build failed!
    cd ..
    pause
    exit /b 1
)
cd ..

echo Copying frontend assets...
xcopy /s /e /q /y frontend\dist\* backend\cmd\dist\ >nul

echo.
echo [3/4] Cross-compiling Go backend... (Mode: %BMODE%)
echo.

set CGO_ENABLED=0

call :build_target windows amd64 .exe "Windows x64"
if %ERRORLEVEL% neq 0 goto :build_fail
call :build_target windows arm64 .exe "Windows arm64"
if %ERRORLEVEL% neq 0 goto :build_fail
call :build_target linux amd64 "" "Linux x64"
if %ERRORLEVEL% neq 0 goto :build_fail
call :build_target linux arm64 "" "Linux arm64"
if %ERRORLEVEL% neq 0 goto :build_fail

set CGO_ENABLED=
set GOOS=
set GOARCH=

echo.
echo [4/4] Build complete!
echo ============================================================
echo   Output: ./build/
echo   Mode:   [%BMODE%]
echo ============================================================
echo.
pause
exit /b 0

:build_target
set GOOS=%~1
set GOARCH=%~2
set EXT=%~3
set LABEL=%~4
set OUTDIR=build\%GOOS%_%GOARCH%

echo   - Building: %LABEL%...
mkdir %OUTDIR% >nul 2>&1

cd backend\cmd
go build -ldflags "-X main.BuildMode=%BMODE% -s -w" -o "..\..\%OUTDIR%\fst%EXT%" main.go
set GOOK=%ERRORLEVEL%
cd ..\..

if %GOOK% neq 0 (
    echo [ERROR] Go build failed: %LABEL%
    exit /b 1
)

if "%BMODE%" == "external" (
    echo     - Copying external assets...
    mkdir %OUTDIR%\dist >nul 2>&1
    xcopy /s /e /q /y frontend\dist\* %OUTDIR%\dist\ >nul
)

if exist .env copy /y .env %OUTDIR%\.env.example >nul

exit /b 0

:build_fail
echo.
echo [ERROR] Build aborted!
pause
exit /b 1

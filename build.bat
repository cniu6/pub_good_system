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
if exist dist rmdir /s /q dist
if exist backend\cmd\dist rmdir /s /q backend\cmd\dist

mkdir build >nul 2>&1
mkdir backend\cmd\dist >nul 2>&1
echo building... > backend\cmd\dist\index.html

echo.
echo [2/4] Building frontend (pnpm build)...
echo.

cd frontend
set VITE_BUILD_MODE=%BMODE%
call pnpm build
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Frontend build failed!
    set VITE_BUILD_MODE=
    cd ..
    exit /b 1
)
set VITE_BUILD_MODE=
cd ..

echo Copying frontend assets...
xcopy /s /e /q /y dist\* backend\cmd\dist\ >nul

echo.
echo [3/4] Cross-compiling Go backend... (Mode: %BMODE%)
echo.

set CGO_ENABLED=0

call :build_target windows amd64 .exe "Windows x64"
if %ERRORLEVEL% neq 0 goto :build_fail
call :build_target linux amd64 "" "Linux x64"
if %ERRORLEVEL% neq 0 goto :build_fail

set CGO_ENABLED=
set GOOS=
set GOARCH=

echo.
echo   Cleaning up intermediate artifacts...
if exist backend\cmd\dist rmdir /s /q backend\cmd\dist
if exist backend\main.exe del /q backend\main.exe

echo.
echo [4/4] Build complete!
echo ============================================================
echo   Output: ./build/
echo   Mode:   [%BMODE%]
echo ============================================================
echo.
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
go build -tags embedded -ldflags "-X main.BuildMode=%BMODE% -s -w" -o "..\..\%OUTDIR%\fst%EXT%" .
set GOOK=%ERRORLEVEL%
cd ..\..

if %GOOK% neq 0 (
    echo [ERROR] Go build failed: %LABEL%
    exit /b 1
)

if "%BMODE%" == "external" (
    echo     - Copying external assets...
    mkdir %OUTDIR%\dist >nul 2>&1
    xcopy /s /e /q /y dist\* %OUTDIR%\dist\ >nul
)

if exist .env.example copy /y .env.example %OUTDIR%\.env.example >nul

exit /b 0

:build_fail
echo.
echo [ERROR] Build aborted!
exit /b 1

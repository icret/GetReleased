@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

set "SERVER_EXE=backend\bin\server.exe"
if not defined SERVER_ADDR set "SERVER_ADDR=:8080"

:menu
cls
echo ============================================
echo   GetReleased Server Launcher
echo ============================================
echo   [1] Start server
echo   [2] Rebuild and start
echo   [3] Exit
echo ============================================
set "choice=1"
set /p "choice=Choose [1]: "

if "%choice%"=="1" goto start
if "%choice%"=="2" goto rebuild
if "%choice%"=="3" goto end
goto menu

:rebuild
echo [INFO] Rebuilding...
pushd backend
go build -o bin/server.exe ./cmd/server
if errorlevel 1 (
    echo [ERROR] Build failed
    popd
    pause
    goto menu
)
popd

:start
if not exist "%SERVER_EXE%" (
    echo [INFO] server.exe not found, building...
    pushd backend
    go build -o bin/server.exe ./cmd/server
    if errorlevel 1 (
        echo [ERROR] Build failed
        popd
        pause
        goto menu
    )
    popd
)

echo ============================================
echo   Listen addr : %SERVER_ADDR%
echo   Press Ctrl+C to exit
echo ============================================
"%SERVER_EXE%"
pause
goto menu

:end
exit /b 0
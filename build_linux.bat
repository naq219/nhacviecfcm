@echo off
REM -------------------------------
REM Build PocketBase Linux binary from Windows
REM -------------------------------

REM Thư mục lưu file build
set BUILD_DIR=%cd%\build_linux
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

REM Thiết lập GOOS và GOARCH cho Linux x64
set GOOS=linux
set GOARCH=amd64

REM Tên file output
set OUTPUT=%BUILD_DIR%\pocketbase_linux

REM Build
echo Building PocketBase for Linux...
go build -o "%OUTPUT%" ./cmd/pocketbase

IF %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b 1
)

echo Build successful! File saved to: %OUTPUT%
pause

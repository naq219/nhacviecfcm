@echo off
REM Build Linux binaries from Windows (Optimized)

echo 🐧 Building optimized Linux binaries...
echo.

cd /d %~dp0..

REM Build for Linux AMD64 (optimized)
echo 📦 Building optimized for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -trimpath -o bin/remiaq-linux-amd64 ./cmd/server

REM Build for Linux ARM64 (optimized)
echo 📦 Building optimized for Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -trimpath -o bin/remiaq-linux-arm64 ./cmd/server

echo.
echo ✅ Build completed!
echo 📁 Binaries are in the bin/ directory:
echo    - remiaq-linux-amd64 (for x86_64 systems)
echo    - remiaq-linux-arm64 (for ARM64 systems)
echo.
echo 📋 To deploy, copy the appropriate binary to your Linux server.

pause
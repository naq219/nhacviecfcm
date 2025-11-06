@echo off
REM Run remiaq server on Windows with environment setup

echo 🚀 Starting remiaq Server on Windows...
echo.

cd /d %~dp0..

REM Set environment variables for Windows
echo ⚙️  Setting environment variables...
set PB_ADDR=127.0.0.1:8090
set PB_CORS=*
set PB_CORS_ALLOWED_HEADERS=Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, X-File-Name
set PB_CORS_ALLOWED_METHODS=GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH
set PB_CORS_ALLOW_CREDENTIALS=true

REM Load additional environment variables from .env file if exists
if exist .env (
    echo 📝 Loading additional environment variables from .env...
    for /f "usebackq tokens=*" %%i in (.env) do (
        if not "%%i"=="" if not "%%i:~0,1"=="#" set %%i
    )
)

echo 🌐 Server will run on: %PB_ADDR%
echo 🔧 CORS enabled for all origins: %PB_CORS%
echo.

echo 📋 Current environment:
echo    PB_ADDR: %PB_ADDR%
echo    PB_CORS: %PB_CORS%
echo    PB_CORS_ALLOWED_HEADERS: %PB_CORS_ALLOWED_HEADERS%
echo    PB_CORS_ALLOWED_METHODS: %PB_CORS_ALLOWED_METHODS%
echo    PB_CORS_ALLOW_CREDENTIALS: %PB_CORS_ALLOW_CREDENTIALS%
echo.

echo 🚀 Starting Go server...
echo.

go run ./cmd/server serve

pause
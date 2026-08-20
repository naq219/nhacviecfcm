@echo off
REM Xóa toàn bộ file trong pb_data và chạy server Go

cd /d %~dp0..
echo 🧹 Đang xoá dữ liệu trong pb_data...
del /q pb_data\*
timeout /t 1 /nobreak >nul

echo 🚀 Đang khởi động server...
go run .\cmd\server serve

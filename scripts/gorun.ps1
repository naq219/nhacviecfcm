chcp 65001
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8


go run .\cmd\server\main.go serve 2>&1 | Select-String -Pattern "SELECT|INSERT|UPDATE|DELETE" -NotMatch
@echo off
echo Starting Backend Server...
start cmd /k "cd /d %~dp0 && go run server/main.go server/auth.go server/db.go server/handlers.go server/models.go"

timeout /t 2 /nobreak > nul

echo Starting Client GUI App...
start cmd /k "cd /d %~dp0 && go run client/main.go client/gui.go client/api_client.go client/crypto_utils.go"

echo.
echo Both applications are starting...
echo Close this window when done.

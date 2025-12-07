@echo off
echo Starting Backend Server...
start cmd /k "cd /d %~dp0 && go run server/main.go"

timeout /t 2 /nobreak > nul

echo Starting Client GUI App...
start cmd /k "cd /d %~dp0 && go run client/main.go"

echo.
echo Both applications are starting...
echo Close this window when done.

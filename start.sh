#!/bin/bash

echo "Starting Backend Server..."
cd "$(dirname "$0")"
go run server/main.go server/auth.go server/db.go server/handlers.go server/models.go &

sleep 2

echo "Starting Client GUI App..."
go run client/main.go client/gui.go client/api_client.go client/crypto_utils.go &

echo ""
echo "Both applications are starting..."
echo "Press Ctrl+C to stop all"

wait

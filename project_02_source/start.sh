#!/bin/bash

echo "Starting Backend Server..."
cd "$(dirname "$0")"
go run server/main.go &

sleep 2

echo "Starting Client GUI App..."
go run client/main.go &

echo ""
echo "Both applications are starting..."
echo "Press Ctrl+C to stop all"

wait

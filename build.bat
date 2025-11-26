@echo off
echo Building Server...
cd server
go build -o server.exe
cd ..

echo Building Client...
cd client
go build -o secure-notes.exe
cd ..

echo Done! Executable files created:
echo - server/server.exe
echo - client/secure-notes.exe

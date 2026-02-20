@echo off
setlocal

:: Load .env if it exists
if exist .env (
    echo ==^> Loading .env
    for /f "usebackq tokens=1,* delims==" %%A in (".env") do (
        set "%%A=%%B"
    )
)

echo ==^> Starting server...
calling_parents-windows-amd64.exe

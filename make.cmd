@echo off
setlocal
REM Windows entrypoint so `.\make build` works without GNU make.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0make.ps1" %*
exit /b %ERRORLEVEL%

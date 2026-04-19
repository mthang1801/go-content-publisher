@echo off
setlocal

cd /d "%~dp0"

if not exist "logs" mkdir "logs"

title Content Bot
echo Content Bot package runner
echo Working directory: "%CD%"
echo Log file: "%CD%\logs\app.log"
echo Press Ctrl+C to stop the app.
echo.

content-bot.exe run

echo.
echo content-bot.exe exited with code %ERRORLEVEL%.
pause

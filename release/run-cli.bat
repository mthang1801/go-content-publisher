@echo off
setlocal

cd /d "%~dp0"

if "%~1"=="" (
  title Content Bot CLI
  echo Content Bot CLI helper
  echo.
  echo Usage:
  echo   run-cli.bat help
  echo   run-cli.bat check-connections
  echo   run-cli.bat show-runtime
  echo   run-cli.bat settings-get auto_publish
  echo.
  cli.exe help
) else (
  title Content Bot CLI - %*
  cli.exe %*
)

echo.
pause

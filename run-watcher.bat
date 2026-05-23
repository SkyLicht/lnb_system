@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
set "EXE_PATH=%SCRIPT_DIR%file-watcher.exe"
set "CONFIG_PATH=%SCRIPT_DIR%watcher.config.json"

if not "%~1"=="" (
  set "CONFIG_PATH=%~1"
)

if not exist "%EXE_PATH%" (
  echo ERROR: Executable not found: "%EXE_PATH%"
  echo Build it first with:
  echo   go build -o bin\file-watcher.exe .\cmd\file-watcher
  exit /b 1
)

if not exist "%CONFIG_PATH%" (
  echo ERROR: Config file not found: "%CONFIG_PATH%"
  exit /b 1
)

"%EXE_PATH%" -config "%CONFIG_PATH%"
endlocal

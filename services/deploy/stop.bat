@echo off
setlocal enabledelayedexpansion

echo Stopping Language Exchange Bot...

REM Change to script directory
cd /d %~dp0

REM Stop all services
echo Stopping all services...
docker compose down

REM Remove orphans if any
docker compose down --remove-orphans

echo All services stopped.

endlocal

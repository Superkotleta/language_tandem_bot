@echo off
setlocal enabledelayedexpansion

echo Starting Language Exchange Bot...

REM Change to script directory
cd /d %~dp0

REM Start services
echo Starting Docker services...
docker compose up -d postgres pgadmin

REM Wait for postgres to be healthy
echo Waiting for database to be ready...
timeout /t 10 /nobreak > nul

REM Check database health
docker compose ps postgres | findstr "healthy" > nul
if %errorlevel% neq 0 (
    echo Database is not healthy yet, waiting more...
    timeout /t 15 /nobreak > nul
)

REM Start bot
echo Starting bot...
docker compose up -d bot

echo Bot is starting. Check logs with:
echo docker compose logs bot -f

endlocal

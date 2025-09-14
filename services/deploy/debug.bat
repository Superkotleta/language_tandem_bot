@echo off
setlocal enabledelayedexpansion

echo === Docker Debug Script ===

REM Check container status
echo === Container status ===
docker compose ps -a

REM Get PostgreSQL IP
echo.
echo === PostgreSQL IP ===
for /f "tokens=*" %%i in ('docker inspect pg -f "{{ .NetworkSettings.Networks.deploy_app-network.IPAddress }}" 2^>nul') do set PG_IP=%%i
if defined PG_IP (
    echo PostgreSQL IP: !PG_IP!
) else (
    echo Failed to get PostgreSQL IP
)

REM Check network connectivity
echo.
echo === Network status ===
docker network inspect deploy_app-network

REM Get bot logs
echo.
echo === Last 20 bot logs ===
docker compose logs bot --tail=20

REM Test database from inside container
echo.
echo === Test PostgreSQL connection ===
docker compose exec postgres psql -U langbot -d languagebot -c "SELECT count(*) FROM interest_translations;" 2>nul && (
    echo Database test successful
) || (
    echo Database connection failed
)

endlocal

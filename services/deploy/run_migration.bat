@echo off
echo Running feedback table migration...
cd /d %~dp0

REM Get container name
FOR /F "tokens=*" %%i IN ('docker ps -q --filter "label=com.docker.compose.service=postgres"') DO SET CONTAINER=%%i

if defined CONTAINER (
    echo Found PostgreSQL container: %CONTAINER%
    echo Executing migration script...
    docker exec -i %CONTAINER% psql -U postgres -d gigmate_db -f /docker-entrypoint-initdb.d/14-init-feedback.sql
    echo Migration completed!
) else (
    echo PostgreSQL container not found. Make sure the services are running.
)

echo Press any key to continue...
pause > nul

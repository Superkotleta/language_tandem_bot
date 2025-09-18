@echo off
REM Language Exchange Bot - Stop ngrok Script for Windows
REM Остановка ngrok процесса

echo 🛑 Остановка ngrok...

REM Останавливаем все процессы ngrok
taskkill /f /im ngrok.exe >nul 2>nul
if %errorlevel% equ 0 (
    echo ✅ ngrok остановлен
) else (
    echo ⚠️  ngrok процесс не найден
)

REM Удаляем временные файлы
if exist ngrok.log del ngrok.log >nul 2>nul
if exist ngrok.pid del ngrok.pid >nul 2>nul

echo ✅ ngrok полностью остановлен
pause

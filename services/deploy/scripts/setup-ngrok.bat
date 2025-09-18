@echo off
REM Language Exchange Bot - ngrok Setup Script for Windows
REM Автоматическая настройка ngrok для разработки

echo 🚀 Настройка ngrok для Language Exchange Bot
echo ==============================================

REM Проверяем, установлен ли ngrok
where ngrok >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ ngrok не установлен!
    echo.
    echo 📥 Установка ngrok:
    echo 1. Скачайте с https://ngrok.com/download
    echo 2. Распакуйте в папку PATH
    echo 3. Запустите скрипт снова
    echo.
    pause
    exit /b 1
)

echo ✅ ngrok найден

REM Проверяем авторизацию
ngrok config check >nul 2>nul
if %errorlevel% neq 0 (
    echo 🔐 Требуется авторизация ngrok
    echo 1. Зарегистрируйтесь на https://ngrok.com
    echo 2. Получите authtoken в панели управления
    echo 3. Выполните: ngrok config add-authtoken YOUR_TOKEN
    echo.
    set /p authtoken="Введите ваш ngrok authtoken: "
    ngrok config add-authtoken "%authtoken%"
)

echo ✅ ngrok авторизован

REM Создаем .env файл если его нет
if not exist .env (
    echo 📝 Создание .env файла...
    copy env.optimized.example .env >nul
    echo ✅ .env файл создан из примера
) else (
    echo ✅ .env файл уже существует
)

REM Запускаем ngrok в фоне
echo 🌐 Запуск ngrok...
start /b ngrok http 8080 --log=stdout > ngrok.log 2>&1

REM Ждем запуска ngrok
timeout /t 3 /nobreak >nul

REM Получаем URL (упрощенная версия для Windows)
echo ✅ ngrok запущен
echo 📝 Проверьте ngrok UI: http://localhost:4040
echo.

REM Обновляем .env файл
set /p ngrok_url="Введите ваш ngrok URL (например: https://abc123.ngrok.io): "
set webhook_url=%ngrok_url%/webhook/telegram

REM Обновляем .env файл (упрощенная версия)
powershell -Command "(Get-Content .env) -replace 'WEBHOOK_URL=.*', 'WEBHOOK_URL=%webhook_url%' | Set-Content .env"
powershell -Command "(Get-Content .env) -replace 'DEBUG=.*', 'DEBUG=false' | Set-Content .env"

echo ✅ .env файл обновлен:
echo    WEBHOOK_URL=%webhook_url%
echo    DEBUG=false

echo.
echo 🎯 Следующие шаги:
echo 1. Настройте TELEGRAM_TOKEN в .env файле
echo 2. Запустите бота: make -f Makefile.optimized up
echo 3. Настройте webhook: scripts\setup-webhook.bat
echo.
echo 🛑 Для остановки ngrok: scripts\stop-ngrok.bat
echo 📊 Для просмотра ngrok UI: http://localhost:4040

pause

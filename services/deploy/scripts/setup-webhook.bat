@echo off
REM Language Exchange Bot - Webhook Setup Script for Windows
REM Автоматическая настройка webhook в Telegram

echo 🔗 Настройка webhook для Language Exchange Bot
echo ==============================================

REM Проверяем .env файл
if not exist .env (
    echo ❌ .env файл не найден!
    echo Сначала запустите: scripts\setup-ngrok.bat
    pause
    exit /b 1
)

REM Загружаем переменные из .env
for /f "tokens=1,2 delims==" %%a in (.env) do (
    if "%%a"=="TELEGRAM_TOKEN" set TELEGRAM_TOKEN=%%b
    if "%%a"=="WEBHOOK_URL" set WEBHOOK_URL=%%b
)

REM Проверяем TELEGRAM_TOKEN
if "%TELEGRAM_TOKEN%"=="" (
    echo ❌ TELEGRAM_TOKEN не настроен в .env файле!
    echo.
    echo 📝 Настройте токен бота:
    echo 1. Создайте бота через @BotFather в Telegram
    echo 2. Получите токен
    echo 3. Добавьте в .env: TELEGRAM_TOKEN=your_actual_token
    pause
    exit /b 1
)

REM Проверяем WEBHOOK_URL
if "%WEBHOOK_URL%"=="" (
    echo ❌ WEBHOOK_URL не настроен в .env файле!
    echo Сначала запустите: scripts\setup-ngrok.bat
    pause
    exit /b 1
)

echo ✅ Токен бота: %TELEGRAM_TOKEN:~0,10%...
echo ✅ Webhook URL: %WEBHOOK_URL%

REM Проверяем, что ngrok работает
curl -s http://localhost:4040/api/tunnels >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ ngrok не запущен!
    echo Запустите: scripts\setup-ngrok.bat
    pause
    exit /b 1
)

REM Настраиваем webhook
echo 🔗 Настройка webhook в Telegram...
curl -s -X POST "https://api.telegram.org/bot%TELEGRAM_TOKEN%/setWebhook" -H "Content-Type: application/json" -d "{\"url\": \"%WEBHOOK_URL%\"}" > webhook_response.json

REM Проверяем ответ
findstr /C:"\"ok\":true" webhook_response.json >nul
if %errorlevel% equ 0 (
    echo ✅ Webhook успешно настроен!
) else (
    echo ❌ Ошибка настройки webhook:
    type webhook_response.json
    del webhook_response.json
    pause
    exit /b 1
)

REM Проверяем настройку
echo 🔍 Проверка webhook...
curl -s "https://api.telegram.org/bot%TELEGRAM_TOKEN%/getWebhookInfo" > webhook_info.json
type webhook_info.json

REM Проверяем, что бот запущен
curl -s http://localhost:8080/health >nul 2>nul
if %errorlevel% neq 0 (
    echo.
    echo ⚠️  Бот не запущен на порту 8080
    echo Запустите: make -f Makefile.optimized up
)

REM Очищаем временные файлы
del webhook_response.json >nul 2>nul
del webhook_info.json >nul 2>nul

echo.
echo 🎉 Настройка завершена!
echo.
echo 📊 Полезные ссылки:
echo    ngrok UI: http://localhost:4040
echo    Bot Health: http://localhost:8080/health
echo    Bot Metrics: http://localhost:8080/metrics
echo.
echo 🧪 Тестирование:
echo    Отправьте /start боту в Telegram
echo    Проверьте логи: make -f Makefile.optimized logs

pause

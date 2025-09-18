@echo off
REM Language Exchange Bot - Development Setup Script for Windows
REM Полная настройка для разработки с ngrok

echo 🚀 Language Exchange Bot - Development Setup
echo =============================================

REM Проверяем зависимости
echo 🔍 Проверка зависимостей...

REM Проверяем Docker
where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Docker не установлен!
    echo Установите Docker Desktop: https://docs.docker.com/desktop/windows/install/
    pause
    exit /b 1
)

REM Проверяем Docker Compose
where docker-compose >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Docker Compose не установлен!
    echo Установите Docker Compose: https://docs.docker.com/compose/install/
    pause
    exit /b 1
)

REM Проверяем curl
where curl >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ curl не установлен!
    echo Установите curl или используйте PowerShell
    pause
    exit /b 1
)

echo ✅ Все зависимости установлены

REM Создаем .env файл
if not exist .env (
    echo 📝 Создание .env файла...
    copy env.optimized.example .env >nul
    echo ✅ .env файл создан
) else (
    echo ✅ .env файл уже существует
)

REM Проверяем TELEGRAM_TOKEN
findstr /C:"TELEGRAM_TOKEN=your_telegram_bot_token_here" .env >nul
if %errorlevel% equ 0 (
    echo ⚠️  Требуется настройка TELEGRAM_TOKEN
    echo.
    echo 📝 Настройка токена бота:
    echo 1. Откройте Telegram и найдите @BotFather
    echo 2. Отправьте /newbot
    echo 3. Следуйте инструкциям
    echo 4. Скопируйте полученный токен
    echo.
    set /p token="Введите токен бота: "
    if not "%token%"=="" (
        powershell -Command "(Get-Content .env) -replace 'TELEGRAM_TOKEN=your_telegram_bot_token_here', 'TELEGRAM_TOKEN=%token%' | Set-Content .env"
        echo ✅ Токен сохранен
    ) else (
        echo ❌ Токен не введен
        pause
        exit /b 1
    )
) else (
    echo ✅ TELEGRAM_TOKEN уже настроен
)

REM Настраиваем ngrok
echo 🌐 Настройка ngrok...
call scripts\setup-ngrok.bat

REM Запускаем сервисы
echo 🐳 Запуск Docker сервисов...
make -f Makefile.optimized up

REM Ждем запуска
echo ⏳ Ожидание запуска сервисов...
timeout /t 10 /nobreak >nul

REM Настраиваем webhook
echo 🔗 Настройка webhook...
call scripts\setup-webhook.bat

echo.
echo 🎉 Настройка завершена!
echo.
echo 📊 Полезные ссылки:
echo    ngrok UI: http://localhost:4040
echo    Bot Health: http://localhost:8080/health
echo    Bot Metrics: http://localhost:8080/metrics
echo    Grafana: http://localhost:3000 (admin/admin)
echo    Prometheus: http://localhost:9090
echo    PgAdmin: http://localhost:5050 (admin@admin.com/admin)
echo.
echo 🧪 Тестирование:
echo    1. Отправьте /start боту в Telegram
echo    2. Проверьте логи: make -f Makefile.optimized logs
echo    3. Мониторинг: make -f Makefile.optimized monitor
echo.
echo 🛑 Остановка:
echo    make -f Makefile.optimized down
echo    scripts\stop-ngrok.bat

pause

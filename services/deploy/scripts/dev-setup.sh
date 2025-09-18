#!/bin/bash

# Language Exchange Bot - Development Setup Script
# Полная настройка для разработки с ngrok

set -e

echo "🚀 Language Exchange Bot - Development Setup"
echo "============================================="

# Проверяем зависимости
echo "🔍 Проверка зависимостей..."

# Проверяем Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен!"
    echo "Установите Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Проверяем Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не установлен!"
    echo "Установите Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# Проверяем jq
if ! command -v jq &> /dev/null; then
    echo "❌ jq не установлен!"
    echo "Установите jq:"
    echo "  Ubuntu/Debian: sudo apt install jq"
    echo "  Mac: brew install jq"
    echo "  Windows: choco install jq"
    exit 1
fi

echo "✅ Все зависимости установлены"

# Создаем .env файл
if [ ! -f .env ]; then
    echo "📝 Создание .env файла..."
    cp env.optimized.example .env
    echo "✅ .env файл создан"
else
    echo "✅ .env файл уже существует"
fi

# Проверяем TELEGRAM_TOKEN
if ! grep -q "TELEGRAM_TOKEN=your_telegram_bot_token_here" .env; then
    echo "✅ TELEGRAM_TOKEN уже настроен"
else
    echo "⚠️  Требуется настройка TELEGRAM_TOKEN"
    echo ""
    echo "📝 Настройка токена бота:"
    echo "1. Откройте Telegram и найдите @BotFather"
    echo "2. Отправьте /newbot"
    echo "3. Следуйте инструкциям"
    echo "4. Скопируйте полученный токен"
    echo ""
    read -p "Введите токен бота: " token
    if [ -n "$token" ]; then
        sed -i.bak "s|TELEGRAM_TOKEN=your_telegram_bot_token_here|TELEGRAM_TOKEN=$token|" .env
        echo "✅ Токен сохранен"
    else
        echo "❌ Токен не введен"
        exit 1
    fi
fi

# Настраиваем ngrok
echo "🌐 Настройка ngrok..."
./scripts/setup-ngrok.sh

# Запускаем сервисы
echo "🐳 Запуск Docker сервисов..."
make -f Makefile.optimized up

# Ждем запуска
echo "⏳ Ожидание запуска сервисов..."
sleep 10

# Настраиваем webhook
echo "🔗 Настройка webhook..."
./scripts/setup-webhook.sh

echo ""
echo "🎉 Настройка завершена!"
echo ""
echo "📊 Полезные ссылки:"
echo "   ngrok UI: http://localhost:4040"
echo "   Bot Health: http://localhost:8080/health"
echo "   Bot Metrics: http://localhost:8080/metrics"
echo "   Grafana: http://localhost:3000 (admin/admin)"
echo "   Prometheus: http://localhost:9090"
echo "   PgAdmin: http://localhost:5050 (admin@admin.com/admin)"
echo ""
echo "🧪 Тестирование:"
echo "   1. Отправьте /start боту в Telegram"
echo "   2. Проверьте логи: make -f Makefile.optimized logs"
echo "   3. Мониторинг: make -f Makefile.optimized monitor"
echo ""
echo "🛑 Остановка:"
echo "   make -f Makefile.optimized down"
echo "   ./scripts/stop-ngrok.sh"

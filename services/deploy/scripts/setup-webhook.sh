#!/bin/bash

# Language Exchange Bot - Webhook Setup Script
# Автоматическая настройка webhook в Telegram

set -e

echo "🔗 Настройка webhook для Language Exchange Bot"
echo "=============================================="

# Проверяем .env файл
if [ ! -f .env ]; then
    echo "❌ .env файл не найден!"
    echo "Сначала запустите: ./scripts/setup-ngrok.sh"
    exit 1
fi

# Загружаем переменные из .env
source .env

# Проверяем TELEGRAM_TOKEN
if [ -z "$TELEGRAM_TOKEN" ] || [ "$TELEGRAM_TOKEN" = "your_telegram_bot_token_here" ]; then
    echo "❌ TELEGRAM_TOKEN не настроен в .env файле!"
    echo ""
    echo "📝 Настройте токен бота:"
    echo "1. Создайте бота через @BotFather в Telegram"
    echo "2. Получите токен"
    echo "3. Добавьте в .env: TELEGRAM_TOKEN=your_actual_token"
    exit 1
fi

# Проверяем WEBHOOK_URL
if [ -z "$WEBHOOK_URL" ] || [ "$WEBHOOK_URL" = "https://yourdomain.com/webhook/telegram" ]; then
    echo "❌ WEBHOOK_URL не настроен в .env файле!"
    echo "Сначала запустите: ./scripts/setup-ngrok.sh"
    exit 1
fi

echo "✅ Токен бота: ${TELEGRAM_TOKEN:0:10}..."
echo "✅ Webhook URL: $WEBHOOK_URL"

# Проверяем, что ngrok работает
if ! curl -s http://localhost:4040/api/tunnels > /dev/null; then
    echo "❌ ngrok не запущен!"
    echo "Запустите: ./scripts/setup-ngrok.sh"
    exit 1
fi

# Настраиваем webhook
echo "🔗 Настройка webhook в Telegram..."
RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_TOKEN/setWebhook" \
     -H "Content-Type: application/json" \
     -d "{\"url\": \"$WEBHOOK_URL\"}")

# Проверяем ответ
if echo "$RESPONSE" | grep -q '"ok":true'; then
    echo "✅ Webhook успешно настроен!"
else
    echo "❌ Ошибка настройки webhook:"
    echo "$RESPONSE"
    exit 1
fi

# Проверяем настройку
echo "🔍 Проверка webhook..."
WEBHOOK_INFO=$(curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getWebhookInfo")
echo "$WEBHOOK_INFO" | jq '.'

# Проверяем, что бот запущен
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo ""
    echo "⚠️  Бот не запущен на порту 8080"
    echo "Запустите: make -f Makefile.optimized up"
fi

echo ""
echo "🎉 Настройка завершена!"
echo ""
echo "📊 Полезные ссылки:"
echo "   ngrok UI: http://localhost:4040"
echo "   Bot Health: http://localhost:8080/health"
echo "   Bot Metrics: http://localhost:8080/metrics"
echo ""
echo "🧪 Тестирование:"
echo "   Отправьте /start боту в Telegram"
echo "   Проверьте логи: make -f Makefile.optimized logs"

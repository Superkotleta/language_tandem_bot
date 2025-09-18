#!/bin/bash

# Language Exchange Bot - ngrok Setup Script
# Автоматическая настройка ngrok для разработки

set -e

echo "🚀 Настройка ngrok для Language Exchange Bot"
echo "=============================================="

# Проверяем, установлен ли ngrok
if ! command -v ngrok &> /dev/null; then
    echo "❌ ngrok не установлен!"
    echo ""
    echo "📥 Установка ngrok:"
    echo "Windows: скачайте с https://ngrok.com/download"
    echo "Linux: wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz"
    echo "Mac: brew install ngrok"
    echo ""
    echo "После установки запустите скрипт снова."
    exit 1
fi

echo "✅ ngrok найден: $(ngrok version)"

# Проверяем авторизацию
if ! ngrok config check &> /dev/null; then
    echo "🔐 Требуется авторизация ngrok"
    echo "1. Зарегистрируйтесь на https://ngrok.com"
    echo "2. Получите authtoken в панели управления"
    echo "3. Выполните: ngrok config add-authtoken YOUR_TOKEN"
    echo ""
    read -p "Введите ваш ngrok authtoken: " authtoken
    ngrok config add-authtoken "$authtoken"
fi

echo "✅ ngrok авторизован"

# Создаем .env файл если его нет
if [ ! -f .env ]; then
    echo "📝 Создание .env файла..."
    cp env.optimized.example .env
    echo "✅ .env файл создан из примера"
fi

# Запускаем ngrok в фоне
echo "🌐 Запуск ngrok..."
ngrok http 8080 --log=stdout > ngrok.log 2>&1 &
NGROK_PID=$!

# Ждем запуска ngrok
sleep 3

# Получаем URL
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')

if [ "$NGROK_URL" = "null" ] || [ -z "$NGROK_URL" ]; then
    echo "❌ Не удалось получить ngrok URL"
    kill $NGROK_PID 2>/dev/null || true
    exit 1
fi

echo "✅ ngrok запущен: $NGROK_URL"
echo "📝 PID процесса: $NGROK_PID"

# Обновляем .env файл
WEBHOOK_URL="${NGROK_URL}/webhook/telegram"
sed -i.bak "s|WEBHOOK_URL=.*|WEBHOOK_URL=$WEBHOOK_URL|" .env
sed -i.bak "s|DEBUG=.*|DEBUG=false|" .env

echo "✅ .env файл обновлен:"
echo "   WEBHOOK_URL=$WEBHOOK_URL"
echo "   DEBUG=false"

# Сохраняем PID для остановки
echo $NGROK_PID > ngrok.pid

echo ""
echo "🎯 Следующие шаги:"
echo "1. Настройте TELEGRAM_TOKEN в .env файле"
echo "2. Запустите бота: make -f Makefile.optimized up"
echo "3. Настройте webhook: ./scripts/setup-webhook.sh"
echo ""
echo "🛑 Для остановки ngrok: ./scripts/stop-ngrok.sh"
echo "📊 Для просмотра ngrok UI: http://localhost:4040"

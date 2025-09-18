#!/bin/bash

# Language Exchange Bot - Stop ngrok Script
# Остановка ngrok процесса

echo "🛑 Остановка ngrok..."

# Останавливаем ngrok по PID
if [ -f ngrok.pid ]; then
    NGROK_PID=$(cat ngrok.pid)
    if kill -0 $NGROK_PID 2>/dev/null; then
        kill $NGROK_PID
        echo "✅ ngrok остановлен (PID: $NGROK_PID)"
    else
        echo "⚠️  ngrok процесс не найден"
    fi
    rm -f ngrok.pid
else
    echo "⚠️  Файл ngrok.pid не найден"
fi

# Останавливаем все процессы ngrok
pkill -f "ngrok http" 2>/dev/null || true

# Удаляем временные файлы
rm -f ngrok.log

echo "✅ ngrok полностью остановлен"

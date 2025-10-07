# 🔐 Настройка Secrets для GitHub Actions

## 📋 **Обязательные Secrets**

Добавьте в **Settings → Secrets and variables → Actions**:

### 🐳 **Docker Hub**

```shell
DOCKER_USERNAME=your-dockerhub-username
DOCKER_TOKEN=your-dockerhub-token
```

### 🖥️ **Production Server**

```shell
SERVER_HOST=your-server-ip-or-domain
SERVER_USER=your-username
SERVER_SSH_KEY=your-private-ssh-key
SERVER_URL=https://your-domain.com
```

### 📱 **Telegram Notifications**

```shell
TELEGRAM_BOT_TOKEN=your-bot-token
TELEGRAM_CHAT_ID=your-chat-id
```

### 💬 **Slack Notifications**

```shell
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR_WORKSPACE_ID/YOUR_CHANNEL_ID/YOUR_WEBHOOK_TOKEN
```

## 🛠️ **Как получить токены:**

### **Docker Hub Token:**

1. Зайдите на [hub.docker.com](https://hub.docker.com)
2. Settings → Security → New Access Token
3. Выберите "Read, Write, Delete"

### **Telegram Bot Token:**

1. Напишите [@BotFather](https://t.me/botfather)
2. `/newbot` → выберите имя
3. Скопируйте токен

### **Telegram Chat ID:**

1. Напишите [@userinfobot](https://t.me/userinfobot)
2. Скопируйте ваш ID

### **SSH Key:**

```bash
# Создайте SSH ключ
ssh-keygen -t ed25519 -C "github-actions"

# Скопируйте приватный ключ в GitHub Secrets
cat ~/.ssh/id_ed25519
```

## 🚀 **После настройки:**

1. **Push в main** → автоматический deploy
2. **Создайте тег** → автоматический релиз
3. **Мониторинг** → каждые 5 минут
4. **Уведомления** → в Telegram/Slack

## 📊 **Доступные workflows:**

- ✅ **CI/CD Pipeline** - тестирование и сборка
- 🚀 **Deploy** - развертывание в production
- 📱 **Notifications** - уведомления в Telegram
- 🏃 **Performance** - нагрузочное тестирование
- 🏷️ **Release** - автоматические релизы
- 🔍 **Monitoring** - мониторинг production

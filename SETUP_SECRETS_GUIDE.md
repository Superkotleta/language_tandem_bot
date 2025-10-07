# 🔐 Полное руководство по настройке GitHub Secrets

## 📍 **Шаг 1: Переход в настройки**

1. Откройте ваш репозиторий на GitHub
2. Нажмите **Settings** (вкладка справа)
3. В левом меню выберите **Secrets and variables** → **Actions**
4. Нажмите **New repository secret**

## 🔑 **Шаг 2: Добавляем все необходимые secrets**

### **🐳 Docker Hub Secrets**

#### **DOCKER_USERNAME**
```
Название: DOCKER_USERNAME
Значение: ваш-dockerhub-username
```

#### **DOCKER_TOKEN**
```
Название: DOCKER_TOKEN
Значение: ваш-dockerhub-token
```

**Как получить Docker Hub токен:**
1. Зайдите на [hub.docker.com](https://hub.docker.com)
2. Войдите в аккаунт
3. Settings → Security → New Access Token
4. Название: `github-actions`
5. Права: **Read, Write, Delete**
6. Скопируйте токен

---

### **🖥️ Production Server Secrets**

#### **SERVER_HOST**
```
Название: SERVER_HOST
Значение: your-server-ip-or-domain.com
```

#### **SERVER_USER**
```
Название: SERVER_USER
Значение: root (или ваш-пользователь)
```

#### **SERVER_SSH_KEY**
```
Название: SERVER_SSH_KEY
Значение: -----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABlwAAAAdzc2gtcn
... (весь приватный ключ)
-----END OPENSSH PRIVATE KEY-----
```

#### **SERVER_URL**
```
Название: SERVER_URL
Значение: https://your-domain.com
```

**Как создать SSH ключ:**
```bash
# Создайте SSH ключ специально для GitHub Actions
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_actions

# Скопируйте ПРИВАТНЫЙ ключ в GitHub Secrets
cat ~/.ssh/github_actions

# Добавьте ПУБЛИЧНЫЙ ключ на сервер
ssh-copy-id -i ~/.ssh/github_actions.pub user@your-server.com
```

---

### **📱 Telegram Notifications**

#### **TELEGRAM_BOT_TOKEN**
```
Название: TELEGRAM_BOT_TOKEN
Значение: 1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
```

#### **TELEGRAM_CHAT_ID**
```
Название: TELEGRAM_CHAT_ID
Значение: 123456789
```

**Как создать Telegram бота:**
1. Напишите [@BotFather](https://t.me/botfather)
2. Отправьте `/newbot`
3. Введите имя бота: `Language Exchange Bot CI`
4. Введите username: `language_exchange_ci_bot`
5. Скопируйте токен

**Как получить Chat ID:**
1. Напишите [@userinfobot](https://t.me/userinfobot)
2. Скопируйте ваш ID
3. Или добавьте бота в группу и получите ID группы

---

### **💬 Slack Notifications (опционально)**

#### **SLACK_WEBHOOK_URL**
```
Название: SLACK_WEBHOOK_URL
Значение: https://hooks.slack.com/services/YOUR_WORKSPACE_ID/YOUR_CHANNEL_ID/YOUR_WEBHOOK_TOKEN
```

**Как создать Slack webhook:**
1. Зайдите в [api.slack.com](https://api.slack.com/apps)
2. Create New App → From scratch
3. App Name: `Language Exchange Bot`
4. Workspace: выберите ваш workspace
5. Incoming Webhooks → Activate Incoming Webhooks
6. Add New Webhook to Workspace
7. Выберите канал для уведомлений
8. Скопируйте Webhook URL

---

## 🧪 **Шаг 3: Тестирование secrets**

### **Создайте тестовый workflow:**

```yaml
name: Test Secrets

on:
  workflow_dispatch:

jobs:
  test-secrets:
    runs-on: ubuntu-latest
    steps:
    - name: Test Docker secrets
      run: |
        echo "Docker username: ${{ secrets.DOCKER_USERNAME }}"
        echo "Docker token: ${#SECRETS.DOCKER_TOKEN} characters"
    
    - name: Test Server secrets
      run: |
        echo "Server host: ${{ secrets.SERVER_HOST }}"
        echo "Server user: ${{ secrets.SERVER_USER }}"
        echo "SSH key length: ${#SECRETS.SERVER_SSH_KEY} characters"
    
    - name: Test Telegram secrets
      run: |
        echo "Bot token: ${#SECRETS.TELEGRAM_BOT_TOKEN} characters"
        echo "Chat ID: ${{ secrets.TELEGRAM_CHAT_ID }}"
```

## ✅ **Шаг 4: Проверка настройки**

После добавления всех secrets:

1. **Перейдите в Actions** → ваш репозиторий
2. **Запустите тестовый workflow**
3. **Проверьте логи** - не должно быть ошибок
4. **Проверьте уведомления** в Telegram/Slack

## 🚨 **Важные замечания:**

- ⚠️ **Никогда не коммитьте secrets в код!**
- 🔒 **Используйте только GitHub Secrets**
- 🧪 **Тестируйте на staging перед production**
- 📝 **Документируйте все secrets для команды**

## 📋 **Чек-лист secrets:**

- [ ] `DOCKER_USERNAME`
- [ ] `DOCKER_TOKEN`
- [ ] `SERVER_HOST`
- [ ] `SERVER_USER`
- [ ] `SERVER_SSH_KEY`
- [ ] `SERVER_URL`
- [ ] `TELEGRAM_BOT_TOKEN`
- [ ] `TELEGRAM_CHAT_ID`
- [ ] `SLACK_WEBHOOK_URL` (опционально)

## 🎯 **После настройки:**

1. **Push в main** → автоматический deploy
2. **Создайте тег** → автоматический релиз
3. **Мониторинг** → каждые 5 минут
4. **Уведомления** → в Telegram/Slack

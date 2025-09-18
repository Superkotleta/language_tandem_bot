
# Фазы проекта

## Раздел 1: Написание программы

### Фаза 1: Подготовка и запуск "Hello, World" на локальной машине

**Цель:** Создать минимально работающего бота, который отвечает на сообщения, и убедиться, что вся связка (бот -> ngrok -> ваш код) работает корректно.

1. **Создание бота в Telegram**
    * Найдите в Telegram официального бота `@BotFather`.
    * Отправьте ему команду `/newbot`.
    * Следуйте инструкциям: придумайте имя для бота (которое видят пользователи) и его username (уникальный, заканчивается на `bot`).
    * **Результат:** BotFather пришлет вам **токен** (API Token). Это ключ к вашему боту, сохраните его в надежном месте и никому не показывайте.
2. **Настройка локального окружения**
    * Установите на свой компьютер **Go**.
    * Установите **PostgreSQL**. Создайте базу данных для вашего проекта (например, `my_telegram_bot_db`).
    * Скачайте и распакуйте **ngrok**.
3. **Написание первого кода на Go**
    * Создайте новый проект на Go.
    * Выберите библиотеку для работы с Telegram Bot API. Рекомендуемые варианты: `go-telegram-bot-api` или `telego`.
    * Напишите простейший код, который:
        * Инициализирует бота с вашим токеном.
        * Запускает веб-сервер (например, на порту `8080`), который будет слушать входящие запросы от Telegram.
        * Создает обработчик (хендлер), который на любое полученное сообщение отвечает "Привет, мир!".
4. **Запуск и проверка через ngrok**
    * В одном окне терминала запустите вашего Go-бота. Он должен начать слушать порт `8080`.
    * В другом окне терминала запустите `ngrok` командой: `ngrok http 8080`.
    * `ngrok` выдаст вам публичный URL вида `https://случайная-строка.ngrok-free.app`. **Скопируйте этот HTTPS-адрес.**
    * "Сообщите" Telegram этот адрес, отправив специальный запрос для установки вебхука. Это можно сделать через `curl` или прямо из вашего Go-кода при старте. URL для запроса будет выглядеть так:
`https://api.telegram.org/botВАШ_ТОКЕН/setWebhook?url=https://случайная-строка.ngrok-free.app`
    * **Проверка:** Зайдите в Telegram, найдите своего бота и отправьте ему любое сообщение. Он должен ответить "Привет, мир!". В окне терминала, где запущен `ngrok`, вы увидите входящие запросы.

### Фаза 2: Разработка основной логики и работа с данными

**Цель:** Научить бота выполнять осмысленные действия и сохранять информацию о пользователях в базе данных.

1. **Проектирование базы данных**
    * Определите, какие данные вам нужно хранить. Для начала это может быть простая таблица `users`:
        * `id` (уникальный идентификатор)
        * `telegram_id` (ID пользователя в Telegram, уникальный)
        * `username` (имя пользователя)
        * `first_name`
        * `created_at` (дата регистрации)
        * `state` (для хранения состояния в диалоге, например, "ожидает_ввода_имени")
2. **Подключение к PostgreSQL из Go**
    * Добавьте в ваш Go-проект драйвер для работы с PostgreSQL (например, `pgx`).
    * Напишите функции для подключения к базе данных и для выполнения базовых операций (CRUD - Create, Read, Update, Delete) с пользователями. Например, функция `FindOrCreateUser`, которая при первом сообщении от пользователя добавляет его в базу данных.
3. **Реализация команд и логики**
    * Расширьте код бота:
        * Обрабатывайте команду `/start`: приветствуйте пользователя и регистрируйте его в БД.
        * Реализуйте простую логику на основе состояний. Например, если бот задал вопрос "Как вас зовут?", он переводит пользователя в состояние `ожидает_ввода_имени` и следующий ответ воспринимает как имя.
        * Добавьте обработку кнопок (inline-клавиатур) для более удобного взаимодействия.
        * Добавьте обработку для отправки заданий: новая команда /admin_send_tasks интегрирует с таблицей tasks, рассылая уведомления пользователям (используя Telegram API для массовой отправки).
        * Обновите состояния для учета статусов профилей ('ready', 'matched' и т.д.) при обработке сообщений.

### Фаза 3: Структурирование проекта и логирование

**Цель:** Сделать проект готовым к росту и упростить отладку.

1. **Структурированное логирование**
    * Вместо простых `fmt.Println()` используйте библиотеку для логирования (например, `logrus` или `slog`).
    * Настройте логирование в файлы в формате JSON. Записывайте важные события: запуск бота, получение сообщения, ошибки при работе с БД, действия пользователей. Это бесценно для отладки на сервере.
2. **Модульный дизайн для масштабируемости**
Структурируйте код с учетом future-proof — используйте интерфейсы в Go для абстракции (например, интерфейс `Matcher` для алгоритма подбора, чтобы легко заменить на ML-версию). Разделите на микросервисы: отдельный сервис для matching (интегрируется с RabbitMQ), другой для уведомлений. Это позволит добавлять фичи (как групповые сессии) без рефакторинга всего проекта.
3. **Рефакторинг кода**
    * Разбейте ваш код на логические модули (пакеты):
        * `main.go`: точка входа, запуск сервера.
        * `telegram/`: все, что связано с обработкой запросов от Telegram.
        * `storage/` или `db/`: функции для работы с базой данных.
        * `models/`: структуры данных (например, `User`).
        * `config/`: чтение конфигурации (токен, параметры БД) из переменных окружения или файла.

### Фаза 4: Развертывание на VPS

**Цель:** Перенести бота с локальной машины на "боевой" сервер, чтобы он работал 24/7 с учетом dockerization.

Dockerfile теперь охватывает не только бота, но и бэкенд с БД. Используйте multi-stage build для Go, чтобы контейнер был легким. Пример `Dockerfile` для бота:

```yml
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bot .
CMD ["./bot"]
```

Для PostgreSQL используйте официальный образ в `docker-compose.yml`. Обновленный `docker-compose.yml` (для бэка, БД и Nginx):

```yml
version: '3.8'
services:
  bot:
    build: .
    ports:
      - "8080:8080"
    environment:
      - TELEGRAM_TOKEN=your_token
      - DB_URL=postgres://user:pass@db:5432/dbname
    depends_on:
      - db

  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
      - POSTGRES_DB=dbname
    volumes:
      - pgdata:/var/lib/postgresql/data

  nginx:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - bot

volumes:
  pgdata:
```

Добавьте RabbitMQ в `docker-compose.yml`:

```yml
version: '3.8'
services:
  # ... (existing services: bot, db, nginx, prometheus, grafana)

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"  # AMQP port
      - "15672:15672"  # Management UI
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  rabbitmq_data:
```

В `Dockerfile` для бота добавьте зависимость от RabbitMQ в коде (библиотека `amqp`).

* **Добавление observability и интеграция с observability.**
Для мониторинга RabbitMQ добавьте exporter в Prometheus (image `kbudde/rabbitmq-exporter`). В `prometheus.yml` укажите скрейпинг метрик с порта 15672. Grafana покажет дашборды по очередям (длина, throughput).

Интегрируйте Prometheus для метрик (например, количество матчей, время ответа бота) и Grafana для дашбордов. В Go добавьте экспортер метрик (библиотека `prometheus/client_golang`). В `docker-compose.yml` добавьте сервисы:

```yml
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

Для логирования используйте ELK (Elasticsearch, Logstash, Kibana) или Loki (см. PLAN TBD ниже). Это позволит мониторить производительность в реальном времени.

1. **Подготовка к развертыванию (Docker)**
    * Напишите `Dockerfile` для вашего Go-приложения. Он описывает, как собрать ваш код в легковесный контейнер.
    * Используйте `docker-compose.yml` для описания всей инфраструктуры:
        * Сервис вашего бота (собирается из `Dockerfile`).
        * Сервис базы данных `PostgreSQL` (используется готовый образ).
        * Сервис `Nginx` (для проксирования запросов).
Это позволит запустить весь проект на VPS одной командой.
2. **Настройка VPS**
    * Арендуйте VPS у любого провайдера.
    * Подключитесь к нему по SSH.
    * Установите Docker и Docker Compose.
    * Настройте брандмауэр, открыв порты 22 (SSH), 80 (HTTP) и 443 (HTTPS).
3. **Развертывание и финальная настройка**
    * Скопируйте ваш проект (включая `Dockerfile` и `docker-compose.yml`) на VPS.
    * Запустите все сервисы командой `docker-compose up -d`.
    * Купите доменное имя и направьте его на IP-адрес вашего VPS.
    * Настройте Nginx для проксирования запросов на контейнер с вашим ботом.
    * Используйте `certbot` для автоматического получения и настройки бесплатного SSL-сертификата от Let's Encrypt.
    * **Финальный шаг:** Обновите вебхук в Telegram, указав ваш **постоянный публичный URL** (`https://ваш-домен.com`).
4. **Информационная безопасность (инфобез)**

    * Добавьте шифрование данных: используйте HTTPS everywhere (уже через Nginx и Let's Encrypt), шифруйте чувствительные данные в PostgreSQL (библиотека `pgcrypto` для полей как `telegram_id`). Внедрите аутентификацию: JWT-токены для admin-команд (библиотека `github.com/golang-jwt/jwt`), чтобы предотвратить несанкционированный доступ. Храните секреты (токены, DB-пароли) в Docker secrets или HashiCorp Vault для production.
5. **Защита от атак**

    * Внедрите rate limiting в боте (библиотека `github.com/ulule/limiter` для ограничения запросов по IP/пользователю, напр. 100/мин). Для DDoS используйте Cloudflare (бесплатный план) как прокси перед VPS — настройте в `docker-compose.yml` интеграцию с их API. Добавьте CAPTCHA для регистрации (через Telegram Login или Google reCAPTCHA) и мониторинг аномалий в Grafana (из observability). В RabbitMQ настройте quotas на очереди, чтобы избежать перегрузки.
6. **Future-proof развертывание**
    * Добавьте Kubernetes-готовность: сделайте `docker-compose.yml` совместимым с Docker Swarm или Minikube для будущего оркестрирования. Интегрируйте auto-scaling (например, через VPS-провайдера как DigitalOcean с их App Platform) для обработки роста пользователей.

Следуя этому плану, вы постепенно, шаг за шагом, создадите надежного и масштабируемого бота, получив при этом ценный практический опыт работы с современным стеком технологий.

### Фаза 5: Мониторинг, безопасность и масштабирование

**Цель:** Обеспечить устойчивость и рост.
    **Действия:**
    - Настройте alerting в Grafana (уведомления в Telegram о сбоях).
    - Проводите security audits (инструмент `gosec` для сканирования Go-кода).
    - Добавьте A/B-тестирование для алгоритмов (библиотека `github.com/feature-flags` для переключения между версиями matching).
    - Подготовьте миграцию на облако (AWS/GCP) с Terraform для инфраструктуры как кода.

## Раздел 2: "План CI/CD и деплой"

**Общая идея:** CI/CD автоматизирует тесты, сборку и деплой. На GitHub это делается через GitHub Actions — бесплатно для публичных репозиториев, просто настраивается YAML-файлами. Пайплайн запускается на пуш в ветки `develop` (для тестов) и `release` (для деплоя на VPS).

**Примерный план ввода CI/CD:**

1. **Настройка репозитория:** Создайте GitHub-репо, добавьте секреты (Settings > Secrets and variables > Actions) для токенов (TELEGRAM_TOKEN, SSH_KEY для VPS, DB creds).
2. **YAML-файл для пайплайна (`./github/workflows/ci-cd.yml`):**

```yml
name: CI/CD Pipeline

on:
  push:
    branches: [develop, release]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v3
        with: { go-version: 1.22 }
      - name: Run tests
        run: go test ./...

  build:
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/release'
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t mybot:latest .
      - name: Push to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      - run: docker push mybot:latest

  deploy:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/release'
    steps:
      - name: Deploy to VPS
        uses: appleboy/ssh-action@v0.1.4
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            docker pull mybot:latest
            docker-compose down
            docker-compose up -d
```

   **Шаги внедрения:**
    - На `develop`: Автоматические тесты (go test).
    - На `release`: Сборка Docker-образа, пуш в registry (Docker Hub), деплой на VPS via SSH.
    - Тестируйте локально, затем коммитьте. Если нужно, добавьте уведомления в Telegram о статусе билда.

Это базовый план; расширьте тестами (unit, integration) по мере роста проекта.

## Раздел 3: "GUI для БД и синхронизация с Google Sheets"

**Решение:** Для GUI используйте pgAdmin (веб-интерфейс для PostgreSQL) — добавьте его в `docker-compose.yml` как сервис:

```yml
pgadmin:
  image: dpage/pgadmin4
  environment:
    - PGADMIN_DEFAULT_EMAIL=admin@admin.com
    - PGADMIN_DEFAULT_PASSWORD=admin
  ports:
    - "5050:80"
  depends_on:
    - db
```

Доступ через браузер (localhost:5050), с возможностью редактирования таблиц (users, match_queue и т.д.).

**Синхронизация в Google Sheets в реальном времени:** (НЕ УВЕРЕН ЧТО БУДУ РЕАЛИЗОВЫВАТЬ НО ПОСМОТРИМ) Используйте Go-скрипт с библиотекой `google.golang.org/api/sheets/v4` для экспорта данных. Добавьте cron-job в боте:

* Авторизуйтесь через Google API (OAuth2).
* Пример функции:

```go
func syncToSheets() {
    // Подключение к Sheets API
    srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
    // Чтение из PostgreSQL
    rows := queryDB("SELECT * FROM users")
    // Запись в Sheet
    valueRange := &sheets.ValueRange{Values: rows}
    srv.Spreadsheets.Values.Update("SHEET_ID", "A1", valueRange).ValueInputOption("RAW").Do()
}
```

Запускайте по расписанию (библиотека `github.com/robfig/cron`). Это обеспечит "прямой эфир" — обновления каждые 5-10 мин.

## Раздел 4: Комментарии к PLAN TBD

* **База данных для Логирования:** Рекомендую Loki with Grafana — легковесно, интегрируется с Prometheus для observability. Loki хранит логи как индексированные метки, Grafana визуализирует. Добавьте в Docker: `loki` image и настройте bot для отправки логов (библиотека `grafana/loki-client-go`). Redis подойдет для кэша, но не для долгосрочных логов.
* **CI/CD:** Как выше, GitHub Actions на `develop` (тесты) и `release` (деплой). GitLab аналогичен, но GitHub проще для новичков.
* **Docker:** Да, развертывание бэка и БД в Docker (см. обновленную фазу 4) — это стандарт для reproducibility.
* **Брокер сообщений (REST API):** Если планируете несколько ботов (масштаб), начните с RabbitMQ — проще в setup, подходит для очередей уведомлений/матчей. Добавьте сервис в `docker-compose`:

```yml
rabbitmq:
  image: rabbitmq:3-management
  ports:
    - "5672:5672"
    - "15672:15672"
```

Интегрируйте в Go с `github.com/streadway/amqp`. Kafka для high-throughput, но overkill сейчас — добавьте позже, если трафик вырастет. Сделайте сразу, если хотите future-proof: RabbitMQ не усложнит код.

* **Алгоритм подбора совместимых партнеров**

* **Предложения по алгоритмам:**
  * **Улучшенный matching:** Держите в уме графовый алгоритм (библиотека `gonum/graph`) для подбора — моделируйте пользователей как узлы, совместимость как ребра, ищите максимальное паросочетание (Hungarian algorithm для оптимизации). Это лучше текущего SQL-запроса при большом количестве пользователей.
  * **Анти-фрод алгоритм:** Добавьте детекцию ботов/спама — анализируйте паттерны (частота сообщений, entropy текста) с библиотекой `github.com/texttheater/golang-levenshtein` для проверки на дубликаты профилей.
  * **Держать в уме:** Генетический алгоритм для оптимизации (библиотека `github.com/MaxHalford/eaopt`) — для сложных сценариев с множественными критериями, как волны подбора.

**Workflow выполнения алгоритма:**

* **Фаза 2: Запуск алгоритма** (обновлено): Добавьте валидацию на безопасность — проверяйте входные данные на SQL-инъекции (хотя pgx уже sanitized), логируйте подозрительную активность в Loki.

**Планы развития алгоритма** (обновлено):

* **Фаза расширения 3** (новое): Интеграция blockchain для верификации профилей (опционально, для high-trust сценариев) и AI-модерация чатов (с Hugging Face моделями для детекции токсичности).

**Рекомендации по внедрению :**

* **Future-proof советы:** Используйте 12-factor app принципы (конфиг в env, stateless сервисы) для легкой миграции. Тестируйте с нагрузкой (инструмент `locust`) для симуляции 1000+ пользователей.
* **Безопасность на будущее:** Соблюдайте GDPR (анонимизация данных в logs), добавьте backup БД (pg_dump в cron-job).
* **Почему эти добавления?** Они делают проект resilient: security предотвращает утечки, алгоритмы улучшают core-функцию (подбор), future-proof обеспечивает рост без перестройки

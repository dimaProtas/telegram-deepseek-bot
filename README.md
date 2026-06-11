# Telegram DeepSeek Bot

Telegram-бот для взаимодействия с [DeepSeek API](https://api.deepseek.com) и [OpenCode Agent](https://github.com/anomalyco/opencode). Поддерживает многоходовые диалоги с контекстом, работу с кодом и несколько моделей DeepSeek.

## Возможности

- **Чат с ИИ** — диалог с сохранением контекста разговора
- **OpenCode Agent** (`/run`) — выполнение задач через локальный OpenCode Agent
- **Переключение моделей** (`/mode`) — `chat`, `coder`, `reasoner`
- **Генерация кода** (`/code`) — написание кода с использованием модели Coder
- **Ревью кода** (`/review <файл>`) — анализ локального файла
- **Объяснение кода** (`/explain <файл>`) — построчное объяснение логики
- **Генерация тестов** (`/test <файл>`) — создание unit-тестов
- **Документирование** (`/docs <файл>`) — генерация документации
- **Статистика** (`/state`) — просмотр потребления токенов
- **Повтор** (`/retry`) — повтор последнего запроса
- **Очистка контекста** (`/clear`) — сброс истории диалога
- **Access control** — белый список пользователей по User ID
- **SOCKS5 прокси** — опциональная поддержка прокси для Telegram API и DeepSeek API

## Требования

- **Go 1.26+**
- Токен Telegram бота ([@BotFather](https://t.me/BotFather))
- API-ключ [DeepSeek](https://platform.deepseek.com)
- [OpenCode CLI](https://github.com/anomalyco/opencode) (для `/run`)

## Установка

```bash
# Клонирование
git clone <repo-url> telegram-deepseek-bot
cd telegram-deepseek-bot

# Сборка
go build -o bot ./cmd/bot/

# Настройка
cp .env.example .env
# Отредактируйте .env, указав свои токены
```

## Конфигурация

Все настройки задаются в `.env`:

| Переменная | По умолчанию | Описание |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | — | Токен бота (обязательно) |
| `DEEPSEEK_API_KEY` | — | API-ключ DeepSeek (обязательно) |
| `DEEPSEEK_MODEL` | `deepseek-chat` | Модель по умолчанию |
| `DEEPSEEK_API_URL` | `https://api.deepseek.com/v1` | Базовый URL API |
| `DEEPSEEK_MAX_TOKENS` | `4096` | Максимальная длина ответа |
| `DEEPSEEK_TEMPERATURE` | `0.7` | Креативность (0–2) |
| `ALLOWED_USER_IDS` | — | ID разрешённых пользователей через запятую (пусто — открыто для всех) |
| `ALLOWED_CHAT_IDS` | — | ID разрешённых чатов через запятую |
| `LOG_LEVEL` | `info` | Уровень логирования (`debug`/`info`/`warn`/`error`) |
| `CONVERSATION_TTL_HOURS` | `24` | Время жизни истории диалога |
| `OPENCODE_ENDPOINT` | `http://localhost:8080` | Эндпоинт OpenCode Agent |
| `OPENCODE_TIMEOUT` | `600` | Таймаут выполнения OpenCode (сек) |
| `OPENCODE_WORKSPACE` | `.` | Рабочая директория для OpenCode |
| `PROXY_ADDR` | — | Адрес SOCKS5 прокси (опционально) |
| `PROXY_USERNAME` | — | Логин прокси |
| `PROXY_PASSWORD` | — | Пароль прокси |

## Запуск

```bash
./bot
# или
go run ./cmd/bot/
```

Бот запускается в режиме long-polling и пишет в stdout: `Bot is running. Press Ctrl+C to stop.`

## Команды бота

| Команда | Действие |
|---|---|
| `/start`, `/help` | Приветствие и список команд |
| Произвольный текст | Отправить запрос в DeepSeek |
| `/chat <текст>` | Явный чат-запрос |
| `/code <запрос>` | Генерация кода |
| `/mode <модель>` | Сменить модель (`chat`/`coder`/`reasoner`) |
| `/run <запрос>` | Запуск задачи через OpenCode Agent |
| `/review <файл>` | Ревью локального файла |
| `/explain <файл>` | Объяснение кода |
| `/test <файл>` | Генерация тестов |
| `/docs <файл>` | Генерация документации |
| `/state` | Статистика токенов |
| `/retry` | Повторить последний запрос |
| `/clear` | Очистить историю диалога |

## Структура проекта

```
.
├── cmd/bot/main.go              # Точка входа
├── internal/
│   ├── config/config.go         # Загрузка конфигурации
│   ├── deepseek/client.go       # HTTP-клиент DeepSeek API
│   ├── logger/logger.go         # Логгер
│   ├── opencode/agent.go        # OpenCode Agent
│   ├── storage/conversations.go # In-memory хранилище диалогов
│   └── telegram/
│       ├── bot.go               # Оркестрация, роутинг команд
│       ├── handlers.go          # Обработчик чата
│       ├── file_handlers.go     # Команды для работы с файлами
│       ├── mode_handlers.go     # /mode и /code
│       └── state_handlers.go    # /state и /retry
├── data/                        # Директория для runtime-данных
├── pkg/                         # (зарезервировано)
├── go.mod / go.sum              # Зависимости Go
├── .env.example                 # Шаблон конфигурации
├── .gitignore
└── README.md
```

## Технологии

- **Go 1.26** — язык реализации
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) — Telegram Bot API
- [DeepSeek API](https://api.deepseek.com) — LLM-модели
- [OpenCode](https://github.com/anomalyco/opencode) — AI-агент для выполнения задач
- In-memory хранилище с TTL — контекст диалогов
- SOCKS5 прокси — поддержка обхода блокировок

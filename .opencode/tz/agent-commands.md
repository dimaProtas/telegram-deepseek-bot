# Техническое задание: Команды для ВСЕХ агентов OpenCode в Telegram-боте

## 1. Цель

Добавить Telegram-команды для **всех оставшихся** специализированных агентов OpenCode, для которых ещё нет команд. На данный момент реализовано 7 агентских команд из 17 доступных агентов (плюс `/run` для default-агента). Задача — покрыть все 10 оставшихся агентов, чтобы пользователь мог вызвать **любого** агента по короткой интуитивной команде.

**Проблема**: пользователь вынужден вручную описывать роль в промпте для `/run`, надеясь, что opencode сам выберет нужного агента. Это ненадёжно и требует лишних усилий. Отдельные команды гарантируют вызов правильного агента и улучшают UX.

## 2. Контекст

### 2.1. Текущая архитектура (AS-IS)

Архитектура Command Router (паттерн Command/Strategy) **уже полностью реализована**:

- **Интерфейс `Command`** (`internal/telegram/commands.go:9-13`) — контракт `Name()` / `Matches()` / `Execute()`
- **`CommandRouter`** (`internal/telegram/commands.go:15-36`) — реестр команд с `Register()` и `Dispatch()`
- **`exactCommand`**, **`prefixCommand`**, **`fallbackCommand`** — адаптеры для существующих обработчиков
- **`AgentCommand`** (`internal/telegram/agent_commands.go:12-73`) — параметризованная структура для opencode-агентов:
  - Поля: `prefix`, `agentName`, `displayName`
  - Конструктор: `NewAgentCommand(prefix, agentName, displayName string)`
  - `Execute()`: валидация промпта (пустой → подсказка, > 4096 → ошибка), отправка статуса, запуск горутины с таймаутом, логирование `[AGENT]`
- **`registerAgentCommands()`** (`internal/telegram/agent_commands.go:75-93`) — регистрирует 7 агентов в роутере
- **`Agent.ExecuteWithAgent()`** (`internal/opencode/agent.go:45-76`) — выполняет `opencode run <prompt> --dangerously-skip-permissions --agent <agentName>`
- **`Agent.buildArgs()`** (`internal/opencode/agent.go:78-84`) — сборка аргументов командной строки
- **`OpenCodeDefaultAgent`** в конфиге (`internal/config/config.go:26`) — позволяет переопределить агента по умолчанию для `/run`

**Диспетчеризация** (`bot.go:60-122`): после `registerAgentCommands()` регистрируются `/run`, `/start`, `/help`, `/clear`, `/state`, `/retry`, `/mode`, `/code`, файловые команды (`/review`, `/explain`, `/test`, `/docs`), `/chat` и fallback.

### 2.2. Уже реализованные агентские команды

| Префикс команды | Имя агента `--agent` | Отображаемое имя |
|---|---|---|
| `/explore` | `explore` | Explore |
| `/go-review` | `go-reviewer` | Go Review |
| `/go-senior` | `go-senior` | Go Senior Developer |
| `/react-dev` | `react-developer` | React Developer |
| `/react-review` | `react-reviewer` | React Review |
| `/write-tests` | `test-writer` | Test Writer |
| `/tz` | `tz-writer` | ТЗ Writer |
| `/run` | (пусто — default agent) | OpenCode Agent |

### 2.3. Агенты, для которых НЕТ команд (цель задачи)

| Имя агента в opencode | Назначение (из конфигурации opencode) |
|---|---|
| `bash-linux` | Bash, Linux, файлы, cron, rsync, права доступа, shell-скрипты |
| `clickhouse-sql` | Эксперт по ClickHouse SQL — оптимизация запросов, MergeTree, TTL, batch processing |
| `data-pipeline-architect` | Архитектор data pipelines — CSV, архивы, ClickHouse, Postgres, Kestra, идемпотентность |
| `db-integration` | Интеграция Go-кода с ClickHouse и Postgres — драйверы, типы, batch insert, транзакции |
| `debugging` | Диагностика ошибок в Go, ClickHouse, Postgres, Bash, Linux, Kestra |
| `documentation` | Техническая документация — README, инструкции, пайплайны, troubleshooting, конфигурация |
| `kestra` | Kestra — flow YAML, schedules, triggers, concurrency, shell tasks, Docker runner, KV secrets |
| `orchestrator` | Главный агент-маршрутизатор для задач по Go, ClickHouse, Postgres, Kestra |
| `postgres-sql` | Эксперт по PostgreSQL — SQL, индексы, миграции, транзакции, EXPLAIN ANALYZE |
| `refactoring` | Рефакторинг Go, SQL и Bash — сохраняя поведение, делая код проще и читабельнее |

**Исключённые агенты:**

- `general` — уже покрыт командой `/run` (default-агент opencode)
- `README` — вызывается только вручную пользователем, команда не нужна

### 2.4. Как opencode вызывает конкретного агента

```bash
opencode run <prompt> --agent <agentName> --dangerously-skip-permissions
```

Это уже реализовано в `buildArgs()` (`internal/opencode/agent.go:78-84`).

### 2.5. Файлы, затрагиваемые изменениями

| Файл | Текущая роль | Что меняется |
|---|---|---|
| `internal/telegram/agent_commands.go` | `AgentCommand` + `registerAgentCommands()` с 7 агентами | Добавить 10 новых агентов в список |
| `internal/telegram/bot.go` | `sendHelp()` с 7 агентскими командами | Добавить 10 новых команд в текст справки |
| `internal/telegram/commands.go` | Интерфейсы `Command`, `CommandRouter` | **Без изменений** — архитектура готова |
| `internal/opencode/agent.go` | `ExecuteWithAgent()`, `buildArgs()` | **Без изменений** — уже параметризован |
| `internal/config/config.go` | `OpenCodeDefaultAgent` | **Без изменений** — поле уже есть |
| `.env.example` | Комментарий `OPENCODE_DEFAULT_AGENT` | **Без изменений** — уже задокументирован |

## 3. Требования

### 3.1. Функциональные требования

**FR-01**: Бот должен поддерживать 10 новых команд для вызова оставшихся opencode-агентов:

| Команда | Агент `--agent` | Отображаемое имя |
|---|---|---|
| `/bash` | `bash-linux` | Bash/Linux |
| `/clickhouse` | `clickhouse-sql` | ClickHouse SQL |
| `/pipeline` | `data-pipeline-architect` | Data Pipeline Architect |
| `/dbint` | `db-integration` | DB Integration |
| `/debug` | `debugging` | Debugging |
| `/docgen` | `documentation` | Documentation |
| `/kestra` | `kestra` | Kestra |
| `/orchestrate` | `orchestrator` | Orchestrator |
| `/postgres` | `postgres-sql` | PostgreSQL |
| `/refactor` | `refactoring` | Refactoring |

**FR-02**: Все существующие команды (7 агентских + `/run` + файловые + управление) должны продолжить работать без изменений.

**FR-03**: Для каждой новой команды бот должен:
- Показать статусное сообщение вида «⏳ Запуск агента <displayName>...»
- Выполнить opencode с параметром `--agent <agentName>`
- Вернуть результат пользователю (с разбивкой `sendLongMessage()` на части по 4096 символов)
- При ошибке — показать понятное сообщение «❌ Агент <displayName> завершился с ошибкой: <текст>»
- Логировать факт вызова в формате `[AGENT] agent=<name> userID=<id> chatID=<id> elapsed=<dur> status=<success|error>`

**FR-04**: Валидация пустого промпта: сообщение «Укажите задачу после команды. Пример: /bash <задача>».

**FR-05**: Команда `/help` должна отображать актуальный список **всех** 17 агентских команд (+ `/run`), сгруппированных по категориям.

**FR-06**: Команда `/docgen` (агент `documentation`) не должна конфликтовать с существующей `/docs` (файловый обработчик через DeepSeek API). Обоснование выбора `/docgen` — см. раздел 5.

**FR-07**: Добавление нового агента не требует изменения архитектуры — достаточно добавить одну строку в `registerAgentCommands()` и одну строку в `sendHelp()`.

### 3.2. Нефункциональные требования

1. **NFR-01**: Время отклика на команду (статусное сообщение) — не более 200 мс.
2. **NFR-02**: Таймаут выполнения агента — настраиваемый (из `OPENCODE_TIMEOUT`, по умолчанию 600 с).
3. **NFR-03**: Логирование всех вызовов с указанием имени агента, chatID, userID, длительности и статуса.
4. **NFR-04**: Код соответствует OCP — добавление агента = добавление записи в список, без изменения логики.
5. **NFR-05**: Обратная совместимость — все существующие команды работают без изменений.
6. **NFR-06**: Ограничение длины промпта — 4096 символов (валидация в `ExecuteWithAgent()`).

## 4. Архитектурное решение

### 4.1. Статус архитектуры

Command Router (паттерн Command/Strategy) **уже реализован**. Никаких архитектурных изменений не требуется. Текущая архитектура:

```
Telegram API ──► Bot.handleMessage() ──► CommandRouter.Dispatch()
                                              │
                                              │ итерация commands[]
                                              ▼
                                         AgentCommand.Matches(prefix)
                                              │
                                              ▼
                                         AgentCommand.Execute()
                                              │
                                              ├─ sendMessage("⏳ Запуск агента ...")
                                              └─ go func() {
                                                    agent.ExecuteWithAgent(ctx, prompt, agentName)
                                                 }
```

Расширение сводится к добавлению новых записей в список агентов в `registerAgentCommands()`.

### 4.2. `AgentCommand` — без изменений

Структура `AgentCommand` (`internal/telegram/agent_commands.go:12-73`) уже параметризована тремя полями:
- `prefix` — префикс команды (например, `"/bash"`)
- `agentName` — имя агента для флага `--agent` (например, `"bash-linux"`)
- `displayName` — человекочитаемое имя для статусных сообщений и логов (например, `"Bash/Linux"`)

Все 10 новых команд создаются как экземпляры `AgentCommand` через `NewAgentCommand()` — **ни одной новой структуры или функции не требуется**.

### 4.3. Порядок регистрации команд

Новые агентские команды добавляются в `registerAgentCommands()` в любом порядке — они не конфликтуют по префиксам ни между собой, ни с существующими командами. Порядок проверки в `CommandRouter.Dispatch()` для них не критичен, так как префиксы уникальны.

Важно: `registerAgentCommands()` вызывается **первым** в `registerCommands()` (`bot.go:61`), до регистрации `/run` и файловых команд. Это гарантирует, что `/bash`, `/debug`, `/refactor` и другие не будут перехвачены fallback-обработчиком.

### 4.4. Диаграмма взаимодействия (без изменений)

```
Пользователь
  │
  │  /bash проверь права доступа в ./scripts/
  ▼
Telegram API ───► Bot.Start() → b.router.Dispatch(b, msg)
                                    │
                                    ▼
                              AgentCommand("/bash").Matches() → true
                                    │
                                    ▼
                              AgentCommand.Execute()
                                    │
                                    ├─► sendMessage("⏳ Запуск агента Bash/Linux...")
                                    │
                                    └─► go func() {
                                          ctx, cancel := context.WithTimeout(...)
                                          result, err := agent.ExecuteWithAgent(ctx, prompt, "bash-linux")
                                          ...
                                        }
```

## 5. Конфликт имён: `/docs` vs агент `documentation`

### 5.1. Проблема

Существующая команда `/docs <файл>` (`bot.go:111-113`) читает локальный файл и отправляет его содержимое в DeepSeek API для генерации документации. Если назвать команду для агента `documentation` тоже `/docs`, возникнет конфликт: неясно, какой обработчик должен срабатывать.

### 5.2. Решение: `/docgen`

Команда для агента `documentation` получает имя `/docgen` (doc generate / documentation generator).

**Обоснование:**
- Не конфликтует с `/docs` (существующая файловая команда)
- Семантически понятно: «сгенерировать документацию»
- Короче альтернатив (`/document`, `/write-docs`, `/gen-docs`)
- Укладывается в паттерн «глагол» как `/write-tests`, `/refactor`

### 5.3. Примечание

В будущем, если функциональность файловых команд будет перенесена на opencode-агентов, `/docs` и `/docgen` можно будет объединить. Но сейчас это разные механизмы (DeepSeek API через бота vs OpenCode Agent с доступом к workspace), поэтому разделение оправдано.

## 6. Обоснование названий новых команд

| Команда | Агент `--agent` | Почему так названа |
|---|---|---|
| `/bash` | `bash-linux` | Коротко, очевидно, ассоциация с shell/bash |
| `/clickhouse` | `clickhouse-sql` | Узнаваемое имя продукта, `/ch` было бы непонятно |
| `/pipeline` | `data-pipeline-architect` | Короче полного имени, понятно для целевой аудитории |
| `/dbint` | `db-integration` | «db integration» в сокращении, `/db-integration` тоже допустимо |
| `/debug` | `debugging` | Глагол, интуитивно понятен |
| `/docgen` | `documentation` | «doc generate», не конфликтует с `/docs` |
| `/kestra` | `kestra` | Имя продукта, конфликтов не предвидится |
| `/orchestrate` | `orchestrator` | Глагол, «оркестрировать задачу» |
| `/postgres` | `postgres-sql` | Узнаваемое имя продукта, `/pg` было бы непонятно |
| `/refactor` | `refactoring` | Глагол, интуитивно понятен |

**Альтернативы (отклонены):**
- `/linux` вместо `/bash` — слишком широко, агент заточен именно под bash/shell
- `/ch` вместо `/clickhouse` — непонятно для новичков
- `/pg` вместо `/postgres` — непонятно для новичков
- `/orch` вместо `/orchestrate` — неочевидное сокращение
- `/doc-write` вместо `/docgen` — длиннее, менее идиоматично

## 7. Обработка ошибок

### 7.1. Сценарии ошибок (без изменений относительно уже реализованного)

| Сценарий | Сообщение пользователю | Логирование |
|---|---|---|
| Пустой промпт | «Укажите задачу после команды. Пример: /bash <задача>» | WARN |
| Промпт > 4096 символов | Возвращается error из `ExecuteWithAgent()` | ERROR |
| Таймаут выполнения | «❌ Агент <displayName> завершился с ошибкой: timeout exceeded...» | ERROR с elapsed |
| Ошибка выполнения агента | «❌ Агент <displayName> завершился с ошибкой: <первые 200 символов stderr>» | ERROR |
| Пустой результат (успех, нет вывода) | «✅ Агент <displayName> выполнил задачу (ответ пуст)» | INFO |

### 7.2. Таймауты

- Единый таймаут `OPENCODE_TIMEOUT` (по умолчанию 600 с) для всех агентов
- При таймауте процесс `opencode` уничтожается через `context.Context` (уже реализовано в `ExecuteWithAgent`)
- Механизм `/retry` **не применяется** к opencode-агентам (они не сохраняют историю диалогов)

## 8. План реализации

### Этап 1: Добавить новых агентов в `registerAgentCommands()`

**Файл**: `internal/telegram/agent_commands.go`

Добавить 10 новых записей в слайс `agents` внутри `registerAgentCommands()`:

```go
{"/bash", "bash-linux", "Bash/Linux"},
{"/clickhouse", "clickhouse-sql", "ClickHouse SQL"},
{"/pipeline", "data-pipeline-architect", "Data Pipeline Architect"},
{"/dbint", "db-integration", "DB Integration"},
{"/debug", "debugging", "Debugging"},
{"/docgen", "documentation", "Documentation"},
{"/kestra", "kestra", "Kestra"},
{"/orchestrate", "orchestrator", "Orchestrator"},
{"/postgres", "postgres-sql", "PostgreSQL"},
{"/refactor", "refactoring", "Refactoring"},
```

**Объём изменений**: ~10 строк. Логика `AgentCommand.Execute()` не меняется — она уже параметризована.

### Этап 2: Обновить `sendHelp()`

**Файл**: `internal/telegram/bot.go`, метод `sendHelp()` (строка 239)

Обновить секцию «🤖 OpenCode агенты» в тексте справки, добавив 10 новых команд. Сгруппировать по назначению для читаемости:

```
🤖 OpenCode агенты:
/run <задача> — общий агент (по умолчанию)

💻 Go-разработка:
/go-senior <задача> — написание Go-кода (Senior уровень)
/go-review <промпт> — ревью Go-кода
/debug <задача> — диагностика ошибок
/refactor <задача> — рефакторинг кода

🗄️ Базы данных:
/postgres <задача> — PostgreSQL (SQL, индексы, миграции)
/clickhouse <задача> — ClickHouse SQL (оптимизация запросов)
/dbint <задача> — интеграция Go с БД

⚙️ Инфраструктура:
/bash <задача> — Bash, Linux, shell-скрипты
/kestra <задача> — Kestra (flow YAML, triggers)
/pipeline <задача> — проектирование data pipelines
/orchestrate <задача> — оркестрация задач

🌐 Frontend:
/react-dev <задача> — разработка React-компонентов
/react-review <промпт> — ревью React-кода

📝 Документирование и тестирование:
/tz <задача> — создание технического задания
/docgen <задача> — генерация документации
/write-tests <промпт> — генерация тестов
/explore <задача> — исследование кодовой базы
```

Существующие секции «💬 Чат с DeepSeek», «📄 Работа с файлами», «⚙️ Управление» остаются без изменений.

**Объём изменений**: ~25 строк в help-тексте.

### Этап 3: Проверка и финальные штрихи

- Убедиться, что все 10 команд регистрируются и отображаются в `/help`
- Проверить отсутствие пересечения префиксов с существующими командами
- Проверить консистентность имён агентов с конфигурацией opencode
- Убедиться, что `go build` и `go vet` проходят без ошибок

### Этап 4: Тестирование (опционально, по запросу)

**Файл**: `internal/telegram/agent_commands_test.go` (создать при необходимости)
- Table-driven тесты: `Matches()` для каждого из 10 новых префиксов
- Тесты: пустой промпт → сообщение с подсказкой
- Тесты: промпт с аргументом → вызов `ExecuteWithAgent` с правильным `agentName`

## 9. Изменения в существующих файлах (сводка)

| Файл | Тип изменения | Описание |
|---|---|---|
| `internal/telegram/agent_commands.go` | **Изменить** | Добавить 10 записей в слайс `agents` внутри `registerAgentCommands()` |
| `internal/telegram/bot.go` | **Изменить** | Обновить текст в `sendHelp()` — добавить 10 новых команд |
| `internal/telegram/commands.go` | **Без изменений** | Интерфейсы и роутер уже готовы |
| `internal/opencode/agent.go` | **Без изменений** | `ExecuteWithAgent()` и `buildArgs()` уже параметризованы |
| `internal/config/config.go` | **Без изменений** | `OpenCodeDefaultAgent` уже есть |
| `.env.example` | **Без изменений** | `OPENCODE_DEFAULT_AGENT` уже задокументирован |

## 10. Риски и компромиссы

### 10.1. Риски

| Риск | Вероятность | Влияние | Митигация |
|---|---|---|---|
| opencode меняет имена агентов (например, `bash-linux` → `bash`) | Низкая | Низкое | Имена зафиксированы в одном месте — `registerAgentCommands()`. Легко обновить. |
| Команда `/dbint` непонятна пользователям | Средняя | Низкое | При необходимости переименовать в `/db-integration` — это одна строка в списке агентов. |
| Команда `/orchestrate` может быть воспринята как «оркестрация музыки» | Низкая | Низкое | Целевая аудитория — разработчики, контекст очевиден. |
| Рост количества команд в `/help` делает справку слишком длинной | Средняя | Низкое | Сгруппированы по категориям. При дальнейшем росте можно рассмотреть пагинацию или `/help <категория>`. |
| Пересечение workspace между параллельными вызовами агентов | Низкая | Высокое | Уже существующий риск, не специфичен для новых команд. Решается настройкой `OPENCODE_WORKSPACE`. |

### 10.2. Компромиссы

1. **Не добавляются подкоманды** (типа `/agent bash-linux`). Обоснование: пользователь явно запросил отдельные команды, и архитектура `AgentCommand` уже заточена под плоский список команд. Это проще, чем парсинг подкоманд, и улучшает discoverability.

2. **Все команды однотипны** — одинаковые статусные сообщения, одинаковый формат ошибок. Это осознанное решение для консистентности UX. Если в будущем какой-то агент потребует уникального поведения — `AgentCommand` можно расширить или создать отдельную реализацию `Command`.

3. **`sendHelp()` — хардкод**, а не динамическая генерация из списка команд. Для 17 агентских команд это приемлемо. При росте до 30+ стоит задуматься об авто-генерации из роутера.

4. **Имена команд на английском**, несмотря на русскоязычную аудиторию. Обоснование: техническая аудитория, short-команды на английском — стандарт в Telegram-ботах. Русские транслитерации (`/баш`, `/отладка`) были бы неудобны для ввода и неочевидны.

### 10.3. Альтернативные решения (отклонены)

**A. Отдельные обработчики для каждого агента** (вместо `AgentCommand`):
- Отклонено: нарушает DRY, 10+ одинаковых структур с копипастой логики

**B. Команда `/agent <имя> <задача>`** — один обработчик с параметром:
- Отклонено: хуже UX, пользователь должен помнить точные имена агентов

**C. Динамическая регистрация через конфиг** (список агентов в `.env`):
- Отклонено: избыточно для текущего масштаба, усложняет валидацию имён агентов

## 11. Метрики приёмки

- [ ] Все 10 новых команд зарегистрированы и доступны: `/bash`, `/clickhouse`, `/pipeline`, `/dbint`, `/debug`, `/docgen`, `/kestra`, `/orchestrate`, `/postgres`, `/refactor`
- [ ] Каждая команда вызывает правильного агента (соответствие `--agent <name>`)
- [ ] Все 7 существующих агентских команд продолжают работать без изменений
- [ ] `/run` работает без изменений
- [ ] `/docs` (файловая команда) и `/docgen` (агент) не конфликтуют
- [ ] `/help` отображает актуальный список всех 17 агентских команд (+ `/run`)
- [ ] Пустой промпт для любой новой команды выдаёт подсказку с примером
- [ ] Таймаут корректно обрывает выполнение для всех агентов
- [ ] Логи содержат `[AGENT] agent=<имя>` для всех новых команд
- [ ] `go build` и `go vet` проходят без ошибок
- [ ] Обратная совместимость: все старые команды (`/test`, `/review`, `/explain`, `/chat` и др.) работают

# Техническое задание: Команды для специализированных OpenCode-агентов

## 1. Цель

Расширить Telegram-бота командами для вызова **каждого специализированного агента OpenCode** по отдельности. Сейчас доступна только одна общая команда `/run`, которая вызывает opencode с агентом по умолчанию (general). Пользователь хочет иметь отдельные команды для всех доступных агентов:

- `explore` — исследование кодовой базы
- `go-reviewer` — code review Go-кода уровня Senior/Staff Engineer
- `go-senior` — написание production-ready Go-кода и архитектуры
- `react-developer` — разработка React-компонентов
- `react-reviewer` — code review React-кода
- `test-writer` — написание unit/integration тестов для Go
- `tz-writer` — создание технических заданий

**Проблема**: единственная команда `/run` не позволяет выбрать агента. Пользователь вынужден формулировать промпт так, чтобы opencode сам догадался, какой агент использовать, что ненадёжно.

## 2. Контекст

### 2.1. Текущая архитектура (AS-IS)

**Диспетчеризация команд** (`handleMessage()` в `internal/telegram/bot.go:91-171`):
- Цепочка `if/else if` проверяет префикс команды в тексте сообщения
- 10+ проверок подряд
- Каждая проверка жёстко завязана на строковый префикс (например, `strings.HasPrefix(text, "/run ")`)

**Вызов OpenCode** (`internal/opencode/agent.go:41-72`):
- Метод `Execute(ctx, prompt)` собирает аргументы:
  ```
  opencode run <prompt> --dangerously-skip-permissions
  ```
- Агент не передаётся — используется default-агент opencode
- Ограничение длины промпта: 4096 символов

**Обработчик `/run`** (`handleRun()` в `bot.go:178-208`):
- Отправляет статусное сообщение «⏳ Запуск OpenCode Agent...»
- Запускает горутину с таймаутом из `cfg.OpenCodeTimeout`
- Логирует время начала, успех/ошибку, длительность
- Результат отправляет через `sendLongMessage()` (разбивка на части по 4096 символов)

**Существующая команда `/test`** (`file_handlers.go:9-56`):
- Работает с **локальными файлами** через DeepSeek API (не через opencode)
- Формирует промпт для генерации тестов и отправляет в DeepSeek
- Конфликтует по имени с будущей командой для агента `test-writer`

### 2.2. Доступные opencode-агенты

| Имя агента | Назначение | Предлагаемая команда |
|---|---|---|
| `explore` | Исследование кодовой базы | `/explore` |
| `general` | Общие задачи (агент по умолчанию) | `/run` (оставить как есть) |
| `go-reviewer` | Code review Go-кода (Senior/Staff) | `/go-review` |
| `go-senior` | Написание production-ready Go-кода | `/go-senior` |
| `react-developer` | Разработка React-компонентов | `/react-dev` |
| `react-reviewer` | Code review React-кода | `/react-review` |
| `test-writer` | Написание тестов для Go | `/write-tests` |
| `tz-writer` | Создание технических заданий | `/tz` |

### 2.3. Как opencode вызывает конкретного агента

Согласно документации opencode, для вызова конкретного агента используется флаг `--agent`:

```bash
opencode run <prompt> --agent go-senior --dangerously-skip-permissions
```

Текущий вызов (без `--agent`):
```bash
opencode run <prompt> --dangerously-skip-permissions
```

## 3. Требования

### 3.1. Функциональные требования

1. **FR-01**: Бот должен поддерживать 7 новых команд для вызова специализированных opencode-агентов:
   - `/explore <задача>` → агент `explore`
   - `/go-review <промпт>` → агент `go-reviewer`
   - `/go-senior <задача>` → агент `go-senior`
   - `/react-dev <задача>` → агент `react-developer`
   - `/react-review <промпт>` → агент `react-reviewer`
   - `/write-tests <промпт>` → агент `test-writer`
   - `/tz <задача>` → агент `tz-writer`

2. **FR-02**: Существующая команда `/run` должна продолжить работать без изменений (агент по умолчанию, без явного `--agent`).

3. **FR-03**: Для каждой новой команды бот должен:
   - Показать статусное сообщение вида «⏳ Запуск агента <имя>...»
   - Выполнить opencode с параметром `--agent <имя_агента>`
   - Вернуть результат пользователю (с разбивкой на части, если ответ длинный)
   - При ошибке — показать понятное сообщение
   - Логировать факт вызова, имя агента, длительность, успех/ошибку

4. **FR-04**: Валидация: пустой промпт (только команда без аргументов) → сообщение с подсказкой о формате команды.

5. **FR-05**: Команда `/help` должна отображать актуальный список всех доступных команд, включая новые.

6. **FR-06**: Команда `/write-tests` не должна конфликтовать с существующей `/test` (которая работает с файлами через DeepSeek API). Обе команды должны сосуществовать.

7. **FR-07**: Обработчик команд должен быть расширяемым — добавление новой команды не должно требовать изменения существующего кода диспетчеризации.

### 3.2. Нефункциональные требования

1. **NFR-01**: Время отклика на команду (статусное сообщение) — не более 200 мс.
2. **NFR-02**: Таймаут выполнения агента — настраиваемый (из `OPENCODE_TIMEOUT`).
3. **NFR-03**: Логирование всех вызовов с указанием имени агента, chatID, userID, длительности и статуса.
4. **NFR-04**: Код должен следовать принципам SOLID, особенно OCP (Open/Closed Principle) — диспетчер команд должен быть открыт для расширения, закрыт для модификации.
5. **NFR-05**: Обратная совместимость — все существующие команды должны работать без изменений в поведении.
6. **NFR-06**: Ограничение длины промпта — 4096 символов (как и сейчас).

## 4. Архитектурное решение

### 4.1. Проблема текущей архитектуры

Текущий `handleMessage()` — это цепочка из ~15 `if/else if`. При добавлении 7 новых команд она вырастет до ~22 проверок, что:
- Нарушает OCP (нужно менять код диспетчера для каждой новой команды)
- Сложно тестировать (все проверки в одном методе)
- Трудно читать и поддерживать

### 4.2. Предлагаемое решение: Command Dispatcher (Pattern: Command / Strategy)

**Идея**: выделить диспетчеризацию команд в отдельную абстракцию — `CommandRouter`.

```
                   ┌─────────────────────┐
                   │     handleMessage   │
                   │   (bot.go)          │
                   └─────────┬───────────┘
                             │
                             ▼
                   ┌─────────────────────┐
                   │   CommandRouter     │
                   │   (commands.go)     │
                   │                     │
                   │  commands []Command │
                   │  + Register(cmd)    │
                   │  + Dispatch(msg)    │
                   └─────────┬───────────┘
                             │
                             ▼ iterates
              ┌──────────────────────────────┐
              │         Command              │
              │  (interface)                 │
              │  + Name() string             │
              │  + Matches(text string) bool │
              │  + Execute(bot, msg) error   │
              └──────────────────────────────┘
                              △
                              │ implements
              ┌───────────────┼───────────────────┐
              │               │                    │
     ┌────────┴──────┐ ┌─────┴──────┐    ┌───────┴────────┐
     │ AgentCommand  │ │ChatCommand │    │FileCommand ... │
     │ (новый)       │ │(сущ.)      │    │(сущ.)          │
     │               │ │            │    │                │
     │ agent: string │ │            │    │                │
     └───────────────┘ └────────────┘    └────────────────┘
```

**Ключевые компоненты**:

1. **Интерфейс `Command`** — определяет контракт для любой команды:
   - `Name() string` — имя команды для логирования/help
   - `Matches(text string) bool` — проверяет, относится ли сообщение к этой команде
   - `Execute(ctx Context, msg *tgbotapi.Message)` — выполняет команду

2. **`CommandRouter`** — реестр команд, обходит список и вызывает первую подошедшую.

3. **`AgentCommand`** — конкретная реализация для opencode-агентов. Одна структура параметризуется именем агента и префиксом команды (DRY — не 7 отдельных типов, а 7 экземпляров одного типа).

4. **Существующие обработчики** (`handleChat`, `handleCode`, `handleFileCommand`, `handleMode`, `handleState`, `handleRetry`) также оборачиваются в `Command` и регистрируются в роутере.

### 4.3. Структура `AgentCommand`

```go
// Псевдокод для иллюстрации идеи
type AgentCommand struct {
    prefix    string   // например, "/go-senior "
    agentName string   // например, "go-senior"
    label     string   // например, "Go Senior Developer"
    agent     *opencode.Agent
    logger    *logger.Logger
}

func (c *AgentCommand) Matches(text string) bool {
    return strings.HasPrefix(text, c.prefix)
}

func (c *AgentCommand) Execute(bot *Bot, msg *tgbotapi.Message) {
    prompt := strings.TrimPrefix(msg.Text, c.prefix)
    if prompt == "" {
        bot.sendMessage(msg.Chat.ID, "Укажите задачу после команды. Пример: "+c.prefix+"<задача>")
        return
    }
    // ... отправка статуса, запуск горутины, вызов agent.ExecuteWithAgent(...)
}
```

### 4.4. Изменения в `Agent`

Текущий метод:
```go
func (a *Agent) Execute(ctx context.Context, prompt string) (string, error)
```

Новый метод:
```go
func (a *Agent) ExecuteWithAgent(ctx context.Context, prompt, agentName string) (string, error)
```

Старый `Execute` остаётся без изменений (для `/run`) и делегирует вызов `ExecuteWithAgent` с пустым именем агента.

Сборка аргументов командной строки:
```go
args := []string{"run", prompt, "--dangerously-skip-permissions"}
if agentName != "" {
    args = append(args, "--agent", agentName)
}
```

### 4.5. Диаграмма взаимодействия (текстовая)

```
Пользователь
  │
  │  /go-senior напиши HTTP-сервер на Go
  ▼
Telegram API ───► Bot.handleMessage()
                      │
                      ▼
                CommandRouter.Dispatch(msg)
                      │
                      │ итерация по commands
                      ▼
                AgentCommand.Matches("/go-senior ...") → true
                      │
                      ▼
                AgentCommand.Execute()
                      │
                      ├─► sendMessage("⏳ Запуск агента Go Senior Developer...")
                      │
                      └─► go func() {
                            ctx, cancel := context.WithTimeout(...)
                            result, err := agent.ExecuteWithAgent(ctx, prompt, "go-senior")
                            if err → sendMessage("Ошибка...")
                            else   → sendLongMessage(result)
                          }
```

### 4.6. Порядок регистрации команд

Команды с префиксами должны проверяться до общей команды `/run`. Поэтому порядок регистрации в роутере важен:

1. `/start`, `/help` (точное совпадение)
2. `/clear`, `/state`, `/retry` (точное совпадение)
3. `/mode` (с параметром)
4. Все новые agent-команды: `/explore`, `/go-review`, `/go-senior`, `/react-dev`, `/react-review`, `/write-tests`, `/tz`
5. `/run` (старая, без агента — проверяется **после** новых команд, так как `/run ` не конфликтует с ними по префиксу)
6. `/code`, `/review`, `/explain`, `/test`, `/docs` (существующие файловые команды)
7. `/chat` (с параметром)
8. Fallback: обычный чат (произвольный текст)

## 5. Модели данных

### 5.1. Новых таблиц/структур не требуется

Все данные остаются in-memory в `MemoryStorage`. Контекст диалогов (`Conversation`) не затрагивается, так как opencode-агенты не используют историю диалогов Telegram-бота — они работают в рамках одного запроса.

### 5.2. Структура `Config` — изменения

Добавить одно новое поле (опционально):

```go
type Config struct {
    // ... существующие поля ...

    OpenCodeDefaultAgent string  // NEW: имя агента по умолчанию для /run (пусто = default opencode agent)
}
```

**Переменная окружения**: `OPENCODE_DEFAULT_AGENT` (необязательная, по умолчанию пустая строка).

**Обоснование**: позволяет администратору бота переопределить агента по умолчанию для `/run` без изменения кода.

### 5.3. `.env.example` — изменения

Добавить строку:
```env
# OPENCODE_DEFAULT_AGENT=general  # Агент по умолчанию для /run (пусто = default opencode)
```

## 6. API / Интерфейсы

### 6.1. Интерфейс `Command` (новый файл `internal/telegram/commands.go`)

```go
type Command interface {
    // Name возвращает человекочитаемое имя команды (для логирования и /help)
    Name() string
    // Matches проверяет, относится ли текст сообщения к этой команде
    Matches(text string) bool
    // Execute выполняет команду. Получает Bot для доступа к sendMessage, логеру, хранилищу.
    Execute(bot *Bot, msg *tgbotapi.Message)
}
```

**Контекст выполнения** — сама структура `*Bot`, так как в ней уже есть все зависимости (api, cfg, ds, store, logger, agent). Выделять отдельный `Context` нецелесообразно — это усложнит код без выигрыша.

### 6.2. `CommandRouter`

```go
type CommandRouter struct {
    commands []Command
}

func NewCommandRouter() *CommandRouter
func (r *CommandRouter) Register(cmd Command)       // регистрирует команду (порядок важен!)
func (r *CommandRouter) Dispatch(bot *Bot, msg *tgbotapi.Message) bool  // возвращает false если ни одна команда не подошла
```

### 6.3. Метод `Agent.ExecuteWithAgent` (изменение `internal/opencode/agent.go`)

```go
// ExecuteWithAgent выполняет opencode с указанным агентом.
// Если agentName пустой, используется агент по умолчанию (без флага --agent).
func (a *Agent) ExecuteWithAgent(ctx context.Context, prompt, agentName string) (string, error)
```

Сигнатура существующего `Execute` **не меняется** для обратной совместимости. Внутри он делегирует `ExecuteWithAgent(ctx, prompt, "")`.

### 6.4. Telegram-команды и их параметры

| Команда | Формат | Агент opencode | Валидация |
|---|---|---|---|
| `/explore` | `/explore <задача>` | `explore` | prompt не пустой, ≤ 4096 символов |
| `/go-review` | `/go-review <файл или промпт>` | `go-reviewer` | prompt не пустой, ≤ 4096 |
| `/go-senior` | `/go-senior <задача>` | `go-senior` | prompt не пустой, ≤ 4096 |
| `/react-dev` | `/react-dev <задача>` | `react-developer` | prompt не пустой, ≤ 4096 |
| `/react-review` | `/react-review <файл или промпт>` | `react-reviewer` | prompt не пустой, ≤ 4096 |
| `/write-tests` | `/write-tests <файл или промпт>` | `test-writer` | prompt не пустой, ≤ 4096 |
| `/tz` | `/tz <задача>` | `tz-writer` | prompt не пустой, ≤ 4096 |
| `/run` | `/run <задача>` | default (из `OPENCODE_DEFAULT_AGENT`) | prompt не пустой, ≤ 4096 |

**Примечание по `/go-review` и `/react-review`**: Для ревью opencode-агентам **не нужно** передавать содержимое файла через промпт. Пользователь может:
1. Передать имя файла (агент прочитает его сам из workspace) — `/go-review ./internal/telegram/bot.go`
2. Передать фрагмент кода текстом — `/go-review func handleMessage...`
3. Передать промпт «проверь последний коммит» — агент сам выполнит `git diff`

Это принципиально отличается от существующих команд `/review`, `/explain`, `/test`, `/docs`, которые **читают файл на стороне бота** и отправляют содержимое в DeepSeek API.

## 7. Конфликт имён: `/test` vs `/write-tests`

### Проблема

Существующая команда `/test <файл>` читает локальный файл и отправляет его в DeepSeek API для генерации тестов. Новая команда для агента `test-writer` должна вызывать opencode.

Если назвать новую команду `/test`, возникнет конфликт: непонятно, какой обработчик должен сработать.

### Решение: `/write-tests`

Новая команда получает имя `/write-tests`. Обоснование:
- Семантически понятно: «напиши тесты»
- Не конфликтует с `/test` (существующая работает с файлами через DeepSeek)
- Аналоги `/gen-tests`, `/generate-tests` менее идиоматичны
- В будущем при рефакторинге можно объединить функциональность, но сейчас это разные механизмы (DeepSeek API vs OpenCode Agent)

### Альтернатива (не рекомендуется)

Можно было бы добавить подкоманду: `/test file <файл>` (DeepSeek) vs `/test agent <промпт>` (opencode). Но это:
- Ломает обратную совместимость для `/test`
- Усложняет пользовательский опыт
- Требует парсинга подкоманд

**Решение принято: `/write-tests`**.

## 8. Обработка ошибок

### 8.1. Сценарии ошибок и реакция

| Сценарий | Сообщение пользователю | Логирование |
|---|---|---|
| Пустой промпт (только команда) | «Укажите задачу после команды. Пример: /go-senior <задача>» | WARN |
| Промпт > 4096 символов | «Текст задачи слишком длинный (максимум 4096 символов)» | WARN |
| Таймаут выполнения | «⏰ Превышено время ожидания ответа от агента <имя>. Попробуйте упростить задачу.» | ERROR с elapsed |
| Ошибка запуска opencode (не найден бинарник) | «❌ OpenCode CLI не найден. Обратитесь к администратору.» | ERROR |
| Ошибка выполнения агента | «❌ Агент <имя> завершился с ошибкой: <первые 200 символов stderr>» | ERROR с полным stderr |
| context.Canceled (пользователь остановил бота) | Не отправляется (контекст отменён) | WARN |
| Пустой результат (успех, но нет вывода) | «✅ Агент <имя> выполнил задачу (ответ пуст)» | INFO |

### 8.2. Таймауты

- Таймаут задаётся через `OPENCODE_TIMEOUT` (по умолчанию 600 секунд)
- Для всех агентов используется одно значение
- При таймауте процесс `opencode` уничтожается через `context.Context`

### 8.3. Механизм повтора (`/retry`)

Команда `/retry` **не применяется** к opencode-агентам, так как они не сохраняют историю диалогов. `/retry` работает только с DeepSeek-чатом (как и сейчас).

## 9. План реализации

### Этап 1: Рефакторинг диспетчеризации (подготовительный)

**Файлы**:
- **Создать** `internal/telegram/commands.go` — интерфейс `Command`, `CommandRouter`, функции `Register`/`Dispatch`
- **Изменить** `internal/telegram/bot.go`:
  - Добавить поле `router *CommandRouter` в структуру `Bot`
  - Инициализировать роутер в `NewBot()`
  - Заменить тело `handleMessage()` на `b.router.Dispatch(b, msg)`
  - Оставить `sendHelp`, `sendMessage`, `sendLongMessage`, `splitMessage`, `createProxyHTTPClient` на месте

**Суть**: все существующие обработчики (`handleChat`, `handleCode`, `handleFileCommand`, `handleMode`, `handleState`, `handleRetry`, `handleRun`) оборачиваются в адаптеры `Command` и регистрируются в роутере. Поведение не меняется.

### Этап 2: Адаптация `Agent` для параметризованного вызова

**Файл**: `internal/opencode/agent.go`
- Добавить метод `ExecuteWithAgent(ctx context.Context, prompt, agentName string) (string, error)`
- Переписать `Execute` как делегат: `return a.ExecuteWithAgent(ctx, prompt, "")`
- Вынести сборку аргументов `opencode` в отдельный метод `buildArgs(prompt, agentName string) []string`

### Этап 3: Создание `AgentCommand` и регистрация новых команд

**Файлы**:
- **Создать** `internal/telegram/agent_commands.go`:
  - Структура `AgentCommand` с полями `prefix`, `agentName`, `displayName`
  - Конструктор `NewAgentCommand(prefix, agentName, displayName string)`
  - Методы `Name()`, `Matches()`, `Execute()`
- В `NewBot()` зарегистрировать 7 экземпляров `AgentCommand` в роутере (до регистрации `/run`)

### Этап 4: Обновление `/help`

**Файл**: `internal/telegram/bot.go` метод `sendHelp()`

Обновить текст справки, сгруппировав команды по категориям:

```
🤖 OpenCode агенты (выполнение задач):
/run <задача> — общий агент (по умолчанию)
/explore <задача> — исследование кодовой базы
/go-senior <задача> — написание Go-кода (Senior уровень)
/go-review <промпт> — ревью Go-кода
/react-dev <задача> — разработка React-компонентов
/react-review <промпт> — ревью React-кода
/write-tests <промпт> — генерация тестов
/tz <задача> — создание технического задания

💬 Чат с DeepSeek:
/chat <текст> — запрос к AI
/code <текст> — генерация кода (coder модель)
/mode <модель> — смена модели (chat/coder/reasoner)

📄 Работа с файлами:
/review <файл> — ревью кода
/explain <файл> — объяснение кода
/test <файл> — генерация тестов
/docs <файл> — документация

⚙️ Управление:
/state — статистика токенов
/retry — повторить последний запрос
/clear — очистить историю
/help — справка
```

### Этап 5: Валидация и финальные штрихи

**Файл**: `internal/telegram/agent_commands.go`
- Убедиться, что пустой промпт корректно обрабатывается
- Единообразие статусных сообщений: «⏳ Запуск агента <displayName>...»
- Логирование: `[AGENT] agent=<name> userID=<id> chatID=<id> elapsed=<duration> status=<success|error>`

### Этап 6: Конфигурация

**Файлы**:
- `internal/config/config.go` — добавить поле `OpenCodeDefaultAgent`
- `.env.example` — добавить строку с комментарием

### Этап 7: Тестирование

**Файл**: `internal/telegram/agent_commands_test.go`
- Unit-тесты для `Matches()` (все префиксы, включая `/run`, `/run_other_command`)
- Unit-тесты для `ExecuteWithAgent` (сборка аргументов)
- Unit-тесты для обработки пустого промпта
- Интеграционный тест роутера: регистрация команд, порядок срабатывания

## 10. Изменения в существующих файлах (сводка)

| Файл | Тип изменения | Описание |
|---|---|---|
| `internal/telegram/commands.go` | **Создать** | Интерфейс `Command`, структура `CommandRouter`, методы `Register`/`Dispatch` |
| `internal/telegram/agent_commands.go` | **Создать** | Структура `AgentCommand` и конструктор `NewAgentCommand` |
| `internal/telegram/agent_commands_test.go` | **Создать** | Тесты для `AgentCommand` |
| `internal/telegram/bot.go` | **Изменить** | Добавить `router` в `Bot`; заменить тело `handleMessage` на `b.router.Dispatch()`; обновить `sendHelp`; вынести регистрацию команд в `registerCommands()` |
| `internal/opencode/agent.go` | **Изменить** | Добавить `ExecuteWithAgent`; рефакторинг `Execute` в делегат; выделить `buildArgs` |
| `internal/config/config.go` | **Изменить** | Добавить поле `OpenCodeDefaultAgent` с чтением из `OPENCODE_DEFAULT_AGENT` |
| `.env.example` | **Изменить** | Добавить `OPENCODE_DEFAULT_AGENT` |
| `internal/telegram/handlers.go` | Без изменений | `handleChat` остаётся как есть, оборачивается в `Command` при регистрации |
| `internal/telegram/file_handlers.go` | Без изменений | `handleFileCommand` остаётся как есть, оборачивается в `Command` |
| `internal/telegram/mode_handlers.go` | Без изменений | `handleMode`, `handleCode` остаются как есть |
| `internal/telegram/state_handlers.go` | Без изменений | `handleState`, `handleRetry` остаются как есть |

## 11. Риски и компромиссы

### 11.1. Риски

| Риск | Вероятность | Влияние | Митигация |
|---|---|---|---|
| opencode меняет CLI-интерфейс (флаг `--agent`) | Низкая | Среднее | Зафиксировать формат аргументов в `buildArgs()`, легко изменить в одном месте |
| Большой вывод агента забивает очередь сообщений Telegram (rate limiting) | Средняя | Низкое | Уже есть `sendLongMessage` с разбивкой по 4096. Добавить задержку между частями (50-100ms) при необходимости |
| Длительное выполнение агента (10+ минут) приводит к накоплению горутин | Средняя | Среднее | Таймаут через `context.WithTimeout`, настраиваемый `OPENCODE_TIMEOUT`. Добавить метрику активных горутин |
| Пересечение workspace между параллельными вызовами агентов | Низкая | Высокое | opencode изолирует workspace через git worktree или отдельную директорию. Уточнить в документации opencode. Как минимум — убедиться, что `OPENCODE_WORKSPACE` уникален для каждого вызова |
| Команда `/tz` может конфликтовать с чем-то в будущем (короткое имя) | Низкая | Низкое | Переименовать при необходимости, сейчас конфликтов нет |

### 11.2. Компромиссы

1. **Не используется библиотека роутинга команд Telegram** (типа `telebot`). Обоснование: минимум внешних зависимостей, простой самописный роутер покрывает все нужды.

2. **Все agent-команды — экземпляры одной структуры `AgentCommand`**, а не отдельные типы. Это осознанное решение DRY. Если в будущем какая-то команда потребует уникального поведения — можно создать отдельный тип.

3. **Статусные сообщения однотипны** для всех агентов — не делаем уникальные эмодзи/тексты для каждого. При необходимости легко расширить через поле `statusMessage` в `AgentCommand`.

4. **`sendHelp` хардкодится**, а не генерируется из списка команд. Для 20+ команд это приемлемо; при росте до 50+ стоит задуматься о динамической генерации.

### 11.3. Альтернативное решение (отклонено)

**Использование одного обработчика с параметром**: `/agent <имя_агента> <задача>` вместо отдельных команд.

Отклонено, потому что:
- Хуже UX (пользователь должен помнить точные имена агентов)
- Сложнее валидация и подсказки
- Отдельные команды можно искать в истории чата
- Пользователь явно запросил отдельные команды

## 12. Метрики приёмки

- [ ] Все 7 новых команд доступны и вызывают правильного агента
- [ ] `/run` продолжает работать без изменений
- [ ] `/test` и `/write-tests` работают независимо
- [ ] `/help` показывает актуальный список команд
- [ ] Пустой промпт выдаёт подсказку
- [ ] Таймаут корректно обрывает выполнение
- [ ] Логи содержат имя агента
- [ ] Тесты проходят (cover новых файлов > 80%)
- [ ] Обратная совместимость: все старые команды работают
- [ ] Код проходит `go vet` и `go build`

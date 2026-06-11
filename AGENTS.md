# AGENTS.md

## Build & Run

```bash
go build -o bot ./cmd/bot/
./bot
# or
go run ./cmd/bot/
```

No lint/typecheck/test scripts exist in this repo.

## Entrypoint & Architecture

- **Entrypoint**: `cmd/bot/main.go` — loads config, wires dependencies, starts long-polling loop
- **Module path**: `telegram-deepseek-bot` (used in all internal imports)
- **Config**: `.env` file loaded via `godotenv`; see `.env.example` for all vars
- **Dependencies**: only two direct deps — `go-telegram-bot-api/v5` and `godotenv` (plus `golang.org/x/net` for SOCKS5)

## Critical: SOCKS5 Proxy for Telegram API

Telegram's API (`api.telegram.org`) is blocked from Russian infrastructure. **The SOCKS5 proxy must be passed at bot-creation time**, not after:

- `tgbotapi.NewBotAPI()` internally calls `getMe` using a default `http.Client` with no proxy.
- If proxy is set after `NewBotAPI()` returns, the initial `getMe` call already timed out.
- Correct: `tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, proxyHTTPClient)`

Two separate `createProxyHTTPClient()` implementations exist — one in `internal/telegram/bot.go` (used for the Telegram bot client) and another in `internal/deepseek/client.go` (used for DeepSeek API). They are near-duplicates.

## Mode-to-Model Mapping Gap

`/mode` stores short names (`chat`, `coder`, `reasoner`) via `storage.SetModel()`, but `handleChat()` uses the stored value directly as the DeepSeek model name. The actual model IDs are `deepseek-chat`, `deepseek-coder`, `deepseek-reasoner`. If a user changes mode, the api call will use the short name, which may fail.

## OpenCode Agent (`/run`)

- Requires `opencode` CLI on `$PATH`
- Runs as a subprocess: `opencode run <prompt> --dangerously-skip-permissions`
- Workspace dir is `OPENCODE_WORKSPACE` from config (defaults to `.`)
- Prompt limited to 4096 chars; execution timeout via `OPENCODE_TIMEOUT` (default 600s)

## Storage

- In-memory only (`storage.MemoryStorage`) — all state lost on restart
- Per-chat conversations keyed by `chatID` (int64)
- TTL cleanup (`CONVERSATION_TTL_HOURS`) is implemented (`Cleanup()`) but **never called** from the main loop

## .gitignore

- `.env` is gitignored (contains secrets) — `.env.example` is the template
- Compiled binary `bot` and `data/` are gitignored

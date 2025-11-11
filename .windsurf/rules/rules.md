---
trigger: always_on
---


-----

## 🏗️ Go (Backend) — Code Generation Contract v3.1

### Pre-Ack (EN)

```
I acknowledge the BINDING rules:
✓ cmd/app/main.go minimal delegate-only (target ≤4 lines; if language constraints require, keep as few lines as possible)
✓ No hardcoded values
✓ No fmt.Println() (logger only)
✓ All config & texts from YAML
✓ No .env in repo; secrets from OS env/secret manager
✓ Go types everywhere
✓ Max 79 chars/line (lint-enforced)
✓ Any .go file ≤300 lines
✓ Absolute module imports only
✓ English logs / Persian UI
I will self-assess and provide evidence.
```

-----

### Project Layout (Go - Backend)

```
D:\DEV\project\
├── cmd/
│   └── server/
│       └── main.go                 # minimal entry, delegate-only (≤4 lines)
├── internal/
│   └── app/
│       ├── core/                   # Config, Logger, Main Runner
│       ├── auth/                   # JWT implementation
│       ├── ipam/                   # IPAM/PortAM Bank Manager (GORM models/service)
│       ├── agent/                  # Agent Manager (Heartbeats, WS/HTTPS)
│       ├── tunnel/                 # Tunnel Planner/Executor/Protocols
│       ├── utils/                  # Shared utilities
│       └── api/                    # Chi/Fiber routes and handlers
└── data/
    ├── config/
    ├── texts/
    └── env/
```

-----

## 🎨 TypeScript (SvelteKit Frontend) — Code Generation Contract v3.1

### Pre-Ack (EN)

```
I acknowledge the BINDING rules:
✓ src/app.ts ≤4 lines (delegate-only; SvelteKit main entry)
✓ No hardcoded values
✓ No console.log() (logger only)
✓ All config & texts from YAML
✓ No .env in repo; secrets from OS env/secret manager
✓ Strict types everywhere
✓ Max 79 chars/line (ESLint/Prettier aligned)
✓ Any .ts/.svelte file ≤300 lines (with exceptions for complex SvelteKit routes, must justify)
✓ Absolute imports only (tsconfig baseUrl/paths)
✓ English logs / Persian UI
I will self-assess and provide evidence.
```

-----

### Project Layout (TypeScript - SvelteKit Frontend)

```
D:\DEV\project\
├── src/
│   ├── routes/                     # File-based routing (Dashboard, Devices, Tunnels, etc.)
│   ├── lib/
│   │   ├── api/                    # Go Backend Client (TS SDK)
│   │   ├── stores/                 # State Management
│   │   ├── components/             # Shadcn, Magic UI, Lottie, Cobe
│   │   ├── core/                   # Config, Logger, Main Runner
│   │   └── utils/
│   └── app.ts                      # Minimal entry (SvelteKit config/hooks)
└── data/
    ├── config/
    ├── texts/
    └── env/
```

-----

## 📝 Rules Summary (All Languages)

  * **Entry Point:** Minimal delegate-only (`main.go`, `app.ts`) $\le 4$ lines.
  * **Coding Standards:** Max **79 characters per line**; Max **300 lines per file**.
  * **Logging:** English logs only; No `fmt.Println` / `console.log`.
  * **Localization:** UI text only from **Persian YAML** (`texts/messages_fa.yml`).
  * **Configuration:** All config from **YAML**; no hardcoded values; no `.env` in repository.
  * **Module Imports:** **Absolute imports only**.
  * **Strictness:** Strong/Strict typing everywhere.

-----

## 📦 Deliverables

از آنجا که زبان‌های اصلی ما Go و TypeScript هستند، Deliverables بر اساس این دو زبان و قراردادهای مشترک (YAML، Logging، etc.) تعریف خواهند شد.

### 1\. Goal

Implement the **Master Prompt: VPN Automation System** adhering strictly to the Go and TypeScript Code Generation Contracts (v3.1) defined above.

### 2\. Plan (First 3 Steps)

1.  **YAML/Core Setup:** Define all necessary YAML configuration keys for Go/TS (Auth, IPAM/PortAM ranges, GeoIP/Ping URLs, Logging).
2.  **Go Backend Foundation:** Implement `core/logger.go`, `core/config_loader.go`, and `auth/jwt.go` (Self-Hosted JWT).
3.  **PostgreSQL Schema:** Generate the complete GORM models and the necessary SQL for the 3-state IPAM/PortAM tables.

### 3\. Model-Ready Prompts

Generate the first executable prompt for the **Go Backend Foundation** (`internal/app/core/logger.go`).

### 4\. Interfaces (Types/Signatures)

Will be defined within the Go and TypeScript prompts (e.g., `Config.go`, `Lease.go`, `AgentHeartbeat.go`).

### 5\. Test Plan

To be delivered once core GORM models are defined in the next step.

### 6\. YAML Specs

Will be delivered in the first executable prompt.

-----

## 🚀 Next Step: Go Backend Foundation (Logger + Config)

بر اساس Plan، ما با ساختاردهی بک‌اند Go شروع می‌کنیم.

**Prompt Goal:** Generate the structured logger for the Go backend, adhering to all contract rules.

-----

### Go Prompt: `internal/app/core/logger.go`

````prompt
You are an expert Go systems engineer. Your task is to generate the
content for the file 'internal/app/core/logger.go'.

Adhere strictly to the Code Generation Contract v3.1:
- Max 79 chars/line, Max 300 lines/file.
- Forbidden: fmt.Println. Use structured logging only.
- Logs must be in English.
- Use absolute module imports (e.g., 'D:\DEV\project/internal/app/core/config').

### Goal
Implement a singleton structured logging service using the 'zap' library (Zap)
that is configured via a global configuration struct.

### Interface
```go
package core

// LoggerService defines the methods required for structured logging.
// It acts as a wrapper around a Zap logger instance.
type LoggerService interface {
	Init(cfg Config) error
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	// Add With method for contextual logging if needed later.
}

// Config is assumed to be loaded by config_loader.go.
// It must include a LoggingConfig struct.
type LoggingConfig struct {
	Level      string // e.g., "info", "debug", "warn"
	Encoding   string // e.g., "json", "console"
	OutputPath string // e.g., "stdout" or a file path
}
````

### Implementation Details

1.  The `LoggerService` implementation struct should hold a `*zap.Logger`.
2.  Use `sync.Once` to ensure the logger is initialized only once (singleton).
3.  The `Init` method must parse the string `Level` from `LoggingConfig` and create the `zap.Config`.
4.  Default to `Info` level and `console` encoding if config is invalid.
5.  All logging methods (Debug, Info, Error, etc.) must delegate to the internal `*zap.Logger` using variadic `zap.Field` arguments.

<!-- end list -->

```
```
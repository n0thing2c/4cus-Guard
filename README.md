# 4cus-guard

> A focused productivity tool that blocks distracting websites and tracks deep-work sessions — powered by a Redis pub/sub microservices architecture.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](./docker-compose.yml)

---

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Configure Environment](#2-configure-environment)
  - [3. Start Infrastructure](#3-start-infrastructure)
  - [4. Build the Binaries](#4-build-the-binaries)
  - [5. Run the Services](#5-run-the-services)
- [CLI Usage](#cli-usage)
- [Configuration Reference](#configuration-reference)
- [Docker Deployment](#docker-deployment)
- [How It Works](#how-it-works)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Platform Notes](#platform-notes)
- [License](#license)

---

## Overview

**4cus-guard** helps you eliminate digital distractions during work sessions. It:

- **Blocks websites** by injecting entries into your system's hosts file (redirects domains to `127.0.0.1`)
- **Tracks focus sessions** with start/stop timestamps persisted in a local SQLite database
- **Flushes DNS cache** automatically after each hosts file change
- **Coordinates services** via Redis pub/sub — keeping components decoupled and independently deployable

All three components are cross-platform and run on **Windows**, **Linux**, and **macOS**.

---

## Prerequisites

| Tool                                               | Version | Notes                                                      |
| -------------------------------------------------- | ------- | ---------------------------------------------------------- |
| [Go](https://go.dev/dl/)                           | 1.24+   | Required to build binaries                                 |
| [Docker](https://www.docker.com/)                  | 24+     | Required to run Redis and the Timer service                |
| [Docker Compose](https://docs.docker.com/compose/) | v2+     | Bundled with Docker Desktop                                |
| Admin / root privileges                            | —       | Required by the Blocker service to write to the hosts file |

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/<your-username>/4cus-guard.git
cd 4cus-guard
```

### 2. Configure Environment

An `.env.example` is included. Copy it and fill in your values:

```bash
cp .env.example .env
```

Edit `.env`:

```dotenv
REDIS_ADDR=localhost:6380
REDIS_PASS=
DB_PATH=./data/focus.db
```

| Variable     | Default           | Description                          |
| ------------ | ----------------- | ------------------------------------ |
| `REDIS_ADDR` | `localhost:6380`  | Redis address (`host:port`)          |
| `REDIS_PASS` | _(empty)_         | Redis password — leave empty if none |
| `DB_PATH`    | `./data/focus.db` | SQLite database path                 |

> If a variable is omitted, the default is applied automatically. The `./data/` directory is created on first run.

### 3. Start Infrastructure

```bash
docker-compose up -d
```

Verify:

```bash
docker-compose ps
# redis-container   Up   0.0.0.0:6380->6379/tcp
```

### 4. Build the Binaries

**Option A — Makefile (recommended):**

```bash
make build-all
```

Runs the following in one shot:

```
go build -o focus ./cmd/cli
sudo cp focus /usr/local/bin/
GOOS=windows go build -o focus.exe ./cmd/cli
GOOS=windows go build -o blocker.exe ./cmd/blocker
```

**Option B — Manual:**

```bash
# CLI — build and install to PATH
go build -o focus ./cmd/cli && sudo cp focus /usr/local/bin/

# Blocker
go build -o blocker ./cmd/blocker

# Timer
go build -o timer ./cmd/timer
```

> **Windows target:** If you are on Linux/macOS and need `blocker.exe` to run on a Windows machine, cross-compile it:
>
> ```bash
> GOOS=windows GOARCH=amd64 go build -o blocker.exe ./cmd/blocker
> ```

> All binaries are statically linked (CGO disabled). No runtime dependencies.

### 5. Run the Services

Both services must be running **before** you use the CLI.

**Timer service** — already started by Docker Compose in step 3.

**Blocker service** (requires administrator privileges to write to the hosts file):

```bash
# Linux / macOS
sudo ./blocker

# Windows — open PowerShell as Administrator, then:
.\blocker.exe
```

Output on successful start:

```
Blocker is active
```

Both services run indefinitely. On `Ctrl+C`, the Timer service automatically marks any active focus session as `finished` before exiting.

---

## CLI Usage

```bash
focus start                  # Start a focus session
focus stop                   # End the current focus session
focus block facebook.com     # Block a website
focus unblock facebook.com   # Unblock a website
```

**URL sanitization** — the following inputs all resolve to the same result (`facebook.com`):

```
facebook.com
www.facebook.com
http://facebook.com
https://www.facebook.com/feed/
```

**Resulting hosts file entries:**

```
127.0.0.1 facebook.com     #4cus-guard
::1       facebook.com     #4cus-guard
127.0.0.1 www.facebook.com #4cus-guard
::1       www.facebook.com #4cus-guard
```

All entries are tagged `#4cus-guard` so `unblock` removes them surgically without touching anything else.

---

## Configuration Reference

**Environment variables:**

| Variable     | Default           | Source                      |
| ------------ | ----------------- | --------------------------- |
| `REDIS_ADDR` | `localhost:6380`  | `internal/config/config.go` |
| `REDIS_PASS` | `""`              | `internal/config/config.go` |
| `DB_PATH`    | `./data/focus.db` | `internal/config/config.go` |

**SQLite schema** (auto-created on startup):

```sql
CREATE TABLE IF NOT EXISTS focus_sessions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    start_time INTEGER NOT NULL,     -- Unix timestamp
    end_time   INTEGER DEFAULT 0,    -- Unix timestamp; 0 = session still active
    status     TEXT    NOT NULL      -- 'active' | 'finished'
);
```

**Redis message schema:**

```json
{
  "action": "start | stop | block | unblock",
  "timestamp": 1709123456,
  "url": "facebook.com"
}
```

---

## Docker Deployment

```bash
# Start Redis + Timer
docker-compose up -d


# Stop everything
docker-compose down
```

| Service                | Image                             | Port        | Notes                 |
| ---------------------- | --------------------------------- | ----------- | --------------------- |
| `redis-message-broker` | `redis:alpine`                    | `6380:6379` | Message broker        |
| `timer`                | Built from `cmd/timer/dockerfile` | —           | Focus session tracker |

The Timer container mounts `./data` as a volume — the SQLite database persists across restarts.

> The Blocker service is **not** containerized by design: it must write to the host OS's hosts file directly and cannot do so from inside a container.

---

## How It Works

```
focus start
  └─► {"action":"start"} ──────────────► channel: "Timer"
        └─► Timer: INSERT INTO focus_sessions (status='active')

focus stop
  └─► {"action":"stop"} ───────────────► channel: "Timer"
        └─► Timer: UPDATE focus_sessions SET status='finished'

focus block facebook.com
  └─► {"action":"block","url":"..."} ──► channel: "Blocker"
        └─► Blocker: append to hosts file → flush DNS

focus unblock facebook.com
  └─► {"action":"unblock","url":"..."} ► channel: "Blocker"
        └─► Blocker: remove tagged lines → flush DNS
```

**DNS flush strategy** (auto-detected at runtime via `runtime.GOOS`):

| OS      | Method                                                                                                                                 |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| Windows | `ipconfig /flushdns`                                                                                                                   |
| Linux   | `resolvectl flush-caches` → `systemd-resolve --flush-caches` → `nscd -i hosts` → `dnsmasq --clear-on-reload` (first one that succeeds) |
| Other   | No-op                                                                                                                                  |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        User                             │
│                    focus <command>                      │
└──────────────┬──────────────────────┬───────────────────┘
   start/stop  │                      │  block/unblock
               ▼                      ▼
     ┌──────────────────┐   ┌───────────────────┐
     │  channel: Timer  │   │ channel: Blocker  │
     └────────┬─────────┘   └────────┬──────────┘
              ▼                      ▼
   ┌─────────────────┐    ┌─────────────────┐
   │  Timer Service  │    │ Blocker Service │
   │  (cmd/timer)    │    │  (cmd/blocker)  │
   │                 │    │                 │
   │ Persists session│    │ Modifies system │
   │ to SQLite DB    │    │ hosts file +    │
   │ (focus.db)      │    │ flushes DNS     │
   └─────────────────┘    └─────────────────┘
```

| Component     | Role                               | Channel                                                      |
| ------------- | ---------------------------------- | ------------------------------------------------------------ |
| `cmd/cli`     | CLI publisher — user entry point   | `start`/`stop` → `"Timer"` · `block`/`unblock` → `"Blocker"` |
| `cmd/blocker` | Subscriber — modifies hosts file   | Subscribes to `"Blocker"`                                    |
| `cmd/timer`   | Subscriber — tracks focus sessions | Subscribes to `"Timer"`                                      |

---

## Project Structure

```
4cus-guard/
├── cmd/
│   ├── cli/                  # CLI entry point (Cobra commands)
│   │   ├── main.go
│   │   ├── rootCmd.go
│   │   ├── startCmd.go       # Publishes to "Timer"
│   │   ├── stopCmd.go        # Publishes to "Timer"
│   │   ├── blockCmd.go       # Publishes to "Blocker"
│   │   └── unblockCmd.go     # Publishes to "Blocker"
│   ├── blocker/              # Blocker subscriber service
│   │   └── main.go
│   └── timer/                # Timer subscriber service
│       ├── main.go
│       └── dockerfile
├── internal/
│   ├── config/               # Env loader with defaults
│   ├── db/                   # SQLite init (database/sql + WAL mode)
│   ├── message/              # Shared Message struct (JSON schema)
│   ├── pubsub/               # Publisher/Subscriber interfaces + Redis adapter
│   └── services/
│       ├── hostFileService.go  # Hosts file read/write/parse
│       └── networkService.go   # URL sanitization + cross-platform DNS flush
├── data/                     # Runtime SQLite database (git-ignored)
├── .env.example
├── .env                      # Local overrides (git-ignored)
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── LICENSE
```

---

## Platform Notes

| OS                | CLI                      | Blocker                       | Timer              |
| ----------------- | ------------------------ | ----------------------------- | ------------------ |
| **Linux / macOS** | `focus` (in PATH)        | `sudo ./blocker`              | Docker             |
| **Windows**       | `focus.exe` (PowerShell) | `blocker.exe` (Administrator) | Docker             |
| **WSL2**          | `focus` inside WSL2      | `blocker.exe` on Windows side | Docker inside WSL2 |

> **WSL2 tip:** Redis runs in Docker on the WSL2 side. Both `blocker.exe` (Windows) and the WSL2 services reach it via `localhost:6380`.

---

## License

Distributed under the [MIT License](./LICENSE).

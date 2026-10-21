# ⚡ GoPower — Smart Grid Energy Dispatching & Telemetry Microservice

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20Layered-orange)](#system-architecture)

**GoPower** is a production-grade, highly-concurrent microservice written in **Go (Golang)** for smart energy grid distribution, real-time load dispatching, and live telemetry streaming. 

Built to simulate modern renewable microgrids (Solar, Wind, Battery Storage), GoPower balances energy generation with consumer demand, executes atomic dispatch transactions, and broadcasts critical telemetry and blackout alerts over real-time WebSockets without blocking the main event loops.

---

## 🏛 System Architecture

The project adheres to **Clean / Layered Architecture** principles, enforcing strict separation of concerns and dependency inversion:

```mermaid
flowchart TD
    Client["Client / IoT Dispatcher Dashboard"]

    subgraph Transport ["1. Transport Layer"]
        REST["REST API (Gin Engine)\n/api/v1/*"]
        WS["WebSocket Hub (Gorilla)\n/ws"]
    end

    subgraph Service ["2. Business Logic Layer"]
        GridService["GridService\n- Energy Dispatching\n- Generation Recording\n- Net Balance & Thresholds"]
    end

    subgraph Persistence ["3. Persistence Layer"]
        Repo["Station & Consumer Repositories\n(Interface Abstraction)"]
        DB[("SQLite Database\n(GORM with ACID Transactions)")]
    end

    subgraph Domain ["4. Domain Layer"]
        Models["Domain Entities\n- PowerStation\n- ConsumerMeter\n- DispatchRecord\n- GridAlert & Telemetry"]
    end

    Client -->|HTTP REST| REST
    Client <-->|Bi-directional WebSocket| WS
    REST --> GridService
    WS -.->|Broadcast Alerts & Telemetry| Client
    GridService --> Repo
    Repo --> DB
    GridService -.-> Models
    Repo -.-> Models
```

---

## 🚀 Key Engineering Highlights

- **High-Concurrency Real-Time Hub:** Implemented an actor-style event loop using Go channels (`register`, `unregister`, `broadcast`) and `sync.RWMutex`. Employs non-blocking broadcasts to automatically drop slow or stalled clients without impacting overall grid throughput.
- **WebSocket Heartbeat (Ping/Pong):** Periodic ping/pong health-checks detect dead connections and clean up TCP socket descriptors.
- **ACID Atomic Transactions:** Energy dispatch transfers utilize database transactions (`tx.Commit` / `tx.Rollback`) ensuring station supply deductions and consumer demand additions remain strictly consistent.
- **Production-Ready Observability:** Structured JSON logging powered by Go's modern `log/slog` standard library.
- **Zero-Downtime Graceful Shutdown:** Traps OS signals (`SIGINT`, `SIGTERM`) to cleanly finish in-flight HTTP requests and gracefully close WebSocket clients.
- **Containerized & Minimalist:** Multi-stage `Dockerfile` producing a secure, non-root Alpine container image under **25MB**.
- **Automated CI Pipeline:** GitHub Actions workflow running `go vet` and concurrency tests with Go's **Race Detector (`-race`)**.

---

## 🛠 Tech Stack

- **Language:** Go (1.24+)
- **HTTP Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **WebSocket Protocol:** [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **ORM & Database:** [GORM](https://gorm.io/) with SQLite driver
- **Testing:** Standard Go testing with built-in `-race` detector
- **Containerization:** Docker & Docker Compose
- **CI/CD:** GitHub Actions

---

## 🚦 Getting Started

### Prerequisites

- Go 1.22 or higher installed.
- (Optional) Docker and Docker Compose installed.

### Option 1: Running Locally

```bash
# Clone the repository
git clone https://github.com/milad-ghaderi-dev/GoPower.git
cd GoPower

# Download dependencies
go mod download

# Run the server
go run ./cmd/server
```

The server will start listening on port `:8080`.

### Option 2: Running with Docker Compose

```bash
docker compose up --build
```

---

## 📡 API Reference

### Health Check
- `GET /health` — Verifies service readiness (Returns `200 OK`).

### Stations Management
- `POST /api/v1/stations` — Register a new power generation station.
- `GET /api/v1/stations` — List all registered stations (supports `?type=SOLAR|WIND|BATTERY`).
- `GET /api/v1/stations/:id` — Retrieve details for a specific station.

### Consumer Meters
- `POST /api/v1/consumers` — Register a new power consumer meter.
- `GET /api/v1/consumers` — List all consumers.
- `GET /api/v1/consumers/:id` — Retrieve consumer details.

### Energy Grid Operations
- `POST /api/v1/generation` — Record output energy generation for an active station.
- `POST /api/v1/dispatch` — Atomically dispatch energy from a station to a consumer.
- `GET /api/v1/grid/summary` — Get real-time net grid balance and system status (`NORMAL`, `WARNING`, `CRITICAL`).

### Real-Time WebSocket Endpoint
- `GET /ws` — Upgrades HTTP connection to bi-directional WebSocket stream.

#### Sample Telemetry Payload
```json
{
  "type": "TELEMETRY",
  "summary": {
    "total_stations": 4,
    "total_generation_mw": 320.5,
    "total_capacity_mw": 450.0,
    "total_consumers": 12,
    "total_demand_mw": 290.0,
    "net_balance_mw": 30.5,
    "status": "NORMAL"
  },
  "timestamp": "2026-10-21T12:00:00Z"
}
```

---

## 🧪 Testing & Verification

Run the test suite including race condition detection:

```bash
go test -v -race ./...
```

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
# Denethor - A remote CLI tool 

An online CLI platform leveraging a Go authoritative server, a Go web client with JS script and Valkey for caching.

## Architecture

This project utilizes a decoupled microservices-oriented approach:

* **Server** (`/server`): A Go-based authoritative service. It manages the PostgreSQL connection, validates app state, and processes high-concurrency logic via Goroutines.
* **Client** (`/client`): A Go-based application combining frontend source, public assets and JS scripts. Handles user input and provides a low-latency UI via client-side prediction.
* **Cache** (`/cache`): Valkey for session management and high-speed state caching. **not implemented**
* **Database** (PostgreSQL): Provides persistent storage for user profiles, inventory, and historical game state.

```
.
├── frontend/
│   ├── public/
│   │   ├── index.js
│   │   └── index.html
│   └── frontend.go
├── backend/
│   ├── backend.go
│   ├── console.go
│   └── internal/
└── cache/
    └── valkey.conf
```

## Infrastructure

The entire stack is containerized for consistent development and deployment:

* **Isolation**: Every component (Client, Server, DB, Valkey) runs in its own Docker container.
* **Persistence**: A managed Docker volume ensures PostgreSQL data persists across container lifecycles.
* **Caching**: Valkey container provides in-memory caching for performance-critical operations.
* **Orchestration**: Docker Compose manages the internal networking and dependency startup order.

## Getting Started

### Prerequisites
* Docker Desktop
* Go 1.21+ (for local server development)
* Node.js 20+ (for client development)

### Quick Start
1. Clone the repository
```bash
git clone https://github.com/hortarion/denethor.git
cd denethor
```

2. Launch the stack
```bash
docker-compose up --build
```

3. Service Access
   - App Client: http://localhost:3000
   - RabbitMQ Admin: http://localhost:15672 (User/Pass: guest/guest)
   - Valkey CLI: `docker exec -it valkey valkey-cli`

## Communication Flow

1. **Ingress**: Client sends user command via WebSocket.
2. **Processing**: The Go Server consumes the action, validates it and writes results to PostgreSQL.
3. **Caching**: Valkey stores frequently accessed state for low-latency retrieval.
4. **Egress**: The Server publishes the "Authoritative State" back to a global exchange.
5. **Sync**: All Clients receive the update and reconcile their local state to match the Server.

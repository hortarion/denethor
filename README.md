# Distributed App Engine: Pub/Sub Architecture

An online app platform leveraging a Go authoritative server, a Go web client with JS script, RabbitMQ for distributed messaging, and Valkey for caching.

## Architecture

This project utilizes a decoupled microservices-oriented approach:

* **Server** (`/server`): A Go-based authoritative service. It manages the PostgreSQL connection, validates game state, and processes high-concurrency logic via Goroutines.
* **Client** (`/client`): A Go-based application combining frontend source, public assets and JS scripts. Handles user input and provides a low-latency UI via client-side prediction.
* **Cache** (`/cache`): Valkey for session management and high-speed state caching.
* **Message Broker** (RabbitMQ): Orchestrates asynchronous communication between peers and the server using the Pub/Sub pattern.
* **Database** (PostgreSQL): Provides persistent storage for user profiles, inventory, and historical game state.
* **Internal** (`/internal`): Shared definitions and schemas (JSON/Protobuf) used to synchronize data structures between Go and TypeScript.

```
.
├── client/
│   ├── src/
│   │   ├── index.ts
│   │   ├── console.ts
│   │   └── api/
│   │       ├── input.ts
│   │       ├── metrics.ts
│   │       └── middleware.ts
│   ├── public/
│   │   └── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── scripts/
│       └── copyIndex.sh
├── server/
│   ├── server.go
│   ├── go.mod
│   └── internal/
└── cache/
    └── valkey.conf

```

## Infrastructure

The entire stack is containerized for consistent development and deployment:

* **Isolation**: Every component (Client, Server, DB, RabbitMQ, Valkey) runs in its own Docker container.
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
git clone https://github.com/your-repo/distributed-game.git
cd distributed-game
```

2. Launch the stack
```bash
docker-compose up --build
```

3. Service Access
   - Game Client: http://localhost:3000
   - RabbitMQ Admin: http://localhost:15672 (User/Pass: guest/guest)
   - Valkey CLI: `docker exec -it valkey valkey-cli`

## Communication Flow

1. **Ingress**: Client publishes "Player Action" to a RabbitMQ exchange.
2. **Processing**: The Go Server consumes the action, validates it against game rules, and writes results to PostgreSQL.
3. **Caching**: Valkey stores frequently accessed state for low-latency retrieval.
4. **Egress**: The Server publishes the "Authoritative State" back to a global exchange.
5. **Sync**: All Clients receive the update and reconcile their local state to match the Server.

## Development

- Schema Changes: Update the shared definitions in `/internal` first to ensure Go and TypeScript structures match.
- Database: Migrations are handled automatically by the Server on startup.
- Caching: Configure Valkey connection strings and TTLs in the server configuration.
- Scaling: Use `docker-compose up --scale client=3` to simulate multiple players locally.

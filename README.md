# Distributed Game Engine: Pub/Sub Architecture

A high-performance, real-time multiplayer environment leveraging a Go authoritative server, a TypeScript web client, and RabbitMQ for distributed messaging.

## Architecture

This project utilizes a decoupled microservices-oriented approach:

* Server (/server): A Go-based authoritative service. It manages the PostgreSQL connection, validates game state, and processes high-concurrency logic via Goroutines.
* Client (/client): A TypeScript/Vite application that handles user input and provides a low-latency UI via client-side prediction.
* Message Broker (RabbitMQ): Orchestrates asynchronous communication between peers and the server using the Pub/Sub pattern.
* Database (PostgreSQL): Provides persistent storage for user profiles, inventory, and historical game state.
* Internal (/internal): Shared definitions and schemas (JSON/Protobuf) used to synchronize data structures between Go and TypeScript.

```
.
├── client/
│   ├── tsconfig.json
│   ├── package.json
│   └── src/
├── server/
│   ├── main.go
│   └── go.mod
└── internal/
    ├── logger.go
    └── go.mod
```

## Infrastructure

The entire stack is containerized for consistent development and deployment:

* Isolation: Every component (Client, Server, DB, RabbitMQ) runs in its own Docker container.
* Persistence: A managed Docker volume ensures PostgreSQL data persists across container lifecycles.
* Orchestration: Docker Compose manages the internal networking and dependency startup order.

## Getting Started

### Prerequisites
* Docker Desktop
* Go 1.21+ (for local server development)
* Node.js 20+ (for client development)

### Quick Start
1. Clone the repository
```
  git clone https://github.com/your-repo/distributed-game.git
   cd distributed-game
```

2. Launch the stack
```
  docker-compose up --build
```

3. Service Access
   - Game Client: http://localhost:3000
   - RabbitMQ Admin: http://localhost:15672 (User/Pass: guest/guest)

## Communication Flow

1. Ingress: Client publishes "Player Action" to a RabbitMQ exchange.
2. Processing: The Go Server consumes the action, validates it against game rules, and writes results to PostgreSQL.
3. Egress: The Server publishes the "Authoritative State" back to a global exchange.
4. Sync: All Clients receive the update and reconcile their local state to match the Server.

## Development

- Schema Changes: Update the shared definitions in /internal first to ensure Go and TypeScript structures match.
- Database: Migrations are handled automatically by the Server on startup.
- Scaling: Use 'docker-compose up --scale client=3' to simulate multiple players locally.

# Denethor - A remote CLI tool 

## About this project
Denethor is a show case project based on a decoupled architecture with server backend and a client frontend.

### Disclaimer
This program is intended as a training and showcase based on boot.dev's back end developer course.
The application is not supposed to be a deployable tool for real life usage. I will start a new single-application project for this.

### Features
- Client:
  - Hosting a webpage with a simple CLI mockup
- Server:
  - Establishes a webSocket connection with the client
  - Validates input
  - Stores user credentials on postgreSQL database (hashed passwords)
  - Launches build-in apps (show case only provides a test app)

### Available commands
- app - App launcher
- clear - Clear the screen
- help - Display available commands
- login - Login to existing user account
- logout - Logout from user account
- ping - Ping the server
- register - Register a new user account
- shout - Broadcast to all clients (_run a second page instance in a private window to test_)

## Architecture

This project utilizes a decoupled microservices-oriented approach:

* **Server** (`/server`): A Go-based authoritative service. It manages the PostgreSQL connection, validates app state, and processes high-concurrency logic via Goroutines.
* **Client** (`/client`): A Go-based application combining frontend source, public assets and JS scripts. Handles user input and provides a low-latency UI via client-side prediction.
* **Cache** (`/cache`): Valkey for session management and high-speed state caching. **not implemented**
* **Database** (PostgreSQL): Provides persistent storage for user profiles, inventory, and historical game state.

```
.
├── client/
│   ├── public/
│   │   ├── favicon.ico
│   │   ├── index.hmtl
│   │   └── index.js
│   └── frontend.go
├── server/
│   ├── backend.go
│   ├── (...)
│   └── internal/
│       ├── apps/
│       ├── auth/
│       ├── database/
│       └── sql/
│           ├── queries/
│           └── schema/
└── cache/ //not implemented
    └── valkey.conf
```

### Prerequisites
* Docker && docker-compose
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
cd denethor/server
docker-compose up --build

cd denethor/client
docker-compose up --build
```

3. Service Access
   - App Client: http://localhost:8090

## Communication Flow

1. Client sends user command via WebSocket.
2. The Go Server consumes the action, validates it and writes user data to PostgreSQL.
3. **Caching**: Valkey not implemented.
4. The Server publishes the "Authoritative State" back to client.
5. All Clients receive the update and reconcile their local state to match the Server.

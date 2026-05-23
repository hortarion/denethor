package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hortarion/server/internal/database"

	"github.com/joho/godotenv"
)

type serverConfig struct {
	Port           string
	DB             *database.Queries
	Platform       string
	AllowedOrigins []string
	Upgrader       websocket.Upgrader
}

type websocketMessage struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
	Data    string `json:"data"`
}

func main() {

	godotenv.Load()

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("PORT must be a valid number")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		log.Fatal("ALLOWED_ORIGINS must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	cfg := serverConfig{
		Port:           port,
		DB:             dbQueries,
		Platform:       platform,
		AllowedOrigins: []string{"http://localhost:8090"},
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return slices.Contains(cfg.AllowedOrigins, origin)
		},
	}

	cfg.Upgrader = upgrader

	mux := http.NewServeMux()

	// Might replace with brocker logic
	mux.HandleFunc("/ws", cfg.handleConnection)
	mux.HandleFunc("/status", handleStatusPage)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	srv := http.Server{
		Handler:      mux,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// this blocks forever, until the server
	// has an unrecoverable error
	fmt.Printf("server started on http://localhost:%s\n", cfg.Port)
	serverErr := srv.ListenAndServe()
	log.Fatal(serverErr)

}

// Client map for broadcasting
var (
	clients   = make(map[string]chan []byte)
	clientsMu sync.Mutex
)

func (cfg *serverConfig) handleConnection(w http.ResponseWriter, r *http.Request) {
	// Create context and establish WS connection
	ctx := r.Context()

	conn, err := cfg.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	connID := uuid.New().String()
	log.Printf("New WebSocket connection: %s from %s", connID, conn.RemoteAddr())
	defer func() {
		clientsMu.Lock()
		delete(clients, connID)
		clientsMu.Unlock()
		conn.Close()
	}()

	// Concurrent sending of outbound messages
	outbound := make(chan []byte, 10)

	// Add connection to client map
	clientsMu.Lock()
	clients[connID] = outbound
	clientsMu.Unlock()

	go func() {
		for msg := range outbound {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[%s] Write error: %v", connID, err)
				break
			}
		}
	}()

	// Authentication channel to pass login creds to auth package
	authChan := make(chan string, 1)
	ctx = context.WithValue(ctx, "authChan", authChan)

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%s] read error: %v", connID, err)
			break
		}

		params := websocketMessage{}
		err = json.Unmarshal(message, &params)
		if err != nil {
			log.Printf("[%s] Failed to unmarshal JSON: %v", connID, err)
			continue
		}
		var response websocketMessage
		switch params.Channel {
		case "sys":
			log.Printf("System received: %s", message)
		case "console":
			response, err = cfg.handleConsole(ctx, conn, params.Data, outbound)
			if err != nil {
				log.Printf("[%s] Console: %v", connID, err)
				response = websocketMessage{
					Channel: "console",
					Token:   "error",
					Data:    err.Error(),
				}
			}
		case "auth":
			select {
			case authChan <- params.Data:
				// Success
			default:
				log.Printf("[%s] auth channel full", connID)
				response = websocketMessage{
					Channel: "auth",
					Token:   "error",
					Data:    "auth channel full",
				}
			}
			if response.Channel == "" {
				response.Channel = "auth"
				response.Data = ""
			}
		default:
			response = websocketMessage{}
		}
		byteResponse, err := json.Marshal(response)
		if err != nil {
			log.Printf("[%s] failed to marshal JSON: %s", connID, err)
			continue
		}
		outbound <- byteResponse
	}
	log.Printf("Connection %s closed", connID)
}

type cliCommand struct {
	name        string
	description string
	callback    func(
		ctx context.Context,
		authChan chan string,
		outbound chan<- []byte,
		args []string,
	) (websocketMessage, error)
}

// Console command registry
func (cfg *serverConfig) getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"clear": {
			name:        "clear",
			description: "Clear the screen",
			callback:    cfg.handleClear,
		},
		"help": {
			name:        "help",
			description: "Display available commands",
			callback:    cfg.handleHelp,
		},
		"login": {
			name:        "login",
			description: "Login to existing user account",
			callback:    cfg.handleLogin,
		},
		"ping": {
			name:        "ping",
			description: "Ping the server",
			callback:    cfg.handlePing,
		},
		"register": {
			name:        "register",
			description: "Register a new user account",
			callback:    cfg.handleRegister,
		},
		"shout": {
			name:        "shout",
			description: "Broadcast to all clients",
			callback:    cfg.handleShout,
		},
	}
}

func (cfg *serverConfig) handleConsole(ctx context.Context, _ *websocket.Conn, message string, outbound chan<- []byte) (websocketMessage, error) {
	authChan, ok := ctx.Value("authChan").(chan string)
	if !ok {
		return websocketMessage{}, fmt.Errorf("auth channel not found")
	}
	cmd := strings.ToLower(strings.Split(message, " ")[0])
	args := strings.Split(message, " ")[1:]
	log.Println("[DEV] cmd:", cmd)
	for idx, arg := range args {
		log.Println("[DEV]", idx+1, ":", arg)
	}

	response := websocketMessage{
		Channel: "console",
	}

	command, exists := cfg.getCommands()[cmd]
	if exists {
		return command.callback(ctx, authChan, outbound, args)
	} else {
		return response, nil
	}
}

func (cfg *serverConfig) handleShout(ctx context.Context, authChan chan string, outbound chan<- []byte, args []string) (websocketMessage, error) {
	message := websocketMessage{
		Channel: "console",
		Token:   "",
		Data:    "Someone shouts very loud",
	}
	cfg.broadcast(message)
	return websocketMessage{
		Channel: "console",
		Data:    "It was you!",
	}, nil
}

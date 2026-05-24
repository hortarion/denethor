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
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hortarion/server/internal/auth"
	"github.com/hortarion/server/internal/database"

	"github.com/joho/godotenv"
)

type serverConfig struct {
	Port           string
	DB             *database.Queries
	Platform       string
	AllowedOrigins []string
	Upgrader       websocket.Upgrader
	Clients        map[string]*Client
	ClientsMu      sync.Mutex
	JWTSecret      string
}

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Outbound chan []byte
	AuthChan chan string
	IsAuthed bool
	closed   bool
}

type websocketMessage struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
	Data    string `json:"data"`
}

func loadEnv() serverConfig {
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(dbConn)

	return serverConfig{
		Port:           port,
		DB:             dbQueries,
		Platform:       platform,
		AllowedOrigins: []string{"http://localhost:8090"},
		Clients:        make(map[string]*Client),
		ClientsMu:      sync.Mutex{},
		JWTSecret:      jwtSecret,
	}
}

func main() {

	cfg := loadEnv()

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

func (cfg *serverConfig) closeConnection(wg *sync.WaitGroup, client Client, conn *websocket.Conn, connID string) {
	cfg.ClientsMu.Lock()
	if existingClient := cfg.Clients[connID]; existingClient != nil {
		existingClient.closed = true
	}
	delete(cfg.Clients, client.ID)
	cfg.ClientsMu.Unlock()

	// Close channel safely
	cfg.ClientsMu.Lock()
	if client := cfg.Clients[client.ID]; client != nil && !client.closed {
		select {
		case <-client.Outbound:
		// Already closed
		default:
			close(client.Outbound)
		}
	}
	cfg.ClientsMu.Unlock()

	wg.Wait()

	cfg.ClientsMu.Lock()
	delete(cfg.Clients, connID)
	cfg.ClientsMu.Unlock()
	conn.Close()
}

func (cfg *serverConfig) handleConnection(w http.ResponseWriter, r *http.Request) {
	// Create ID and associate with context
	connID := uuid.New().String()
	ctx := context.WithValue(r.Context(), "connID", connID)

	// Establish WS connection
	conn, err := cfg.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Set deadlines
	conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

	// Outbound channel
	outbound := make(chan []byte)

	client := &Client{
		ID:       connID,
		Conn:     conn,
		Outbound: outbound,
		IsAuthed: false,
	}

	// Add connection to client map
	cfg.ClientsMu.Lock()
	cfg.Clients[connID] = client
	cfg.ClientsMu.Unlock()
	log.Printf("[SYS] New WebSocket connection: %s from %s", connID, conn.RemoteAddr())

	// Always create fresh AuthChan
	authChan := make(chan string, 1)
	client.AuthChan = authChan
	ctx = context.WithValue(ctx, "authChan", authChan)

	var wg sync.WaitGroup
	wg.Add(1)

	go cfg.writeMessages(&wg, client, conn)

	// Cleanup connection on close
	defer cfg.closeConnection(&wg, *client, conn, connID)

	cfg.handleMessages(ctx, conn, client)

	log.Printf("Connection %s closed", connID)
}

func (cfg *serverConfig) writeMessages(wg *sync.WaitGroup, client *Client, conn *websocket.Conn) {
	defer wg.Done()
	for msg := range client.Outbound {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("[SYS] %s Write error: %v", client.ID, err)
			break
		}
	}
}

func (cfg *serverConfig) handleMessages(ctx context.Context, conn *websocket.Conn, client *Client) {
	// Handle incoming messages
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[SYS] %s read error: %v", client.ID, err)
			break
		}

		params := websocketMessage{}
		err = json.Unmarshal(message, &params)
		if err != nil {
			log.Printf("[SYS] %s Failed to unmarshal JSON: %v", client.ID, err)
			continue
		}
		// DEV log
		log.Printf("[DEV] %s sent: %s\n", client.ID, params)
		var response websocketMessage
		switch params.Channel {
		case "sys":
			log.Printf("[SYS] received: %s", message)
		case "console":
			response, err = cfg.handleConsole(ctx, conn, params.Data, client)
			if err != nil {
				log.Printf("[SYS] %s Console: %v", client.ID, err)
				response = websocketMessage{
					Channel: "console",
					Token:   "error",
					Data:    err.Error(),
				}
			}
		case "auth":
			if params.Token == "jwt" {
				jwtToken := params.Data

				userID, err := auth.ValidateJWT(jwtToken, cfg.JWTSecret)
				if err != nil {
					log.Printf("[SYS] %s invalid JWT", client.ID)
				} else {
					user, err := cfg.DB.GetUserByID(ctx, userID)
					if err != nil {
						log.Printf("[SYS] %s JWT not connected to known user", client.ID)
					} else {
						// Update client with authenticated user
						cfg.ClientsMu.Lock()
						delete(cfg.Clients, client.ID)
						client.IsAuthed = true
						client.ID = user.Username
						cfg.Clients[user.Username] = client
						cfg.ClientsMu.Unlock()

						response := websocketMessage{
							Channel: "sys",
							Token:   "authenticated",
							Data:    user.Username,
						}
						byteResponse, err := json.Marshal(response)
						if err != nil {
							log.Printf("[SYS] %s failed to marshal", client.ID)
						}
						client.Outbound <- byteResponse
					}
				}
				continue
			}
			select {
			case client.AuthChan <- params.Data:
				// Success
				continue
			default:
				log.Printf("[SYS] %s auth channel full", client.ID)
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
			log.Printf("[SYS] %s failed to marshal JSON: %s", client.ID, err)
			continue
		}
		client.Outbound <- byteResponse
	}
}

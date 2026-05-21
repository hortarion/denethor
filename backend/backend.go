package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	internalRegistry "github.com/hortarion/server/internal/apps"
	"github.com/hortarion/server/internal/auth"
	"github.com/hortarion/server/internal/database"

	"github.com/joho/godotenv"
)

// TODO: load URL and frontend port fron .env
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:8090"
	},
}

type websocketMessage struct {
	Token string `json:"token"`
	Data  string `json:"data"`
}

func (cfg ServerConfig) handleConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authChan := make(chan string, 1)
	ctx = context.WithValue(ctx, "authChan", authChan)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	connID := uuid.New().String()
	log.Printf("New WebSocket connection: %s from %s", connID, conn.RemoteAddr())

	outbound := make(chan []byte, 10)

	go func() {
		for msg := range outbound {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[%s] Write error: %v", connID, err)
				break
			}
		}
	}()

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
		// Handler logic here
		switch params.Token {
		case "sys":
			log.Printf("System received: %s", message)
		case "console":
			response, err = cfg.handleConsole(ctx, conn, params.Data, outbound)
			if err != nil {
				log.Printf("[%s] Console: %v", connID, err)
				continue
			}
		case "auth":
			authChan <- params.Data

			response.Token = "auth"
			response.Data = ""
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

type ServerConfig struct {
	Port     string
	DB       *database.Queries
	Platform string
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

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	cfg := ServerConfig{
		Port:     port,
		DB:       dbQueries,
		Platform: platform,
	}

	mux := http.NewServeMux()

	// Might replace with brocker logic
	mux.HandleFunc("/ws", cfg.handleConnection)
	mux.HandleFunc("/status", handleStatusPage)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	// port := os.Getenv("PORT")
	srv := http.Server{
		Handler:      mux,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	internalRegistry.InternalRegistry()

	// this blocks forever, until the server
	// has an unrecoverable error
	fmt.Printf("server started on http://localhost:%s\n", port)
	serverErr := srv.ListenAndServe()
	log.Fatal(serverErr)

}

// TODO: refactor commands and create commandRegistry
func (cfg ServerConfig) handleConsole(ctx context.Context, conn *websocket.Conn, message string, outbound chan<- []byte) (websocketMessage, error) {
	authChan, ok := ctx.Value("authChan").(chan string)
	cmd := strings.ToLower(strings.Split(message, " ")[0])
	args := strings.Split(message, " ")[1:]
	log.Println("[DEV] cmd:", cmd)
	for idx, arg := range args {
		log.Println("[DEV]", idx+1, ":", arg)
	}

	response := websocketMessage{
		Token: "console",
		Data:  "",
	}

	switch cmd {
	case "clear":
		response.Token = "sys"
		response.Data = "clear"
	case "help":
		response.Data = `Available commands:
clear - clear window
register <username> - sign up
login <username> - login`
	case "register":
		if len(args) == 0 {
			response.Data = "no username provided"
			return response, nil
		}
		if len(args[0]) == 0 {
			response.Data = "no username provided"
			return response, nil
		}
		exists, err := cfg.DB.CheckUserByName(ctx, args[0])
		if err != nil {
			return websocketMessage{}, err
		}
		if !exists {
			if !ok {
				return websocketMessage{}, fmt.Errorf("auth channel not found")
			}
			// GO func
			go cfg.registerUser(ctx, authChan, args[0], outbound)
			response.Token = "auth"
			response.Data = "type in your password"

		} else {
			response.Data = "username already taken"
		}
	case "login":
		if len(args) == 0 {
			response.Data = "no username provided"
			return response, nil
		}
		if len(args[0]) == 0 {
			response.Data = "no username provided"
			return response, nil
		}
		exists, err := cfg.DB.CheckUserByName(ctx, args[0])
		if err != nil {
			return websocketMessage{}, err
		}
		if exists {

			if !ok {
				return websocketMessage{}, fmt.Errorf("auth channel not found")
			}
			// GO func
			go cfg.loginUser(ctx, authChan, args[0], outbound)
			response.Token = "auth"
			response.Data = "type in your password"

		} else {
			response.Data = "username not registered"
		}
	case "ping":
		response.Data = "pong"
	default:
		return websocketMessage{}, nil
	}

	return response, nil
}

func (cfg *ServerConfig) registerUser(ctx context.Context, authChan <-chan string, username string, outbound chan<- []byte) {
	password := <-authChan
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
		Username:       username,
		HashedPassword: hash,
	})
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	response := websocketMessage{
		Token: "console",
		Data:  fmt.Sprintf("%s has been registered", user.Username),
	}
	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	outbound <- byteResponse
}

func (cfg *ServerConfig) loginUser(ctx context.Context, authChan <-chan string, username string, outbound chan<- []byte) {
	password := <-authChan
	user, err := cfg.DB.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
		return
	}
	response := websocketMessage{
		Token: "auth",
		Data:  "incorrect password",
	}
	valid, err := auth.CheckPasswordHash(password, user.HashedPassword)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
	}
	if valid {
		response.Token = "auth"
		response.Data = fmt.Sprintf("logged in as %s", user.Username)
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
		return
	}
	outbound <- byteResponse

}

func handleStatusPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	const page = `<html>
<style>
    :root {
    	--bg-color: #1e1e1e;
     	--text-color: #ffffff;
    }
    body {
            background-color: var(--bg-color);
            color: var(--text-color);
        }
</style>
<head></head>
<body>
	<p> Server Status OK </p>
</body>
</html>
`
	w.Write([]byte(page))
}

func (cfg *ServerConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	err := cfg.DB.ResetDB(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database has been reset."))
}

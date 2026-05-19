package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	internalRegistry "github.com/hortarion/server/internal/apps"
	"github.com/hortarion/server/internal/auth"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	connID := uuid.New().String()
	log.Printf("New WebSocket connection: %s from %s", connID, conn.RemoteAddr())

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[%s] read error: %v", connID, err)
			break
		}

		type parameters struct {
			Token string `json:"token"`
			Data  string `json:"data"`
		}
		params := parameters{}
		err = json.Unmarshal(message, &params)
		if err != nil {
			log.Printf("[%s] Failed to unmarshal JSON: %v", connID, err)
			continue
		}
		log.Printf("DEBUG[%s] sent: %v", connID, params)
		var response []byte
		// Handler logic here
		switch params.Token {
		case "sys":
			log.Printf("System received: %s", message)
		case "console":
			response, err = handleConsole(conn, params.Data)
			if err != nil {
				log.Printf("[%s] Console: %v", connID, err)
				continue
			}
		case "auth":
			_, err = auth.HandleAuth(conn, params.Data)
			if err != nil {
				log.Printf("[%s] Auth: %v", connID, err)
				continue
			}
		default:
			response = []byte{}
		}
		if err := conn.WriteMessage(messageType, response); err != nil {
			log.Printf("[%s] Write error: %v", connID, err)
			break
		}
	}
	log.Printf("Connection %s closed", connID)
}

func main() {

	type serverConfig struct {
		Port string
		DB   string
	}

	godotenv.Load()
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

	cfg := serverConfig{
		Port: port,
		DB:   dbURL,
	}

	mux := http.NewServeMux()

	// Might replace with brocker logic
	mux.HandleFunc("/ws", handleConnection)
	mux.HandleFunc("/status", handleStatusPage)

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
	err := srv.ListenAndServe()
	log.Fatal(err)

}

func handleConsole(conn *websocket.Conn, message string) ([]byte, error) {
	cmd := strings.ToLower(strings.Split(message, " ")[0])
	args := strings.Split(message, " ")[1:]
	log.Println("DEV cmd:", cmd)
	for idx, arg := range args {
		log.Println("DEV", idx+1, ":", arg)
	}

	var response []byte
	switch cmd {
	case "clear":
		response = []byte("clear")
	case "help":
		response = []byte(`Available commands:
clear - clear window
register <username> - sign up
login <username> - login`)
	case "register":
		response = []byte("maskedInput")
	case "login":
		response = []byte("not yet implemented")
	case "ping":
		response = []byte("pong")
	default:
		return nil, nil
	}
	return response, nil
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

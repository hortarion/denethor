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

	"github.com/gorilla/websocket"
	"github.com/hortarion/server/api"
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

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received: %s", message)
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
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
	mux.HandleFunc("/", handleStatusPage)
	mux.HandleFunc("/api/console", api.MiddlewareCORS(handleConsole))
	mux.HandleFunc("/api/auth", api.MiddlewareCORS(auth.HandleAuth))

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

func handleConsole(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Input string `json:"input"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("Failed to decode json")
		return
	}
	cmd := strings.ToLower(strings.Split(params.Input, " ")[0])
	args := strings.Split(params.Input, " ")[1:]
	fmt.Println("cmd:", cmd)
	for idx, arg := range args {
		fmt.Println(idx+1, ":", arg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if cmd == "clear" {
		w.Write([]byte("clear"))
		return
	}
	if cmd == "help" {
		helpMessage := `Available commands:
clear - clear window
register <username> - sign up
login <username> - login`
		w.Write([]byte(helpMessage))
		return
	}
	if cmd == "register" {
		w.Write([]byte("maskedInput"))
		return
	}
	if cmd == "login" {
		w.Write([]byte("not yet implemented"))
		return
	}
	if cmd == "ping" {
		w.Write([]byte("pong"))
		return
	}
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

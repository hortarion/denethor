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

	"github.com/hortarion/server/api"
	internalRegistry "github.com/hortarion/server/internal/apps"
	"github.com/joho/godotenv"
)

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
	mux.HandleFunc("/", handleStatusPage)
	mux.HandleFunc("/api/console", api.MiddlewareCORS(handleInput))

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

func handleInput(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte("Available commands:\nclear - clear window"))
		return
	}
	w.Write([]byte("hello"))
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

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const PORT = "8080"

func main() {
	m := http.NewServeMux()

	m.HandleFunc("/", handlePage)

	// port := os.Getenv("PORT")
	port := PORT
	srv := http.Server{
		Handler:      m,
		Addr:         ":" + port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// this blocks forever, until the server
	// has an unrecoverable error
	fmt.Printf("server started on http://localhost:%s\n", port)
	err := srv.ListenAndServe()
	log.Fatal(err)

}

func handlePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	const page = `<html>
<head></head>
<body>
	<p> Server </p>
</body>
</html>
`
	w.Write([]byte(page))
}

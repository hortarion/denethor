package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
}

const PORT = "8090"
const filePathRoot = "public"

func main() {

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		platform:       "dev",
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.RedirectHandler("/public/", http.StatusFound))
	mux.Handle("/public/", apiCfg.middlewareMetricsInc(http.StripPrefix("/public", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)

	// port := os.Getenv("PORT")
	port := PORT
	srv := http.Server{
		Handler:      mux,
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

func (apiCfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, apiCfg.fileserverHits.Load())))
}

func (apiCfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

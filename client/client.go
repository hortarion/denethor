package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
	port           string
	filePathRoot   string
}

func main() {

	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		platform = "prod"
		log.Println("PLATFORM set to prod")
	}
	filePathRoot := os.Getenv("FILE_PATH_ROOT")
	if filePathRoot == "" {
		log.Fatal("FILE_PATH_ROOT must be set")
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		platform:       platform,
		port:           port,
		filePathRoot:   filePathRoot,
	}

	mux := http.NewServeMux()
	// Serve webpage
	mux.Handle("/", http.RedirectHandler("/public/", http.StatusFound))
	mux.Handle("/favicon.ico", http.RedirectHandler("/public/favicon.ico", http.StatusFound))
	mux.Handle("/public/", apiCfg.middlewareMetricsInc(http.StripPrefix("/public", http.FileServer(http.Dir(filePathRoot)))))
	// APIs
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)

	srv := http.Server{
		Handler:      mux,
		Addr:         ":" + apiCfg.port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	fmt.Printf("server started on http://localhost:%s\n", port)
	err := srv.ListenAndServe()
	log.Fatal(err)

}

func (apiCfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Appendf(nil, `<html>
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
	<body>
    	<h1>Denethor - Admin</h1>
    	<p>Denthor has been visited %d times!</p>
    </body>
</html>`, apiCfg.fileserverHits.Load())))
}

func (apiCfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/public/" {
			apiCfg.fileserverHits.Add(1)
		}
		next.ServeHTTP(w, r)
	})
}

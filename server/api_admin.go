package main

import "net/http"

func (cfg *serverConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	err := cfg.DB.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset users: " + err.Error()))
		return
	}
	err = cfg.DB.ResetRefreshTokens(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset refresh_tokens: " + err.Error()))
	}
	err = cfg.DB.ResetApps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset apps: " + err.Error()))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database has been reset."))
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

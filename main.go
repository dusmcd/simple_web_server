package main

import (
	"fmt"
	"net/http"
)

func main() {
	err := runServer()
	if err != nil {
		fmt.Println(err)
	}
}

func runServer() error {
	serveMux := http.NewServeMux()
	config := &apiConfig{
		fileServerHits: 0,
	}

	serveMux.Handle("GET /app/*", config.middlewareMetricsInc(fileServerHandler()))
	serveMux.HandleFunc("GET /api/healthz", readinessHandler)
	serveMux.HandleFunc("GET /admin/metrics", config.showMetricsHandler)
	serveMux.HandleFunc("GET /api/reset", config.resetMetricsHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", validateHandler)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: serveMux,
	}

	fmt.Println("Server listening on port 8080...")
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

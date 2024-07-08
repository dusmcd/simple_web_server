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
	serveMux.HandleFunc("GET /healthz", handler)
	serveMux.HandleFunc("GET /metrics", config.showMetricsHandler)
	serveMux.HandleFunc("GET /reset", config.resetMetricsHandler)

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

package main

import (
	"fmt"
	"log"
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
	db, err := NewDB("database.json")
	if err != nil {
		log.Println(err)
		return err
	}
	config := &apiConfig{
		fileServerHits: 0,
		db:             db,
	}
	registerHandlers(serveMux, config)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: serveMux,
	}

	fmt.Println("Server listening on port 8080...")
	err = server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func registerHandlers(serveMux *http.ServeMux, config *apiConfig) {
	serveMux.Handle("GET /app/*", config.middlewareMetricsInc(fileServerHandler()))
	serveMux.HandleFunc("GET /api/healthz", readinessHandler)
	serveMux.HandleFunc("GET /admin/metrics", config.showMetricsHandler)
	serveMux.HandleFunc("GET /api/reset", config.resetMetricsHandler)
	serveMux.HandleFunc("POST /api/chirps", config.saveChirpsHandler)
	serveMux.HandleFunc("GET /api/chirps", config.getChirpsHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", config.getChirpByIdHandler)
	serveMux.HandleFunc("POST /api/users", config.saveUserHandler)
	serveMux.HandleFunc("POST /api/login", config.loginUsersHandler)
}

package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
}

func (config *apiConfig) showMetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("Hits: %d", config.fileServerHits)))
}

func (config *apiConfig) resetMetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	config.fileServerHits = 0
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Metrics reset"))
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response := []byte("OK")
	w.Write(response)
}

func fileServerHandler() http.Handler {
	return http.StripPrefix("/app", http.FileServer(http.Dir(".")))
}

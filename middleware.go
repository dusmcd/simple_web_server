package main

import (
	"net/http"
)

func (config *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		config.fileServerHits++
		next.ServeHTTP(w, req)
	})
}

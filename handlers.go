package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
}

func (config *apiConfig) showMetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>	
	</html`, config.fileServerHits)
	w.Write([]byte(html))
}

func (config *apiConfig) resetMetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	config.fileServerHits = 0
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Metrics reset"))
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response := []byte("OK")
	w.Write(response)
}

func fileServerHandler() http.Handler {
	fileServer := http.FileServer(http.Dir("."))
	return http.StripPrefix("/app", fileServer)
}

func validateHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type response struct {
		Valid bool `json:"valid"`
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, response{Valid: true})
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	type errorRes struct {
		Error string `json:"error"`
	}
	response := errorRes{
		Error: msg,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Print(err)
		respondWithError(w, 500, "error marshalling json")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "error marshalling json")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)

}

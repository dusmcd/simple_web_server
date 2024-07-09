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

	type errorResponse struct {
		Error string `json:"error"`
	}

	type response struct {
		Valid bool `json:"valid"`
	}

	responseData := []byte{}
	responseCode := 0

	if len(params.Body) > 140 {
		errorResponse := errorResponse{
			Error: "Chirp is too long",
		}
		responseCode = 400
		responseData, err = json.Marshal(errorResponse)
	} else {
		response := response{
			Valid: true,
		}
		responseCode = 200
		responseData, err = json.Marshal(response)
	}

	if err != nil {
		errorResponse := errorResponse{
			Error: err.Error(),
		}
		responseData, err = json.Marshal(errorResponse)
		log.Println(err)
	}
	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)

}

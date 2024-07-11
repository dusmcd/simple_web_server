package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

func saveChirpsHandler(w http.ResponseWriter, req *http.Request) {
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

	db, err := NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := saveToDB(db, params.Body)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	//cleanBody := cleanMessage(params.Body)

	respondWithJSON(w, 201, chirp)
}

func getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	db, err := NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, chirps)
}

func saveToDB(db *DB, body string) (Chirp, error) {
	chirp, err := db.CreateChirp(body)
	if err != nil {
		return Chirp{}, err
	}
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	dbStructure.Chirps[chirp.ID] = chirp
	// concurrently write updated dbStructure to disk using write function
	err = db.WriteDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
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
		w.WriteHeader(500)
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

func cleanMessage(body string) string {
	badWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	cleanBody := make([]string, len(words))
	for i, word := range words {
		result := ""
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				result = "****"
				break
			} else {
				result = word
			}
		}
		cleanBody[i] = result
	}
	return strings.Join(cleanBody, " ")
}

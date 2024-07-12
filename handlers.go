package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
	db             *DB
}

/*
route: /admin/metrics
method: GET
*/
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

/*
route: /api/reset
method: GET
*/
func (config *apiConfig) resetMetricsHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	config.fileServerHits = 0
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Metrics reset"))
}

/*
route: /api/healthz
method: GET
*/
func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response := []byte("OK")
	w.Write(response)
}

/*
route: /app/*
method: GET
*/
func fileServerHandler() http.Handler {
	fileServer := http.FileServer(http.Dir("."))
	return http.StripPrefix("/app", fileServer)
}

/*
route: /api/chirps
method: POST
*/
func (config *apiConfig) saveChirpsHandler(w http.ResponseWriter, req *http.Request) {
	params, err := decodeJSON(req)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !validateChirp(params.Body) {
		respondWithError(w, 400, "chirp is too long")
		return
	}

	chirp, err := saveToDB(config.db, cleanMessage(params.Body))
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, chirp)
}

/*
route: /api/chirps
method: GET
*/
func (config *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := config.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, chirps)
}

/*
route: /api/chirps/{chirpID}
method: GET
*/
func (config *apiConfig) getChirpByIdHandler(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	chirp, err := config.db.GetChirpById(chirpID)
	if err != nil {
		if err.Error() == "Chirp not found" {
			respondWithError(w, 404, err.Error())
			return
		}
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, chirp)
}

/*
route: /api/users
method: POST
*/
func (config *apiConfig) saveUserHandler(w http.ResponseWriter, req *http.Request) {
	params, err := decodeJSON(req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user, err := config.db.CreateUser(params.Email)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	dbStructure, err := config.db.LoadDB()
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	dbStructure.Users[user.ID] = user
	err = config.db.WriteDB(dbStructure)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, user)
}

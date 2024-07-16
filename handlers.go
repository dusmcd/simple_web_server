package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
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
	html := fmt.Sprintf(`<!DOCTYPE html><html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>	
	</html>`, config.fileServerHits)
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

	req body shape: {
		body string
	}
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

	chirp, err := saveChirpToDB(config.db, cleanMessage(params.Body))
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

	req body shape: {
		email string
		password string
	}
*/
func (config *apiConfig) saveUserHandler(w http.ResponseWriter, req *http.Request) {
	params, err := decodeJSON(req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user, err := saveUserToDB(config.db, params.Email, params.Password)
	if err != nil {
		if err.Error() == "the provided email has already been registered" {
			respondWithError(w, 401, err.Error())
			return
		}
		respondWithError(w, 500, err.Error())
		return
	}

	// removing password from response
	response := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    user.ID,
		Email: user.Email,
	}
	respondWithJSON(w, 201, response)
}

/*
route: /api/login
method: POST

	req body shape: {
		email string
		password string
	}
*/
func (config *apiConfig) loginUsersHandler(w http.ResponseWriter, req *http.Request) {
	params, err := decodeJSON(req)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	foundUser, err := findUser(config.db, params.Email)

	if err != nil {
		if err.Error() == "user not found" {
			respondWithError(w, 404, err.Error())
			return
		}
		respondWithError(w, 500, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(params.Password))
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	response := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{
		ID:    foundUser.ID,
		Email: foundUser.Email,
	}

	respondWithJSON(w, 200, response)
}

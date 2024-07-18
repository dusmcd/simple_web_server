package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileServerHits int
	db             *DB
	jwtSecret      string
	polkaKey       string
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
	jwtToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	id, err := validateToken(jwtToken)

	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

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

	chirp, err := saveChirpToDB(config.db, cleanMessage(params.Body), id)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, chirp)
}

/*
route: /api/chirps?author_id={optional}&sort={asc | desc | optional}
method: GET
*/
func (config *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	authorID := req.URL.Query().Get("author_id")
	authorIDnum, _ := strconv.Atoi(authorID)
	sort := req.URL.Query().Get("sort")

	chirps, err := config.db.GetChirps(authorIDnum, sort)
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
		ID          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}{
		ID:          user.ID,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
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

	token, err := createJWT(config.jwtSecret, foundUser.ID)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	refreshToken, err := saveRefreshToken(config.db, foundUser.ID)
	if err != nil {
		respondWithError(w, 500, "error generating refresh token")
		return
	}

	response := struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}{
		ID:           foundUser.ID,
		Email:        foundUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  foundUser.IsChirpyRed,
	}

	respondWithJSON(w, 200, response)
}

/*
route: /api/users
method: PUT

	req body shape: {
		email string
		password string
	}

	req headers: {
		Authorization string (jwtToken)
	}
*/
func (config *apiConfig) updateUsersHandler(w http.ResponseWriter, req *http.Request) {
	jwtToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	id, err := validateToken(jwtToken)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	params, err := decodeJSON(req)
	if err != nil {
		respondWithError(w, 500, "error decoding request body")
		return
	}

	updatedUser, err := updateUserInDB(config.db, id, params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "error updating user")
		return
	}

	response := struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		Token       string `json:"token"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}{
		ID:          id,
		Email:       updatedUser.Email,
		Token:       jwtToken,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	respondWithJSON(w, 200, response)
}

/*
route: /api/refresh
method: POST

	req headers: {
		Authorization string (jwtToken)
	}
*/
func (config *apiConfig) refreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	refreshToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	validToken, userId := validateRefreshToken(config.db, refreshToken)

	if !validToken {
		respondWithError(w, 401, "refresh token is invalid")
		return
	}

	newToken, err := createJWT(config.jwtSecret, userId)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{Token: newToken})
}

/*
route: /api/revoke
method: POST

	req headers: {
		Authorization: string (JWT)
	}
*/
func (config *apiConfig) revokeRefreshHandler(w http.ResponseWriter, req *http.Request) {
	refreshToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	err := deleteRefreshTokenFromDB(config.db, refreshToken)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	w.WriteHeader(204)
}

/*
route: /api/chirps
method: DELETE

	req headers: {
		Authorization: string (JWT)
	}
*/
func (config *apiConfig) deleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	jwtToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	id, err := validateToken(jwtToken)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	chirpID := req.PathValue("chirpID")
	chirp, err := config.db.GetChirpById(chirpID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	if id != chirp.AuthorID {
		respondWithError(w, 403, "you are not authorized to delete this chirp")
		return
	}

	chirpIDNum, _ := strconv.Atoi(chirpID)
	err = deleteChirpFromDB(config.db, chirpIDNum)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	w.WriteHeader(204)
}

/*
route: /api/polka/webhooks
method: POST

	req body shape: {
		event: string
		data: {
			user_id: int
		}
	}
*/
func (config *apiConfig) webhookHandler(w http.ResponseWriter, req *http.Request) {
	key := strings.TrimPrefix(req.Header.Get("Authorization"), "ApiKey ")
	if key != config.polkaKey {
		respondWithError(w, 401, "unauthorized request")
		return
	}

	params, err := decodeJSON(req)
	if err != nil {
		respondWithError(w, 500, "error decoding json")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	err = upgradeUserInDB(config.db, params.Data.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			respondWithError(w, 404, err.Error())
			return
		}
		respondWithError(w, 500, err.Error())
	}

	w.WriteHeader(204)
}

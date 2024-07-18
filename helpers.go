package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type parameters struct {
	Body             string `json:"body"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func saveChirpToDB(db *DB, body string) (Chirp, error) {
	chirp, err := db.CreateChirp(body)
	if err != nil {
		return Chirp{}, err
	}
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	dbStructure.Chirps[chirp.ID] = chirp

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

func validateChirp(body string) bool {
	return len(body) <= 140
}

func decodeJSON(req *http.Request) (parameters, error) {
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return params, err
	}

	return params, nil

}

func saveUserToDB(db *DB, email, password string) (User, error) {
	user, err := db.CreateUser(email, password)
	if err != nil {
		return User{}, err
	}
	dbStructure, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}

	dbStructure.Users[user.ID] = user
	err = db.WriteDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil

}

func findUser(db *DB, email string) (User, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}

	foundUser := User{}
	found := false

	for id := range dbStructure.Users {
		if email == dbStructure.Users[id].Email {
			foundUser = dbStructure.Users[id]
			found = true
			break
		}
	}
	if !found {
		return User{}, errors.New("user not found")
	}

	return foundUser, nil

}

func createJWT(secretKey string, userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour).UTC()),
		Subject:   strconv.Itoa(userId),
	})
	s, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", err
	}

	return s, nil
}

func validateToken(jwtToken string) (int, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
	if err != nil {
		return 0, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return 0, errors.New("error getting user id")
	}

	numberId, err := strconv.Atoi(id)
	if err != nil {
		return 0, errors.New("error converting id to integer")
	}

	return numberId, nil
}

func updateUserInDB(db *DB, id int, email, password string) (User, error) {
	user, err := db.UpdateUserById(id, email, password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func generateRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func saveRefreshToken(db *DB, userId int) (string, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return "", err
	}

	user, found := dbStructure.Users[userId]
	if !found {
		return "", errors.New("user not found")
	}

	token := generateRefreshToken()
	refreshToken := RefreshToken{
		Token:      token,
		CreatedAt:  time.Now().UTC(),
		DaysActive: 60,
	}
	user.RefreshToken = refreshToken
	dbStructure.Users[userId] = user

	err = db.WriteDB(dbStructure)
	if err != nil {
		return "", err
	}

	return token, nil

}

func validateRefreshToken(db *DB, refreshToken string) (bool, int) {
	dbStructure, _ := db.LoadDB()

	for userId := range dbStructure.Users {
		if refreshToken == dbStructure.Users[userId].RefreshToken.Token {
			return true, userId
		}
	}

	return false, 0
}

func deleteRefreshTokenFromDB(db *DB, refreshToken string) error {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return err
	}

	for userId := range dbStructure.Users {
		if refreshToken == dbStructure.Users[userId].RefreshToken.Token {
			user := dbStructure.Users[userId]
			user.RefreshToken = RefreshToken{}
			dbStructure.Users[userId] = user
			db.WriteDB(dbStructure)
			return nil
		}
	}

	return nil
}

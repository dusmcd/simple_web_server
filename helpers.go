package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type parameters struct {
	Body     string `json:"body"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

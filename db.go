package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.Mutex
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.Mutex{},
	}
	err := db.ensureDB()

	if err != nil {
		log.Print(err)
		return &DB{}, err
	}

	return &db, nil
}

func (db *DB) ensureDB() error {
	_, err := os.Open(db.path)
	if err != nil {
		return db.createFile()
	}
	return nil
}

func (db *DB) createFile() error {
	_, err := os.Create(db.path)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) LoadDB() (DBStructure, error) {
	fileContent, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	dbStructure := DBStructure{
		Chirps: make(map[int]Chirp),
	}
	if len(fileContent) == 0 {
		return dbStructure, nil
	}
	err = json.Unmarshal(fileContent, &dbStructure)
	if err != nil {
		fmt.Println("JSON error in LoadDB function")
		return DBStructure{}, err
	}
	return dbStructure, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		Body: body,
		ID:   id,
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.LoadDB()
	chirps := []Chirp{}
	if err != nil {
		return []Chirp{}, err
	}
	for id := range dbStructure.Chirps {
		chirps = append(chirps, dbStructure.Chirps[id])
	}

	return chirps, nil
}

func (db *DB) WriteDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		log.Println("JSON error! from WriteDB function")
		return err
	}

	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

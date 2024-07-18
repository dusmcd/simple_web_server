package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.Mutex
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

type User struct {
	ID           int          `json:"id"`
	Email        string       `json:"email"`
	Password     string       `json:"password"`
	RefreshToken RefreshToken `json:"refresh_token"`
	IsChirpyRed  bool         `json:"is_chirpy_red"`
}

type RefreshToken struct {
	Token      string    `json:"token"`
	CreatedAt  time.Time `json:"created_at"`
	DaysActive int       `json:"days_active"`
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
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
		Users:  make(map[int]User),
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

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		Body:     body,
		ID:       id,
		AuthorID: authorID,
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

	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })
	return chirps, nil
}

func (db *DB) WriteDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetChirpById(id string) (Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}

	integerId, err := strconv.Atoi(id)
	if err != nil {
		return Chirp{}, err
	}

	chirp, found := dbStructure.Chirps[integerId]
	if !found {
		return Chirp{}, errors.New("Chirp not found")
	}

	return chirp, nil

}

func (db *DB) CreateUser(email, password string) (User, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1

	// make sure duplicate emails are not created
	for id := range dbStructure.Users {
		if dbStructure.Users[id].Email == email {
			return User{}, errors.New("the provided email has already been registered")
		}
	}
	// hash password for storage
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)

	if err != nil {
		return User{}, err
	}

	user := User{
		ID:          id,
		Email:       email,
		Password:    string(hashedPassword),
		IsChirpyRed: false,
	}

	return user, nil
}

func (db *DB) UpdateUserById(id int, email, password string) (User, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}

	user, found := dbStructure.Users[id]
	if !found {
		return User{}, errors.New("invalid id")
	}

	user.Email = email
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		return User{}, err
	}

	user.Password = string(hashedPassword)

	dbStructure.Users[id] = user
	err = db.WriteDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

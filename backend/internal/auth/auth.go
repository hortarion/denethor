package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexedwards/argon2id"
)

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Input string `json:"input"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("Failed to decode json")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Println("read:", params.Input)
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, err
}

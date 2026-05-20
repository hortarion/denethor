package auth

import (
	"log"

	"github.com/alexedwards/argon2id"
	"github.com/gorilla/websocket"
)

func HandleAuth(conn *websocket.Conn, message string) (string, error) {
	log.Println("DEV auth:", message)
	hash, err := HashPassword(message)
	if err != nil {
		return "", err
	}
	log.Println("DEV auth: hash ", hash)
	match, err := CheckPasswordHash(message, hash)
	if err != nil {
		return "", err
	}
	log.Println("DEV auth: match ", match)
	return "", nil
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

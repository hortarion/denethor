package auth

import (
	"log"

	"github.com/alexedwards/argon2id"
	"github.com/gorilla/websocket"
)

func HandleAuth(conn *websocket.Conn, message string) ([]byte, error) {
	log.Println("read:", message)
	return []byte{}, nil
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

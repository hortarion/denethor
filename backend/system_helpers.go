package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hortarion/server/internal/auth"
)

func (cfg *serverConfig) handleClear(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	response := websocketMessage{
		Channel: "sys",
		Token:   "clear",
		Data:    "",
	}
	return response, nil
}

func (cfg *serverConfig) sysAuthenticated(client *Client) {
	response := websocketMessage{
		Channel: "sys",
		Token:   "authenticated",
		Data:    client.ID,
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SYS] %s failed to marshal system message", client.ID)
	}

	client.Outbound <- byteResponse
}

func (cfg *serverConfig) sysLogout(client *Client) {
	response := websocketMessage{
		Channel: "sys",
		Token:   "logout",
		Data:    "",
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SYS] %s failed to marshal system message", client.ID)
	}

	client.Outbound <- byteResponse
}

func (cfg *serverConfig) sysJWT(ctx context.Context, client *Client) {
	user, err := cfg.DB.GetUserByUsername(ctx, client.ID)
	if err != nil {
		log.Printf("[SYS] %v", err)
	}
	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		log.Printf("[SYS] %v", err)
	}
	response := websocketMessage{
		Channel: "sys",
		Token:   "JWT",
		Data:    token,
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SYS] %s failed to marshal system message", client.ID)
	}
	client.Outbound <- byteResponse
}

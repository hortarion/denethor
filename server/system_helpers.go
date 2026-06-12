package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hortarion/server/internal/auth"
	"github.com/hortarion/server/internal/database"
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
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		log.Printf("[SYS] %v", err)
		return
	}
	response := websocketMessage{
		Channel: "sys",
		Token:   "JWT",
		Data:    token,
	}

	cfg.marshalAndSend(response, client)
}

func (cfg *serverConfig) sysRFT(ctx context.Context, client *Client) {
	RFtoken := auth.MakeRefreshToken()
	user, err := cfg.DB.GetUserByUsername(ctx, client.ID)
	if err != nil {
		log.Printf("[SYS] %s not found", client.ID)
		return
	}
	cfg.DB.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		Token:     RFtoken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	response := websocketMessage{
		Channel: "sys",
		Token:   "RFT",
		Data:    RFtoken,
	}
	cfg.marshalAndSend(response, client)
}

func (cfg *serverConfig) getUserByRefreshToken(ctx context.Context, refreshToken string) (database.User, error) {
	user, err := cfg.DB.GetUserFromRefreshToken(ctx, refreshToken)
	if err != nil {
		return database.User{}, err
	}
	return user, nil
}

func (cfg *serverConfig) handlerRevoke(ctx context.Context, client *Client) {
	user, err := cfg.DB.GetUserByUsername(ctx, client.ID)
	if err != nil {
		log.Printf("[SYS] %s not found", client.ID)
		return
	}
	cfg.DB.RevokeRefreshToken(ctx, user.ID)
}

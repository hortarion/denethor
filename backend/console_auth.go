package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hortarion/server/internal/auth"
	"github.com/hortarion/server/internal/database"
)

func (cfg *serverConfig) registerUser(ctx context.Context, client *Client, username string) {
	password := <-client.AuthChan
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
		Username:       username,
		HashedPassword: hash,
	})
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	response := websocketMessage{
		Channel: "console",
		Data:    fmt.Sprintf("%s has been registered", user.Username),
	}
	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[REGIST] error: %s", err)
		return
	}
	client.Outbound <- byteResponse
}

func (cfg *serverConfig) loginUser(ctx context.Context, client *Client, username string) {
	password := <-client.AuthChan
	user, err := cfg.DB.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
		return
	}
	response := websocketMessage{
		Channel: "auth",
		Data:    "incorrect password",
	}
	valid, err := auth.CheckPasswordHash(password, user.HashedPassword)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
	}
	if valid {
		response.Channel = "auth"
		response.Data = fmt.Sprintf("logged in as %s", user.Username)
		client.IsAuthed = true
		client.ID = username
		cfg.sysAuthenticated(client)
		cfg.sysJWT(ctx, client)
		log.Printf("[SYS] %s logged in", client.ID)
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
		return
	}
	client.Outbound <- byteResponse
}

func (cfg *serverConfig) handleRegister(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	if client.IsAuthed {
		return websocketMessage{
			Channel: "console",
			Token:   "",
			Data:    fmt.Sprintf("Already logged in as %s", client.ID),
		}, nil
	}
	response := websocketMessage{
		Channel: "",
		Token:   "",
		Data:    "",
	}
	if len(args) == 0 {
		response.Channel = "console"
		response.Data = "no username provided"
		return response, nil
	}
	if len(args[0]) == 0 {
		response.Channel = "console"
		response.Data = "no username provided"
		return response, nil
	}
	exists, err := cfg.DB.CheckUserByName(ctx, args[0])
	if err != nil {
		return websocketMessage{}, err
	}
	if !exists {
		// GO func
		go cfg.registerUser(ctx, client, args[0])
		response.Channel = "auth"
		response.Token = "password"

	} else {
		response.Channel = "console"
		response.Data = "username already taken"
	}
	return response, nil
}

func (cfg *serverConfig) handleLogin(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	response := websocketMessage{
		Channel: "",
		Token:   "",
		Data:    "",
	}

	cfg.ClientsMu.Lock()
	client.AuthChan = make(chan string, 1)
	cfg.ClientsMu.Unlock()

	if len(args) == 0 {
		response.Channel = "console"
		response.Data = "no username provided"
		return response, nil
	}
	if len(args[0]) == 0 {
		response.Channel = "console"
		response.Data = "no username provided"
		return response, nil
	}
	exists, err := cfg.DB.CheckUserByName(ctx, args[0])
	if err != nil {
		return websocketMessage{}, err
	}
	if exists {
		// GO func
		go cfg.loginUser(ctx, client, args[0])
		response.Channel = "auth"
		response.Token = "password"

	} else {
		response.Channel = "console"
		response.Data = "username not registered"
	}
	return response, nil
}

func (cfg *serverConfig) handleLogout(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	// Reset authentication state
	cfg.ClientsMu.Lock()
	if existingClient, exists := cfg.Clients[client.ID]; exists {
		existingClient.IsAuthed = false
		// Create new auth channel to clear any pending auth state
		existingClient.AuthChan = make(chan string, 1)
	}
	cfg.ClientsMu.Unlock()

	// Clear JWT
	response := websocketMessage{
		Channel: "sys",
		Token:   "logout",
		Data:    "",
	}

	byteResponse, _ := json.Marshal(response)
	client.Outbound <- byteResponse

	clearResponse := websocketMessage{
		Channel: "sys",
		Token:   "clear",
		Data:    "",
	}

	byteClearResponse, _ := json.Marshal(clearResponse)
	client.Outbound <- byteClearResponse

	return websocketMessage{
		Channel: "console",
		Data:    "Logged out successfully",
	}, nil
}

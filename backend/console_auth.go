package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
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
		Data:    fmt.Sprintf("%s has been registered and logged in", user.Username),
	}
	client.IsAuthed = true
	client.ID = username
	cfg.sysAuthenticated(client)
	cfg.sysJWT(ctx, client)
	cfg.sysRFT(ctx, client)
	cfg.marshalAndSend(response, client)
	cfg.marshalAndSend(websocketMessage{Channel: "auth"}, client)
	log.Printf("[SYS] %s logged in", client.ID)
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
		cfg.sysRFT(ctx, client)
		log.Printf("[SYS] %s logged in", client.ID)
	}

	cfg.marshalAndSend(response, client)
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
	newID := uuid.New().String()

	cfg.handlerRevoke(ctx, client)

	// Reset authentication state
	cfg.ClientsMu.Lock()
	defer cfg.ClientsMu.Unlock()
	_, exists := cfg.Clients[client.ID]
	if exists {
		delete(cfg.Clients, client.ID)
	}

	client.ID = newID
	client.IsAuthed = false
	client.AuthChan = make(chan string, 1)
	cfg.Clients[newID] = client

	cfg.sysLogout(client)

	cfg.marshalAndSend(
		websocketMessage{
			Channel: "sys",
			Token:   "clear",
			Data:    "",
		},
		client,
	)

	return websocketMessage{
		Channel: "console",
		Data:    "Logged out successfully",
	}, nil
}

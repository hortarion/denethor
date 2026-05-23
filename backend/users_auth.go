package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hortarion/server/internal/auth"
	"github.com/hortarion/server/internal/database"
)

func (cfg *serverConfig) registerUser(ctx context.Context, authChan <-chan string, username string, outbound chan<- []byte) {
	password := <-authChan
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
	outbound <- byteResponse
}

func (cfg *serverConfig) loginUser(ctx context.Context, authChan <-chan string, username string, outbound chan<- []byte) {
	password := <-authChan
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
	}

	byteResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[LOGIN] error: %s", err)
		return
	}
	outbound <- byteResponse
}

func (cfg *serverConfig) handleRegister(ctx context.Context, authChan chan string, outbound chan<- []byte, args []string) (websocketMessage, error) {
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
		go cfg.registerUser(ctx, authChan, args[0], outbound)
		response.Channel = "auth"
		response.Token = "password"

	} else {
		response.Channel = "console"
		response.Data = "username already taken"
	}
	return response, nil
}

func (cfg *serverConfig) handleLogin(ctx context.Context, authChan chan string, outbound chan<- []byte, args []string) (websocketMessage, error) {
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
	if exists {
		// GO func
		go cfg.loginUser(ctx, authChan, args[0], outbound)
		response.Channel = "auth"
		response.Token = "password"

	} else {
		response.Channel = "console"
		response.Data = "username not registered"
	}
	return response, nil
}

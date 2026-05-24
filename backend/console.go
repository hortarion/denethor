package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

type cliCommand struct {
	name        string
	description string
	callback    func(
		ctx context.Context,
		client *Client,
		args []string,
	) (websocketMessage, error)
}

// Console command registry
func (cfg *serverConfig) getConsoleCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"clear": {
			name:        "clear",
			description: "Clear the screen",
			callback:    cfg.handleClear,
		},
		"help": {
			name:        "help",
			description: "Display available commands",
			callback:    cfg.handleHelp,
		},
		"login": {
			name:        "login",
			description: "Login to existing user account",
			callback:    cfg.handleLogin,
		},
		"logout": {
			name:        "logout",
			description: "Logout from user account",
			callback:    cfg.handleLogout,
		},
		"ping": {
			name:        "ping",
			description: "Ping the server",
			callback:    cfg.handlePing,
		},
		"register": {
			name:        "register",
			description: "Register a new user account",
			callback:    cfg.handleRegister,
		},
		"shout": {
			name:        "shout",
			description: "Broadcast to all clients",
			callback:    cfg.handleShout,
		},
	}
}

func (cfg *serverConfig) handleConsole(ctx context.Context, _ *websocket.Conn, message string, client *Client) (websocketMessage, error) {
	authChan, ok := ctx.Value("authChan").(chan string)
	client.AuthChan = authChan
	if !ok {
		return websocketMessage{}, fmt.Errorf("auth channel not found")
	}
	cmd := strings.ToLower(strings.Split(message, " ")[0])
	args := strings.Split(message, " ")[1:]

	response := websocketMessage{}

	command, exists := cfg.getConsoleCommands()[cmd]
	if exists {
		return command.callback(ctx, client, args)
	} else {
		return response, nil
	}
}

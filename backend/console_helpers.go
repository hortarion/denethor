package main

import (
	"context"
	"fmt"
	"strings"
)

func (cfg *serverConfig) handleClear(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	response := websocketMessage{
		Channel: "sys",
		Token:   "",
		Data:    "clear",
	}
	return response, nil
}

func (cfg *serverConfig) handleHelp(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	builder := strings.Builder{}
	for _, command := range cfg.getCommands() {
		builder.WriteString(fmt.Sprintf("%s - %s\n", command.name, command.description))
	}
	response := websocketMessage{
		Channel: "console",
		Token:   "",
		Data:    builder.String(),
	}
	return response, nil
}

func (cfg *serverConfig) handlePing(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	return websocketMessage{
		Channel: "console",
		Token:   "",
		Data:    "pong",
	}, nil
}

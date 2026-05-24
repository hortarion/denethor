package main

import (
	"context"
	"fmt"
	"sort"
)

func (cfg *serverConfig) handleHelp(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	commands := cfg.getConsoleCommands()
	var cmdKeys = make([]string, 0, len(commands))
	for k := range commands {
		cmdKeys = append(cmdKeys, k)
	}
	sort.Strings(cmdKeys)
	var helpMessage string
	for _, key := range cmdKeys {
		helpMessage += fmt.Sprintf("	> %s - %s\n", commands[key].name, commands[key].description)
	}
	response := websocketMessage{
		Channel: "console",
		Token:   "",
		Data:    helpMessage,
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

func (cfg *serverConfig) handleShout(ctx context.Context, client *Client, args []string) (websocketMessage, error) {
	messageData := "Someone shouts very loud"
	if client.IsAuthed {
		messageData = fmt.Sprintf("%s shouts very loud", client.ID)
	}
	message := websocketMessage{
		Channel: "console",
		Token:   "broadcast",
		Data:    messageData,
	}
	cfg.broadcast(message)
	return websocketMessage{}, nil
}

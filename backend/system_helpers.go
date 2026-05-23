package main

import (
	"context"
	"encoding/json"
	"log"
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

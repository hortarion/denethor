package main

import (
	"encoding/json"
	"log"
)

func (cfg *serverConfig) broadcast(message websocketMessage) {
	cfg.ClientsMu.Lock()
	defer cfg.ClientsMu.Unlock()

	byteMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal broadcast")
	}

	for _, client := range cfg.Clients {
		select {
		case client.Outbound <- byteMessage:
		default:
			log.Printf("[%s] Outbound channel full, dropping message", client.ID)
		}
	}
}

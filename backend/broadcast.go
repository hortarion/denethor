package main

import (
	"encoding/json"
	"log"
)

func (cfg *serverConfig) broadcast(message websocketMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	byteMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal broadcast")
	}

	for connID, outbound := range clients {
		select {
		case outbound <- byteMessage:
		default:
			log.Printf("[%s] Outbound channel full, dropping message", connID)
		}
	}
}

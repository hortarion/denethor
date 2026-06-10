package testapp

import (
	"encoding/json"
	"log"
)

type websocketMessage struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
	Data    string `json:"data"`
}

func Main() []byte {
	message := websocketMessage{
		Channel: "app",
		Token:   "",
		Data:    "testApp ran successfully",
	}
	byteMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("[SYS] testApp failed to marshal message")
	}
	return byteMessage
}

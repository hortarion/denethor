package apps

import "fmt"

type websocketMessage struct {
	Channel string `json:"channel"`
	Token   string `json:"token"`
	Data    string `json:"data"`
}

var serverContent = make(map[int]string)

func registerContent(content string) {
	serverContent[len(serverContent)] = content
}

func returnContent() []string {
	content := make([]string, 0)
	for _, name := range serverContent {
		content = append(content, name)
	}
	return content
}

func InternalRegistry() {
	registerContent("rockPaperScissors")
	registerContent("more to come")
	content := returnContent()
	fmt.Println("Registered applications:")
	for idx, item := range content {
		fmt.Println(idx, "-", item)
	}
}

func Apps(token, data string) (websocketMessage, error) {
	var response websocketMessage
	switch token {
	case "launch":
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    "not implemented - this command will start app launcher",
		}
	}
	return response, nil
}

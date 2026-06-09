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

func AppLauncher(_, token, data string) (websocketMessage, string, error) {
	app := "appLauncher"
	var response websocketMessage
	// DEV log
	fmt.Printf("[DEV] APP received token: %s, data: %s\n", token, data)
	switch data {
	case "launch":
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    "not implemented - this command will start app launcher",
		}
	case "back":
		response = websocketMessage{
			Channel: "app",
			Token:   "Console",
			Data:    "returning back to console",
		}
		app = "console"
	case "help":
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    "App Launcher help message not implemented",
		}
	}
	return response, app, nil
}

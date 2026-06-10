package apps

import (
	"fmt"
	"sort"
)

type appCommand struct {
	name        string
	description string
	message     websocketMessage
}

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

func handleHelp(commands map[string]appCommand) string {
	var cmdKeys = make([]string, 0, len(commands))
	for k := range commands {
		cmdKeys = append(cmdKeys, k)
	}
	sort.Strings(cmdKeys)
	var helpMessage string
	for _, key := range cmdKeys {
		helpMessage += fmt.Sprintf("	> %s - %s\n", commands[key].name, commands[key].description)
	}
	return helpMessage
}

func appCommands() map[string]appCommand {
	commands := map[string]appCommand{
		"back": {
			name:        "back",
			description: "Back to console",
			message: websocketMessage{
				Channel: "app",
				Token:   "console",
				Data:    "returning back to console",
			},
		},
		"clear": {
			name:        "clear",
			description: "Clear the screen",
			message: websocketMessage{
				Channel: "sys",
				Token:   "clear",
				Data:    "",
			},
		},
		"help": {
			name:        "help",
			description: "Display available commands",
			message: websocketMessage{
				Channel: "app",
				Token:   "",
				Data:    "",
			},
		},
		"launch": {
			name:        "launch",
			description: "Launches an app",
			message: websocketMessage{
				Channel: "app",
				Token:   "",
				Data:    "not implemented - this command will start app launcher",
			},
		},
	}
	return commands
}

func AppLauncher(_, _, data string) (websocketMessage, string, error) {
	app := "appLauncher"
	var response websocketMessage
	// DEV log
	fmt.Printf("[DEV] APP received data: %s\n", data)
	if data == "help" {
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    handleHelp(appCommands()),
		}
		return response, app, nil
	}
	command, exists := appCommands()[data]
	if exists {
		response = command.message
	} else {
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    "",
		}
	}
	if data == "back" {
		app = "console"
	}
	return response, app, nil
}

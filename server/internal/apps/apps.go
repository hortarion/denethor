package apps

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	testapp "github.com/hortarion/server/internal/apps/testApp.go"
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

type app struct {
	name        string
	description string
	callback    func() []byte
}

var serverContent = make(map[string]app)

var launched = false

func registerContent(name, description string, callback func() []byte) {
	serverContent[name] = app{
		name:        name,
		description: description,
		callback:    callback,
	}
}

func internalRegistry() {
	registerContent("testApp", "for testing", testapp.Main)
	registerContent("rockPaperScissors", "not implemented", func() []byte { return []byte{} })
}

func listRegistry() string {
	// Composes app registry message
	var apps = make([]string, 0, len(serverContent))
	for k := range serverContent {
		apps = append(apps, k)
	}
	sort.Strings(apps)
	var registryMessage string
	for _, key := range apps {
		registryMessage += fmt.Sprintf("    > %s - %s\n", serverContent[key].name, serverContent[key].description)
	}
	return registryMessage
}

func handleHelp(commands map[string]appCommand) string {
	// Composes help message
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

func handleLaunch(appName string) websocketMessage {
	byteResponse := serverContent[appName].callback()
	message := websocketMessage{}
	err := json.Unmarshal(byteResponse, &message)
	if err != nil {
		log.Printf("[SYS] failed to launch app %s", appName)
	}
	return message
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
				Data:    "",
			},
		},
		"list": {
			name:        "list",
			description: "Lists available apps",
			message: websocketMessage{
				Channel: "app",
				Token:   "",
				Data:    "",
			},
		},
	}
	return commands
}

func AppLauncher(_, _, data string) (websocketMessage, string, error) {
	if !launched {
		internalRegistry()
		log.Printf("Registry: %s\n", listRegistry())
		launched = true
	}

	var args []string
	args = strings.Split(data, " ")

	app := "appLauncher"
	var response websocketMessage

	// DEV log
	log.Printf("[DEV] APP received data: %s\n", data)

	// App commands that require specific callbacks
	if args[0] == "help" {
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    handleHelp(appCommands()),
		}
		return response, app, nil
	}
	if args[0] == "launch" {
		if len(args) != 2 {
			response = websocketMessage{
				Channel: "app",
				Token:   "",
				Data:    "to launch an app type: launch <appName>",
			}
			return response, app, nil
		}
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    handleLaunch(args[1]).Data,
		}
		return response, app, nil
	}
	if args[0] == "list" {
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    listRegistry(),
		}
		return response, app, nil
	}

	// Command handling
	command, exists := appCommands()[args[0]]
	if exists {
		response = command.message
	} else {
		response = websocketMessage{
			Channel: "app",
			Token:   "",
			Data:    "",
		}
	}

	// Update client.activeApp
	if args[0] == "back" {
		app = "console"
	}
	return response, app, nil
}

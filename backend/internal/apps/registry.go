package internalRegistry

import "fmt"

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
	registerContent("notImplemented")
	content := returnContent()
	fmt.Println("Registered applications:")
	for _, item := range content {
		fmt.Println(">>", item)
	}
}

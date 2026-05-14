package internal

var serverContent = make(map[int]string)

func RegisterContent(content string) {
	serverContent[len(serverContent)] = content
}

func PrintContent() []string {
	content := make([]string, 0)
	for _, name := range serverContent {
		content = append(content, name)
	}
	return content
}

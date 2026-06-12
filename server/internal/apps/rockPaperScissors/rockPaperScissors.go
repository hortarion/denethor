package main

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
)

func main() {
	choices := []string{"rock", "paper", "scissors"}

	fmt.Println("=== Rock Paper Scissors ===")
	fmt.Println("Enter rock, paper, or scissors to play.")
	fmt.Println("Enter quit to exit.")
	fmt.Println()

	for {
		fmt.Print("Your choice: ")
		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "quit" || input == "q" {
			break
		}

		if !slices.Contains(choices, input) {
			fmt.Println("Invalid choice. Please enter rock, paper, or scissors.")
			continue
		}

		computerChoice := choices[rand.Intn(len(choices))]
		fmt.Printf("You: %s, Computer: %s\n", input, computerChoice)
		fmt.Println(determineWinner(input, computerChoice))
		fmt.Println()
	}
}

func determineWinner(user, computer string) string {
	if user == computer {
		return "Tie!"
	}

	switch user {
	case "rock":
		if computer == "scissors" {
			return "You win! Rock crushes scissors."
		}
		return "You lose! Paper covers rock."
	case "paper":
		if computer == "rock" {
			return "You win! Paper covers rock."
		}
		return "You lose! Scissors cut paper."
	case "scissors":
		if computer == "paper" {
			return "You win! Scissors cut paper."
		}
		return "You lose! Rock crushes scissors."
	}
	return ""
}

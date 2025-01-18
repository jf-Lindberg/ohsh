package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		userInput, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		splitInput := strings.Split(userInput, " ")
		command := strings.Trim(splitInput[0], "\n")

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Print(strings.Join(splitInput[1:], " "))
		default:
			fmt.Println(command + ": command not found")
		}
	}
}

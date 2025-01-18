package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprint(os.Stdout, "$ ")

	// Wait for user input
	var command string
	fmt.Scanln(&command)
	fmt.Printf("%s: command not found", command)
}

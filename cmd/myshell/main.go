package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func findInPath(path string, name string) (string, error) {
	var foundPath string
	err := filepath.WalkDir(path, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Name() == name {
			foundPath = s
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", errors.New("not found")
	}

	return foundPath, nil
}

func main() {
	for {
		fmt.Print("$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return
		}

		input = strings.TrimRight(input, "\n")
		args := strings.Split(input, " ")
		cmd := args[0]
		args = args[1:]

		if cmd == "exit" && len(args) == 1 {
			code, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Print(err)
			}
			os.Exit(code)
		} else if cmd == "echo" && len(args) > 1 {
			fmt.Print(strings.Join(args, " "))
		} else if cmd == "type" && len(args) == 1 {
			switch args[0] {
			case "exit", "echo", "type":
				fmt.Printf("%s is a shell builtin", args[0])
			default:
				path := os.Getenv("PATH")
				splitPath := strings.Split(path, ":")
				found := false
				for _, dir := range splitPath {
					foundPath, err := findInPath(dir, args[0])
					if err == nil {
						fmt.Print(foundPath)
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("%s not found", args[0])
				}
			}
		} else {
			fmt.Printf("%s: command not found", input)
		}

		fmt.Println()
	}

}

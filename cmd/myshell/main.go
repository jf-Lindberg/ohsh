package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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
			fmt.Println()
		} else if cmd == "type" && len(args) == 1 {
			switch args[0] {
			case "exit", "echo", "type", "pwd", "cd":
				fmt.Printf("%s is a shell builtin", args[0])
				fmt.Println()
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
				fmt.Println()
			}
		} else if cmd == "pwd" {
			pwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Print(pwd)
			fmt.Println()
		} else if cmd == "cd" {
			err := os.Chdir(args[0])
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory", args[0])
				fmt.Println()
			}
		} else {
			command := exec.Command(cmd, args...)
			out, err := command.Output()
			if err != nil {
				fmt.Printf("%s: command not found", input)
			} else {
				fmt.Print(strings.Trim(string(out), "\n"))
			}
			fmt.Println()
		}
	}
}

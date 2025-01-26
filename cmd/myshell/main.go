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

		const (
			StateNormal = iota
			StateSingleQuote
			StateDoubleQuote
			StateEscaped
		)

		var words []string
		start := 0
		state := StateNormal
		previousState := StateNormal
		collected := ""
		i := 0
		for i < len(input) {
			r := input[i]
			switch state {
			case StateNormal:
				switch r {
				case '\'':
					start = i + 1
					state = StateSingleQuote
				case '"':
					start = i + 1
					state = StateDoubleQuote
				case '\\':
					state = StateEscaped
					previousState = StateNormal
				case ' ', '\n':
					word := collected + input[start:i]
					if len(strings.TrimSpace(word)) != 0 {
						words = append(words, word)
					}
					start = i + 1
					collected = ""
				}
				i++

			case StateSingleQuote:
				if r == '\'' {
					collected = collected + input[start:i]
					start = i + 1
					state = StateNormal
				}
				i++

			case StateDoubleQuote:
				if r == '\\' {
					state = StateEscaped
					previousState = StateDoubleQuote
				}
				if r == '"' {
					collected = collected + input[start:i]
					start = i + 1
					state = StateNormal
				}
				i++

			case StateEscaped:
				if r == 'n' {
					if previousState == StateNormal {
						collected = collected + input[start:i-1] + "n"
					} else {
						collected = collected + input[start:i-1] + "\\n"
					}
					start = i + 1
				}
				if r == '\'' {
					if previousState == StateNormal {
						collected = collected + input[start:i-1]
						start = i
						i++
					}
				}
				if r == ' ' || r == '\\' || r == '"' || r == '$' {
					collected = collected + input[start:i-1]
					start = i
					i++
				}
				state = previousState
			}
		}

		// Add final word
		if start < len(input) {
			words = append(words, input[start:])
		}

		var args []string
		cmd := strings.ToLower(words[0])
		if len(words) > 1 {
			args = words[1:]
		}

		if cmd == "exit" {
			code := 0
			var err error
			if len(args) == 1 {
				code, err = strconv.Atoi(args[0])
				if err != nil {
					code = 0
				}
			}
			os.Exit(code)
		} else if cmd == "echo" && len(args) > 0 {
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
			err = nil
			dir := args[0]
			if dir == "~" {
				home := os.Getenv("HOME")
				err = os.Chdir(home)
			} else {
				err = os.Chdir(dir)
			}
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory", args[0])
				fmt.Println()
			}
		} else {
			escapedArgs := make([]string, len(args))
			for i, arg := range args {
				escapedArgs[i] = strings.ReplaceAll(arg, "\n", "\\n")
			}
			command := exec.Command(cmd, escapedArgs...)
			out, err := command.Output()
			if err != nil {
				if cmd == "cat" {
					fmt.Println(args)
					fmt.Println(escapedArgs)
				}
				fmt.Printf("%s: command not found", cmd)
			} else {
				fmt.Print(strings.Trim(string(out), "\n"))
			}
			fmt.Println()
		}
	}
}

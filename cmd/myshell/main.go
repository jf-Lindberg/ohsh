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

		var args []string
		var escaped bool
		var inSingleQuote bool
		var inDoubleQuote bool
		var sb strings.Builder
		var ab []string
		for i := 0; i < len(input); i++ {
			if !inSingleQuote && !inDoubleQuote {
				if !escaped && (input[i] == ' ' || input[i] == '\n') {
					if sb.Len() > 0 {
						ab = append(ab, sb.String())
						sb.Reset()
					}
					continue
				}

				if input[i] == '"' && !escaped {
					inDoubleQuote = true
					continue
				}

				if input[i] == '\'' && !escaped {
					inSingleQuote = true
					continue
				}

				if input[i] == '\\' && !escaped {
					escaped = true
					continue
				}

				char := fmt.Sprintf("%c", input[i])
				sb.WriteString(char)
				escaped = false
			}

			if inSingleQuote {
				if input[i] == '\'' {
					if i+1 < len(input) && input[i+1] == '\'' {
						inSingleQuote = false
						continue
					}
					ab = append(ab, sb.String())
					sb.Reset()
					inSingleQuote = false
					continue
				}

				char := fmt.Sprintf("%c", input[i])
				sb.WriteString(char)
			}

			if inDoubleQuote {
				if input[i] == '\\' {
					switch input[i+1] {
					case '\\', '$', '"', '\n':
						continue
					}
				}

				if input[i] == '"' {
					if i+1 < len(input) && input[i+1] == '"' {
						inDoubleQuote = false
						continue
					}
					ab = append(ab, sb.String())
					sb.Reset()
					inDoubleQuote = false
					continue
				}

				char := fmt.Sprintf("%c", input[i])
				sb.WriteString(char)
			}
		}

		cmd := strings.ToLower(ab[0])
		if len(ab) > 1 {
			args = ab[1:]
		}

		if cmd == "exit" && len(args) == 1 {
			code, err := strconv.Atoi(args[0])
			if err != nil {
				code = 0
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
			command := exec.Command(cmd, args...)
			out, err := command.Output()
			if err != nil {
				fmt.Printf("%s: command not found", cmd)
			} else {
				fmt.Print(strings.Trim(string(out), "\n"))
			}
			fmt.Println()
		}
	}
}

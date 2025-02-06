package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"sort"
	"strings"
)

var autocompleted bool
var oldState *term.State

func longestCommonPrefix(strs []string) string {
	var longestPrefix string = ""
	var endPrefix = false

	if len(strs) > 0 {
		sort.Strings(strs)
		first := string(strs[0])
		last := string(strs[len(strs)-1])

		for i := 0; i < len(first); i++ {
			if !endPrefix && string(last[i]) == string(first[i]) {
				longestPrefix += string(last[i])
			} else {
				endPrefix = true
			}
		}
	}
	return longestPrefix
}

func getExternals() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	seen := make(map[string]bool)
	var executables []string

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.Mode()&0111 != 0 {
				if !seen[entry.Name()] {
					seen[entry.Name()] = true
					executables = append(executables, entry.Name())
				}
			}
		}
	}

	sort.Strings(executables)
	return executables, nil
}

func readInput(builtins []string, externals []string) (string, error) {
	var input []rune
	var pos int
	for {
		char := make([]byte, 1)
		_, err := os.Stdin.Read(char)
		if err != nil {
			return "", err
		}

		switch char[0] {
		case 3: // Ctrl-C
			fmt.Print("^C\n\r$ ")
			input = input[:0]
			pos = 0
			// should exit
			continue
		case 4: // Ctrl-D
			if len(input) == 0 {
				return "", nil
			}
		case 9: // Tab
			word := string(input)
			var autocompletions []string
			found := false
			for i := range builtins {
				fullCmd := builtins[i]
				if strings.HasPrefix(fullCmd, word) {
					input = append(append(input, []rune(fullCmd[len(word):])...), ' ')
					fmt.Print(builtins[i][len(word):] + " ")
					pos += len(fullCmd) - len(word) + 1 // +1 for the space
					found = true
					break
				}
			}
			if !found {
				for i := range externals {
					fullCmd := externals[i]
					if strings.HasPrefix(fullCmd, word) {
						autocompletions = append(autocompletions, fullCmd)
					}
				}
				if len(autocompletions) > 1 {
					prefix := longestCommonPrefix(autocompletions)
					if prefix == "" || prefix != word || len(prefix) > len(word) {
						input = append(append(input, []rune(prefix[len(word):])...))
						fmt.Print(prefix[len(word):])
						pos += len(prefix) - len(word)
						break
					}

					if !autocompleted {
						fmt.Print("\a")
						autocompleted = true
						break
					}

					fmt.Print("\n\r")
					for i := range autocompletions {
						if i > 0 {
							fmt.Print("  ")
						}
						fmt.Print(autocompletions[i])
					}
					fmt.Print("\n")
					fmt.Print("\r$ ")
					fmt.Print(string(input))
					pos = len(input)
				}
				if len(autocompletions) == 1 {
					input = append(append(input, []rune(autocompletions[0])...), ' ')
					fmt.Print(autocompletions[0][len(word):] + " ")
					pos += len(autocompletions[0]) - len(word) + 1 // +1 for the space
				}
				if len(autocompletions) == 0 {
					fmt.Print("\a")
				}
			}

		case 13, 10: // Enter
			fmt.Print("\r\n")
			return string(input), nil
		case 127:
			if pos > 0 {
				input = append(input[:pos-1], input[pos:]...)
				pos--
				fmt.Print("\b \b")
			}
		default:
			if char[0] >= 32 { // Printable characters
				// Insert character at current position
				input = append(input[:pos], append([]rune{rune(char[0])}, input[pos:]...)...)
				pos++
				fmt.Printf("%c", char[0])
			}
		}
	}
}

func main() {
	var err error
	oldState, err = term.MakeRaw(int(os.Stdin.Fd()))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting raw mode: %v\n", err)
		os.Exit(1)
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Print initial prompt
	fmt.Print("\r$ ")
	// shell = newShell("ohsh")
	// shell.start()
	// shell should have:
	// 	- parser
	//	- executor
	//	- output handler
	//	- maybe environment (handling dirs etc)
	// switch stdin into 'raw' mode
	for {
		builtins := []string{"type", "echo", "exit", "pwd"}
		externals, err := getExternals() // can probably just append builtins to PATH?
		if err != nil {
			fmt.Println("get externals error")
			os.Exit(1)
		}
		shell := newShell("ohsh")
		parser := newParser()

		var input string
		input, err = readInput(builtins, externals)
		if input == "" {
			fmt.Print("\r$ ")
			continue
		}

		input = input + "\n" // remove this dependency

		argIsRedirectFile := false
		redirectFileName := ""
		shouldAppend := false
		var redirectFile *os.File
		for shell.position < len(input) {
			r := input[shell.position]
			switch shell.state {
			case Normal:
				switch r {
				case '\'', '"':
					shell.offsetTokenStartAgainstPosition(shell.enterQuotedString(r))
				case '\\':
					shell.enterEscaped()
				case '>':
					shouldAppend = shell.position+1 < len(input) && input[shell.position+1] == '>'
					validRedirect := shell.handleRedirect(input[shell.position-1])
					parser.addToCollected(input[shell.tokenStart:shell.position])
					if validRedirect {
						collected := parser.collected
						parser.resetCollected()
						parser.addToCollected(collected[:len(collected)-1])
					}

					if strings.TrimSpace(parser.collected) != "" {
						parser.appendToken(parser.collected)
						parser.resetCollected()
					}

					argIsRedirectFile = true
					shell.incrementPosition()
					if shouldAppend {
						shell.incrementPosition()
					}
					shell.setTokenStartToPosition()
				case ' ', '\n':
					token := strings.TrimSpace(parser.collected + input[shell.tokenStart:shell.position])
					if len(token) != 0 {
						if argIsRedirectFile {
							redirectFileName = token
							if shouldAppend {
								redirectFile, err = os.OpenFile(redirectFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
							} else {
								redirectFile, err = os.Create(redirectFileName)
							}
							if err != nil {
								fmt.Println(err)
							}
							argIsRedirectFile = false
						} else {
							parser.appendToken(token)
						}
					}
					shell.offsetTokenStartAgainstPosition(1)
					parser.resetCollected()
				}
				shell.incrementPosition()

			case SingleQuote:
				if r == '\'' {
					parser.addToCollected(input[shell.tokenStart:shell.position])
					shell.offsetTokenStartAgainstPosition(shell.exitQuotedString())
				}
				shell.incrementPosition()

			case DoubleQuote:
				if r == '\\' {
					shell.enterEscaped()
				}
				if r == '"' {
					parser.addToCollected(input[shell.tokenStart:shell.position])
					shell.offsetTokenStartAgainstPosition(shell.exitQuotedString())
				}
				shell.incrementPosition()

			case Escaped:
				addition, offset := shell.handleEscapedChar(r)
				parser.addToCollected(input[shell.tokenStart:shell.position-1] + addition)
				shell.offsetTokenStartAgainstPosition(offset)
				if offset == 0 {
					// All chars with offset 0 should be skipped
					shell.incrementPosition()
				}
				shell.exitEscaped()
			}
		}

		if shell.tokenStart < len(input) {
			parser.appendToken(input[shell.tokenStart:])
		}

		var args []string
		var tokens = parser.tokens
		cmd := strings.ToLower(tokens[0])
		if len(tokens) > 1 {
			args = tokens[1:]
		}

		term.Restore(int(os.Stdin.Fd()), oldState)
		output, errOutput, err := shell.handleCommand(cmd, args)

		if shell.redirectTo == StdOut {
			if errOutput != "" {
				fmt.Print(errOutput)
			}

			_, err := redirectFile.WriteString(output)
			if err != nil {
				return
			}
		} else if shell.redirectTo == StdErr {
			if output != "" {
				fmt.Print(output)
			}

			_, err := redirectFile.WriteString(errOutput)
			if err != nil {
				return
			}
		} else {
			if errOutput != "" {
				fmt.Printf("\r%s", errOutput)
			}
			if output != "" {
				fmt.Printf("\r%s", output)
			}
		}
		oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
		fmt.Print("\r$ ")
	}
}

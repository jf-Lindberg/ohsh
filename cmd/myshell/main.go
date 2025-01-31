package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		shell := newShell("ohsh")
		parser := newParser()

		fmt.Print("$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return
		}

		argIsRedirectFile := false
		redirectFileName := ""
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
					argIsRedirectFile = true
					shell.incrementPosition()
					shell.setTokenStartToPosition()
				case ' ', '\n':
					token := strings.TrimSpace(parser.collected + input[shell.tokenStart:shell.position])
					if len(token) != 0 {
						if argIsRedirectFile {
							redirectFileName = token
							redirectFile, err = os.Create(redirectFileName)
							if err != nil {
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

		output, errOutput, err := shell.handleCommand(cmd, args)

		if errOutput != "" {
			fmt.Print(errOutput)
		}

		if redirectFile != nil && output != "" {
			_, err := redirectFile.WriteString(output)
			if err != nil {
				return
			}
		} else if output != "" {
			fmt.Print(output)
		}
	}
}

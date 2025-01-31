package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ShellState int

const (
	Normal ShellState = iota
	SingleQuote
	DoubleQuote
	Escaped
)

// Eventually the shell should have a "start" function, beginning the main REPL
type shell struct {
	name          string
	state         ShellState
	previousState ShellState
	position      int
	tokenStart    int
}

type query struct {
	redirectFile *os.File
}

func newShell(name string) *shell {
	return &shell{
		name:          name,
		state:         Normal,
		previousState: Normal,
		position:      0,
		tokenStart:    0,
	}
}

func (s *shell) enterEscaped() {
	s.previousState = s.state
	s.state = Escaped
}

func (s *shell) exitEscaped() {
	s.state = s.previousState
}

func (s *shell) enterQuotedString(quote byte) int {
	if quote == '\'' {
		s.state = SingleQuote
	} else {
		s.state = DoubleQuote
	}

	return 1 // start position offset
}

func (s *shell) exitQuotedString() int {
	s.state = Normal
	return 1
}

func (s *shell) findInPath(path string, name string) (string, error) {
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

func (s *shell) handleEscapedChar(r byte) (string, int) {
	switch r {
	case 'n':
		if s.previousState == Normal {
			return "n", 1
		}
		return "\\n", 1
	case '\'':
		if s.previousState == Normal {
			return "", 0
		}
	case ' ', '\\', '"', '$':
		return "", 0
	}

	return "", -1
}

func (s *shell) incrementPosition() {
	s.position++
}

func (s *shell) offsetTokenStartAgainstPosition(n int) {
	s.tokenStart = s.position + n
}

func (s *shell) setTokenStartToPosition() {
	s.offsetTokenStartAgainstPosition(0)
}

// Commands
func (s *shell) handleCommand(command string, args []string) (string, string, error) {
	switch command {
	case "exit":
		s.handleExit(args[0])
		return "", "", nil
	case "echo":
		return s.handleEcho(args), "", nil
	case "type":
		return s.handleType(args[0]), "", nil
	case "pwd":
		return s.handlePwd(), "", nil
	case "cd":
		return s.handleCd(args[0]), "", nil
	default:
		return s.handleExternalCommand(command, args)
	}
}

func (s *shell) handleCd(directory string) string {
	if directory == "~" {
		directory = os.Getenv("HOME")
	}

	if err := os.Chdir(directory); err != nil {
		return fmt.Sprintf("cd: %s: No such file or directory\n", directory)
	}

	return ""
}

func (s *shell) handleExit(codeArg string) {
	code, err := strconv.Atoi(codeArg)
	if err != nil {
		code = 0
	}
	os.Exit(code)
}

func (s *shell) handleEcho(echoables []string) string {
	echo := strings.Join(echoables, " ") + "\n"
	return echo
}

func (s *shell) handleType(command string) string {
	switch command {
	case "exit", "echo", "type", "pwd", "cd":
		builtinInfo := fmt.Sprintf("%s is a shell builtin\n", command)
		return builtinInfo
	default:
		path := os.Getenv("PATH")
		splitPath := strings.Split(path, ":")
		for _, dir := range splitPath {
			foundPath, err := s.findInPath(dir, command)
			if err == nil {
				return fmt.Sprintf("%s is %s\n", command, foundPath)
			}
		}
		return fmt.Sprintf("%s not found\n", command)
	}
}

func (s *shell) handlePwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return pwd + "\n"
}

func (s *shell) handleExternalCommand(command string, args []string) (string, string, error) {
	cmd := exec.Command(command, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()

	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return "", fmt.Sprintf("%s: command not found\n", command), errors.New("command not found")
		}
		return string(out), stderr.String(), err
	}

	return string(out), "", nil
}

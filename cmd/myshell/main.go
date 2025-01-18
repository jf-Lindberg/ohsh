package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Builtin interface {
	Name() string
	Run(args []string) error
}

type Exit struct{}

func (e Exit) Name() string {
	return "exit"
}

func (e Exit) Run(args []string) error {
	os.Exit(0)
	return nil
}

type Echo struct{}

func (e Echo) Name() string {
	return "echo"
}

func (e Echo) Run(args []string) error {
	fmt.Print(strings.Join(args, " ") + "\n")
	return nil
}

type Type struct {
	shell *Shell
}

func (t Type) Name() string {
	return "type"
}

func (t Type) Run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("type: usage: type command_name")
	}

	commandName := args[0]
	if _, ok := t.shell.builtins[commandName]; ok {
		fmt.Printf("%s is a shell builtin\n", commandName)
		return nil
	}

	fmt.Printf("%s: not found\n", commandName)
	return nil
}

type Shell struct {
	builtins map[string]Builtin
}

func NewShell() *Shell {
	s := &Shell{
		builtins: make(map[string]Builtin),
	}
	s.registerBuiltins()
	return s
}

func (s *Shell) registerBuiltins() {
	builtins := []Builtin{
		Exit{},
		Echo{},
		Type{shell: s},
	}

	for _, builtin := range builtins {
		s.builtins[builtin.Name()] = builtin
	}
}

func (s *Shell) Execute(command string, args []string) error {
	if builtin, ok := s.builtins[command]; ok {
		return builtin.Run(args)
	}
	fmt.Printf("%s: command not found\n", command)
	return nil
}

func main() {
	shell := NewShell()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		fields := strings.Fields(strings.TrimSpace(userInput))
		if len(fields) == 0 {
			continue
		}

		command := fields[0]
		args := fields[1:]

		if err := shell.Execute(command, args); err != nil {
			fmt.Println(err)
		}
	}
}

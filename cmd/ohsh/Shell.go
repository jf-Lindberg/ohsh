package ohsh

import (
	"bufio"
	"fmt"
	"os"
)

type Shell struct {
	name        string
	parser      *Parser
	interpreter *Interpreter
}

func NewShell(name string) *Shell {
	return &Shell{
		name:        name,
		parser:      NewParser(),
		interpreter: NewInterpreter(),
	}
}

func (s *Shell) Start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if err := s.ProcessLine(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func (s *Shell) ProcessLine(line string) error {
	s.parser.Reset()
	s.interpreter.Reset()

	command, err := s.parser.ParseLine(line)
	if err != nil {
		return err
	}

	return s.interpreter.ExecuteCommand(command)
}

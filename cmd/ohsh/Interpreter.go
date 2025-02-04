package ohsh

type Interpreter struct {
	redirector *Redirector
}

func NewInterpreter() *Interpreter {
	redirector := NewRedirector()
	return &Interpreter{
		redirector,
	}
}

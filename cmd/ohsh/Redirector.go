package ohsh

import "os"

type StreamType int
type RedirectMode int

const (
	StdOut StreamType = iota
	StdErr
)

const (
	Standby RedirectMode = iota
	Redirect
	Append
)

type Redirector struct {
	stream   StreamType
	fileName string
	mode     RedirectMode
	file     *os.File
}

func NewRedirector() *Redirector {
	return &Redirector{
		stream:   StdOut,
		fileName: "",
		mode:     Standby,
		file:     nil,
	}
}

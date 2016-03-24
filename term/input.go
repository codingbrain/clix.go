package term

import (
	"io"

	"golang.org/x/crypto/ssh/terminal"
)

type Input struct {
	term  *terminal.Terminal
	fd    int
	state *terminal.State
}

func NewInput(rw io.ReadWriter, fd uintptr) (*Input, error) {
	if s, err := terminal.MakeRaw(int(fd)); err != nil {
		return nil, err
	} else {
		return &Input{
			term:  terminal.NewTerminal(rw, ""),
			fd:    int(fd),
			state: s,
		}, nil
	}
}

func (i *Input) Close() error {
	return terminal.Restore(i.fd, i.state)
}

func (i *Input) Prompt(prompt string) *Input {
	i.term.SetPrompt(prompt)
	return i
}

func (i *Input) ReadLine() (string, error) {
	return i.term.ReadLine()
}

func (i *Input) ReadPassword(prompt string) (string, error) {
	return i.term.ReadPassword(prompt)
}

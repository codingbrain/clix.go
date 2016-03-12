package args

import (
	"os"

	"github.com/codingbrain/clix.go/clix"
)

type ArgsEmitter struct {
	args   []string
	parser clix.Receptor
}

func NewArgsWith(args []string) *ArgsEmitter {
	return &ArgsEmitter{args: args}
}

func NewArgs() *ArgsEmitter {
	return NewArgsWith(os.Args)
}

func (e *ArgsEmitter) EmitTo(receptor clix.Receptor) {
	e.parser = receptor
}

func (e *ArgsEmitter) Parse() error {
	for _, arg := range e.args {
		if err := e.parser.Consume(arg); err != nil {
			return err
		}
	}
	return e.parser.End()
}

package args

import (
	"os"
)

type CliDef struct {
	Cli *Command `yaml:"cli,omitempty"`
}

func (def *CliDef) Normalize() error {
	if def.Cli != nil {
		return def.Cli.Normalize()
	}
	return nil
}

func (def *CliDef) ParseArgs(args []string) error {
	r := def.Cli.Parser().ParseArgs(args)
	// TODO handle result
	return r.Error
}

func (def *CliDef) Parse() error {
	r := def.Cli.Parser().ParseArgs(os.Args)
	// TODO handle result
	return r.Error
}

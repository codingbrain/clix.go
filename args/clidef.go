package args

// CliDef is the top-level definition of command-line
type CliDef struct {
	Cli *Command `yaml:"cli,omitempty"`
}

// Normalize normalizes the parsed cli definition
func (def *CliDef) Normalize() error {
	if def.Cli != nil {
		return def.Cli.Normalize()
	}
	return nil
}

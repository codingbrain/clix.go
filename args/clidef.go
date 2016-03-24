package args

// CliDef is the top-level definition of command-line
type CliDef struct {
	Cli *Command `yaml:"cli,omitempty"`

	exts []ExtRegistrar
}

// Normalize normalizes the parsed cli definition
func (d *CliDef) Normalize() error {
	if d.Cli != nil {
		return d.Cli.Normalize()
	}
	return nil
}

func (d *CliDef) Use(extRegs ...ExtRegistrar) *CliDef {
	d.exts = append(d.exts, extRegs...)
	return d
}

func (d *CliDef) Parser() *Parser {
	return d.Cli.Parser().Use(d.exts...)
}

func (d *CliDef) ParseArgs(args ...string) *ParseResult {
	return d.Parser().ParseArgs(args...)
}

func (d *CliDef) Parse() *ParseResult {
	return d.Parser().Parse()
}

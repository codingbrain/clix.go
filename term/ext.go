package term

import "github.com/codingbrain/clix.go/flag"

const (
	// DefaultColorOptName is the default option name
	DefaultColorOptName = "color"
)

// TermExt is extenstion for flag
type TermExt struct {
	// ColorOptName is name of color option, default is "color"
	ColorOptName string
	// ColorOptAlias is alias list of color option
	ColorOptAlias []string
}

// NewExt creates the extension
func NewExt() *TermExt {
	return &TermExt{ColorOptName: DefaultColorOptName}
}

// ColorOpt overrides the default option name and aliases
func (x *TermExt) ColorOpt(name string, alias ...string) *TermExt {
	x.ColorOptName = name
	x.ColorOptAlias = alias
	return x
}

// RegisterExt implements ExtRegistrar
func (x *TermExt) RegisterExt(parser *flag.Parser) {
	parser.AddExecExt(x)
}

// ExecuteCmd implements execution extension
func (x *TermExt) ExecuteCmd(ctx *flag.ExecContext) {
	name := x.ColorOptName
	if name != "" {
		x.applyColorOpt(name, ctx)
	}
}

func (x *TermExt) applyColorOpt(name string, ctx *flag.ExecContext) {
	l := len(ctx.Result.CmdStack)
	for i := l - 1; i >= 0; i-- {
		parsedCmd := ctx.CmdAt(i)
		if _, exist := parsedCmd.Opts[name]; !exist {
			continue
		}
		val, ok := parsedCmd.Vars[name].(bool)
		if ok {
			Std.Color = val
			break
		}
	}
}

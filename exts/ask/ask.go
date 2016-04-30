package ask

import (
	"bytes"
	"io"

	"github.com/codingbrain/clix.go/exts/help"
	"github.com/codingbrain/clix.go/flag"
	"github.com/codingbrain/clix.go/term"
)

// AskExt defines the ask extension which must be hooked up to
// - Execution
type AskExt struct {
	Terminal *term.Terminal
}

// NewExt creates ask extension
func NewExt() *AskExt {
	return &AskExt{
		Terminal: term.Std,
	}
}

// UseTerminal explicitly specifies the terminal
func (x *AskExt) UseTerminal(t *term.Terminal) *AskExt {
	x.Terminal = t
	return x
}

// ExecuteCmd implements execution extension
func (x *AskExt) ExecuteCmd(ctx *flag.ExecContext) {
	t := x.Terminal
	if t == nil {
		t = term.Std
	}
	if ctx.Result.Error != nil || ctx.Result.MissingCmd || !t.IsInTTY() {
		return
	}
	for _, pcmd := range ctx.Result.CmdStack {
		if len(pcmd.Errs) == 0 {
			continue
		}
		errs := make([]*flag.VarError, 0, len(pcmd.Errs))
		for _, err := range pcmd.Errs {
			var msg string
			switch err.ErrType {
			case flag.VarErrNoVal:
				msg = "expects a value"
			case flag.VarErrBadVal:
				msg = "has an invalid value: " + *err.Value
			default:
				errs = append(errs, err)
				continue
			}
			printer := term.NewPrinter(t).Styles(term.StyleErr)
			if err.Def.IsArg {
				printer.
					Print("Argument ").
					Styles(term.StyleB, term.StyleHi).Print(help.ArgDisplayName(err.Def) + " ").Pop().
					Println(msg)
			} else {
				printer.
					Print("Option ").
					Styles(term.StyleB, term.StyleHi).Print(help.OptionDisplayName(err.Def) + " ").Pop().
					Println(msg)
			}
			if err.Def.Desc != "" {
				printer.Reset().Println().Println(err.Def.Desc).Println()
			}

			var prompt bytes.Buffer
			term.NewPrinter(&prompt).
				Styles(term.StyleAsk).Print("Enter ").
				Styles(term.StyleB).Print(help.OptionDisplayName(err.Def)).Pop().
				Print(": ")

			for {
				if line, e := t.ReadLine(prompt.String()); e != nil {
					errs = append(errs, err)
					if e == io.EOF {
						printer.Println()
					}
					break
				} else if line == "" {
					continue
				} else if _, e = pcmd.Assign(err.Def, line); e != nil {
					t.Errorln(e.Error())
				} else {
					break
				}
			}
		}
		pcmd.Errs = errs
	}
}

// RegisterExt implements ExtRegistrar
func (x *AskExt) RegisterExt(parser *flag.Parser) {
	parser.AddExecExt(x)
}

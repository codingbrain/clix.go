package help

import (
	"io"
	"strings"

	"github.com/codingbrain/clix.go/args"
	"github.com/codingbrain/clix.go/term"
)

type DefaultRender struct {
	Output io.Writer
	Plain  bool // no styles
}

type twoColRow struct {
	col []string
}

type twoColRender struct {
	rows []*twoColRow
}

func (r *twoColRender) render(printer *term.Printer) {
	maxlen := 0
	for _, row := range r.rows {
		if l := len(row.col[0]); l > maxlen {
			maxlen = l
		}
	}
	padding := " "
	for i := 0; i < maxlen; i++ {
		padding += " "
	}
	for _, row := range r.rows {
		col0 := row.col[0]
		spaces := padding
		example := false
		if col0 == "" { // example
			example = true
			col0 = "\n  .E.g "
			spaces = "       "
		} else {
			for j := len(col0); j < maxlen; j++ {
				col0 += " "
			}
			col0 += " "
		}
		lines := strings.Split(row.col[1], "\n")
		for l, line := range lines {
			noNewline := strings.TrimSpace(line)
			printer.Print("  ")
			if l == 0 {
				if !example {
					printer.Styles(term.StyleB)
				}
				printer.Print(col0).Reset().Println(noNewline)
			} else {
				printer.Println(spaces + noNewline)
			}
		}
		if example {
			printer.Println()
		}
	}
}

func (r *DefaultRender) RenderStart() {
}

func (r *DefaultRender) RenderComplete() {
}

func (r *DefaultRender) RenderBanner(info *BannerInfo) {
	for _, line := range info.Banner {
		r.printer().Println(line)
	}
}

func (r *DefaultRender) RenderUsage(info *UsageInfo) {
	printer := r.printer()
	printer.
		Styles(term.StyleHi, term.StyleI).Print("Usage").Reset().Print(": ").
		Styles(term.StyleB).Print(strings.Join(info.Cmds, " ")).Reset()
	strs := append([]string{}, info.Opts...)
	strs = append(strs, info.Args...)
	strs = append(strs, info.Tail...)
	printer.Print(" " + strings.Join(strs, " ")).Println().Println()
}

func (r *DefaultRender) RenderCommands(cmds []*args.Command) {
	cr := &twoColRender{}
	for _, cmd := range cmds {
		name := strings.Join(append([]string{cmd.Name}, cmd.Alias...), "|")
		cr.rows = append(cr.rows, &twoColRow{col: []string{name, cmd.Desc}})
		if cmd.Example != "" {
			cr.rows = append(cr.rows, &twoColRow{col: []string{"", cmd.Example}})
		}
	}
	printer := r.printer()
	printer.Styles(term.StyleHi, term.StyleI).Print("Commands").Reset().Println(":")
	cr.render(printer)
	printer.Println()
}

func (r *DefaultRender) RenderArguments(opts []*args.Option) {
	cr := &twoColRender{}
	for _, opt := range opts {
		cr.rows = append(cr.rows, &twoColRow{col: []string{ArgDisplayName(opt), opt.Desc}})
		if opt.Example != "" {
			cr.rows = append(cr.rows, &twoColRow{col: []string{"", opt.Example}})
		}
	}
	printer := r.printer()
	printer.Styles(term.StyleHi, term.StyleI).Print("Arguments").Reset().Println(":")
	cr.render(printer)
	printer.Println()
}

func (r *DefaultRender) RenderOptions(opts []*args.Option) {
	cr := &twoColRender{}
	for _, opt := range opts {
		var short, long []string
		for _, a := range opt.Alias {
			if len(a) == 1 {
				short = append(short, "-"+a)
			} else {
				long = append(long, "--"+a)
			}
		}
		row := &twoColRow{col: make([]string, 2)}
		if len(opt.Name) == 1 {
			short = append(short, "-"+opt.Name)
			row.col[0] = strings.Join(short, "|")
			if opt.ExpectValue() {
				row.col[0] += " " + OptVarName(opt)
			}
		} else {
			long = append(long, "--"+opt.Name)
			if len(short) > 0 {
				row.col[0] = strings.Join(short, "|") + ","
			}
			row.col[0] += strings.Join(long, "|")
			if opt.ExpectValue() {
				row.col[0] += "=" + OptVarName(opt)
			}
		}
		row.col[1] = opt.Desc
		cr.rows = append(cr.rows, row)
		if opt.Example != "" {
			cr.rows = append(cr.rows, &twoColRow{col: []string{"", opt.Example}})
		}
	}
	printer := r.printer()
	printer.Styles(term.StyleHi, term.StyleI).Print("Options").Reset().Println(":")
	cr.render(printer)
	printer.Println()
}

func (r *DefaultRender) RenderErrors(errs []*ErrInfo) {
	for _, err := range errs {
		r.printer().Styles(term.StyleErr).Println("ERROR: " + err.Msg).Reset()
	}
}

func (r *DefaultRender) printer() *term.Printer {
	if r.Output != nil {
		return term.NewPrinter(r.Output)
	}
	return term.NewPrinter(term.Std)
}

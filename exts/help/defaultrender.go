package help

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codingbrain/clix.go/args"
)

type DefaultRender struct {
	Output io.Writer
}

type twoColRow struct {
	col []string
}

type twoColRender struct {
	rows []*twoColRow
}

func (r *twoColRender) render(out io.Writer) {
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
			if l == 0 {
				fmt.Fprintln(out, "    "+col0+noNewline)
			} else {
				fmt.Fprintln(out, "    "+spaces+noNewline)
			}
		}
		if example {
			fmt.Fprintln(out, "")
		}
	}
}

func (r *DefaultRender) RenderBanner(info *BannerInfo) {
	for _, line := range info.Banner {
		r.println(line)
	}
}

func (r *DefaultRender) RenderUsage(info *UsageInfo) {
	out := append([]string{"Usage:"}, info.Cmds...)
	out = append(out, info.Opts...)
	out = append(out, info.Args...)
	out = append(out, info.Tail...)
	r.println(strings.Join(out, " "))
	r.println("")
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
	r.println("Commands:")
	cr.render(r.out())
	r.println("")
}

func (r *DefaultRender) RenderArguments(opts []*args.Option) {
	cr := &twoColRender{}
	for _, opt := range opts {
		cr.rows = append(cr.rows, &twoColRow{col: []string{opt.ValueName(), opt.Desc}})
		if opt.Example != "" {
			cr.rows = append(cr.rows, &twoColRow{col: []string{"", opt.Example}})
		}
	}
	r.println("Arguments:")
	cr.render(r.out())
	r.println("")
}

func (r *DefaultRender) RenderOptions(opts []*args.Option) {
	cr := &twoColRender{}
	for _, opt := range opts {
		short := make([]string, 0)
		long := make([]string, 0)
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
				row.col[0] += " " + strings.ToUpper(opt.ValueName())
			}
		} else {
			long = append(long, "--"+opt.Name)
			if len(short) > 0 {
				row.col[0] = strings.Join(short, "|") + ","
			}
			row.col[0] += strings.Join(long, "|")
			if opt.ExpectValue() {
				row.col[0] += "=" + strings.ToUpper(opt.ValueName())
			}
		}
		row.col[1] = opt.Desc
		cr.rows = append(cr.rows, row)
		if opt.Example != "" {
			cr.rows = append(cr.rows, &twoColRow{col: []string{"", opt.Example}})
		}
	}
	r.println("Options:")
	cr.render(r.out())
	r.println("")
}

func (r *DefaultRender) RenderErrors(errs []*ErrInfo) {
	for _, err := range errs {
		r.println("ERROR: " + err.Msg)
	}
}

func (r *DefaultRender) out() io.Writer {
	if r.Output != nil {
		return r.Output
	}
	return os.Stderr
}

func (r *DefaultRender) println(msg string) {
	fmt.Fprintln(r.out(), msg)
}

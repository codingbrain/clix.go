package term

import (
	"fmt"
	"io"
)

type Printer struct {
	Out     io.Writer
	Palette *Palette

	styles [][]string
}

func NewPrinter(out io.Writer) *Printer {
	return &Printer{Out: out, Palette: &DefaultPalette}
}

func (p *Printer) Styles(names ...string) *Printer {
	p.styles = append(p.styles, names)
	return p
}

func (p *Printer) Pop() *Printer {
	if len(p.styles) > 0 {
		p.styles = p.styles[0 : len(p.styles)-1]
	}
	return p
}

func (p *Printer) Reset() *Printer {
	p.styles = [][]string{}
	return p
}

func (p *Printer) Print(args ...interface{}) *Printer {
	fmt.Fprint(p, args...)
	return p
}

func (p *Printer) Println(args ...interface{}) *Printer {
	fmt.Fprintln(p, args...)
	return p
}

func (p *Printer) Printf(format string, args ...interface{}) *Printer {
	fmt.Fprintf(p, format, args...)
	return p
}

func (p *Printer) Write(raw []byte) (int, error) {
	out := p.Out
	if out == nil {
		out = Std
	}
	pal := p.Palette
	if pal == nil {
		pal = &DefaultPalette
	}
	str := string(raw)
	for _, styles := range p.styles {
		str = pal.Apply(str, styles...)
	}
	return p.Out.Write([]byte(ResetStyler()(pal, str)))
}

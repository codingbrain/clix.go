package term

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type Terminal struct {
	In    io.Reader
	Out   io.Writer
	ANSI  bool
	Color bool

	in, out       *os.File
	inTTY, outTTY bool
	escStrip      ANSIEscStrip
}

var (
	Std = New(os.Stderr, os.Stdin)

	ErrorInputUnavail = errors.New("input not available")
)

func IsTTY(fd uintptr) bool {
	return terminal.IsTerminal(int(fd))
}

func Print(a ...interface{}) *Terminal {
	return Std.Print(a...)
}

func Printf(format string, a ...interface{}) *Terminal {
	return Std.Printf(format, a...)
}

func Println(a ...interface{}) *Terminal {
	return Std.Println(a...)
}

func Success(a ...interface{}) *Terminal {
	return Std.Success(a...)
}

func Successf(format string, a ...interface{}) *Terminal {
	return Std.Successf(format, a...)
}

func Successln(a ...interface{}) *Terminal {
	return Std.Successln(a...)
}

func Warn(a ...interface{}) *Terminal {
	return Std.Warn(a...)
}

func Warnf(format string, a ...interface{}) *Terminal {
	return Std.Warnf(format, a...)
}

func Warnln(a ...interface{}) *Terminal {
	return Std.Warnln(a...)
}

func Error(a ...interface{}) *Terminal {
	return Std.Error(a...)
}

func Errorf(format string, a ...interface{}) *Terminal {
	return Std.Errorf(format, a...)
}

func Errorln(a ...interface{}) *Terminal {
	return Std.Errorln(a...)
}

func Fatal(a ...interface{}) {
	Std.Fatal(a...)
}

func Fatalf(format string, a ...interface{}) {
	Std.Fatalf(format, a...)
}

func Fatalln(a ...interface{}) {
	Std.Fatalln(a...)
}

func OK() *Terminal {
	return Std.OK()
}

func New(out, in *os.File) *Terminal {
	t := &Terminal{In: in, Out: out, in: in, out: out}
	if out == nil {
		t.Out = ioutil.Discard
	} else if t.outTTY = IsTTY(out.Fd()); t.outTTY {
		termType := os.Getenv("TERM")
		if termType != "" && termType != "dumb" && termType != "cons25" {
			t.ANSI = true
			t.Color = true
		}
	}
	if in != nil {
		t.inTTY = IsTTY(in.Fd())
	}
	return t
}

func (t *Terminal) HasInput() bool {
	return t.In != nil
}

func (t *Terminal) IsInTTY() bool {
	return t.inTTY
}

func (t *Terminal) IsTTY() bool {
	return t.outTTY
}

func (t *Terminal) Read(p []byte) (int, error) {
	if t.In != nil {
		return t.In.Read(p)
	}
	return 0, io.EOF
}

func (t *Terminal) Write(p []byte) (int, error) {
	if t.Out == nil {
		return 0, nil
	}
	if !t.Color {
		t.escStrip.Write(p)
		defer t.escStrip.Reset()
		return t.Out.Write(t.escStrip.Bytes())
	} else {
		return t.Out.Write(p)
	}
}

func (t *Terminal) Input() (*Input, error) {
	if t.in != nil && t.inTTY {
		return NewInput(t, t.in.Fd())
	}
	return nil, ErrorInputUnavail
}

func (t *Terminal) ReadLine(prompt string) (string, error) {
	if input, err := NewInput(t, t.in.Fd()); err != nil {
		return "", err
	} else {
		defer input.Close()
		return input.Prompt(prompt).ReadLine()
	}
}

func (t *Terminal) ReadPassword(prompt string) (string, error) {
	if input, err := NewInput(t, t.in.Fd()); err != nil {
		return "", err
	} else {
		defer input.Close()
		return input.ReadPassword(prompt)
	}
}

func (t *Terminal) Print(a ...interface{}) *Terminal {
	fmt.Fprint(t, a...)
	return t
}

func (t *Terminal) Printf(format string, a ...interface{}) *Terminal {
	fmt.Fprintf(t, format, a...)
	return t
}

func (t *Terminal) Println(a ...interface{}) *Terminal {
	fmt.Fprintln(t, a...)
	return t
}

func (t *Terminal) Success(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleOK).Print(a...)
	return t
}

func (t *Terminal) Successf(format string, a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleOK).Printf(format, a...)
	return t
}

func (t *Terminal) Successln(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleOK).Println(a...)
	return t
}

func (t *Terminal) Warn(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleWarn).Print(a...)
	return t
}

func (t *Terminal) Warnf(format string, a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleWarn).Printf(format, a...)
	return t
}

func (t *Terminal) Warnln(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleWarn).Println(a...)
	return t
}

func (t *Terminal) Error(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleErr).Print(a...)
	return t
}

func (t *Terminal) Errorf(format string, a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleErr).Printf(format, a...)
	return t
}

func (t *Terminal) Errorln(a ...interface{}) *Terminal {
	NewPrinter(t).Styles(StyleErr).Println(a...)
	return t
}

func (t *Terminal) Fatal(a ...interface{}) {
	NewPrinter(t).Styles(StyleErr).Print(a...)
	os.Exit(1)
}

func (t *Terminal) Fatalf(format string, a ...interface{}) {
	NewPrinter(t).Styles(StyleErr).Printf(format, a...)
	os.Exit(1)
}

func (t *Terminal) Fatalln(a ...interface{}) {
	NewPrinter(t).Styles(StyleErr).Println(a...)
	os.Exit(1)
}

func (t *Terminal) OK() *Terminal {
	NewPrinter(t).Styles(StyleOK).Println("OK")
	return t
}

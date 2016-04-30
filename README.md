# CLI Experience for Go

It's a highly customizable utility library for writing CLI applications in Go.

It includes a few components help building CLI applications:

- Commandline parsing
- TTY support with readline and password
- Colorized output

## Commandline parsing

This library provides a commandline parsing framework using extensions to hook up
into options/arguments parsing stage and also execution stage.
And all commands and options can be defined using a YAML file.

Here's an example (from `test/istty`):

```go
import (
  ...

	"github.com/codingbrain/clix.go/flag"
	"github.com/codingbrain/clix.go/exts/ask"
	"github.com/codingbrain/clix.go/exts/bind"
	"github.com/codingbrain/clix.go/exts/help"
	"github.com/codingbrain/clix.go/term"
)

type isttyCmd struct {
	Path string
}

func (c *isttyCmd) Execute(args []string) error {
	...
}

...

func main() {
  cli := &flag.CliDef{
		Cli: &flag.Command{
			Name: "istty",
			Arguments: []*flag.Option{
				&flag.Option{
					Name:     "path",
					Alias:    []string{"fd"},
					Desc:     "path to device or file descriptor",
					Required: true,
				},
			},
		},
	}
	cli.Normalize()
	cli.Use(ask.NewExt()).
		Use(bind.NewExt().Bind(&isttyCmd{})).
		Use(help.NewExt()).
		Parse().Exec()
}
```

The parsing framework is as simple as defining a `CliDef`, and all magic happens with extensions:

- `ask` asks user interactively to enter the values of all missing options/arguments which is required
- `bind` maps the values of options/arguments to specified struct and also exec `Execute` if the struct implements `Executable`
- `help` hooks up to flags `--help/-h/-?` to display usage, and it's also responsible to display any errors and exits the application.

## TTY support with readline and password

The `term` package provides simple and essential TTY support.
By wrapping a `Terminal` object over input/output files (`os.File`),
it determines whether input/output is a TTY or not,
and if ANSI escape codes and colors are support or not.
In the application, it's free to emit output with ANSI escape codes,
terminal will automatically strip them if the output doesn't support.

It also provides basic support for reading line and password.

```go

t := term.New(os.Stdin, os.Stderr)
t.Print(...).Printf(...).Println(...)
t.Success(...).Successf(...).Successln(...)
t.Warn(...).Warnf(...).Warnln(...)
t.Error(...).Errorf(...).Errorln(...)
t.Fatal(...)
t.Fatalf(...)
t.Fatalln(...)
t.OK()

t.ReadLine("prompt")
t.ReadPassword("prompt")
```

## Colorized output

This library introduces `Palette` to do flexible output styling.

Here's example from `ask` extension:

```go
term.NewPrinter(writer).
				Styles(term.StyleAsk).Print("Enter ").
				Styles(term.StyleB).Print(help.OptionDisplayName(err.Def)).Pop().
				Print(": ")
```

package main

import (
	"os"
	"strconv"

	"github.com/codingbrain/clix.go/args"
	"github.com/codingbrain/clix.go/exts/ask"
	"github.com/codingbrain/clix.go/exts/bind"
	"github.com/codingbrain/clix.go/exts/help"
	"github.com/codingbrain/clix.go/term"
)

type isttyCmd struct {
	Path string
}

func (c *isttyCmd) Execute(args []string) error {
	fd, err := strconv.Atoi(c.Path)
	if err != nil {
		return checkFile(c.Path)
	}
	return checkFd(uintptr(fd))
}

func checkFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return checkFd(f.Fd())
}

func checkFd(fd uintptr) error {
	if term.IsTTY(fd) {
		term.Successln("Yes")
	} else {
		term.Warnln("No")
	}
	return nil
}

func main() {
	cli := &args.CliDef{
		Cli: &args.Command{
			Name: "istty",
			Arguments: []*args.Option{
				&args.Option{
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

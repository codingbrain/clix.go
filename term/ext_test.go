package term

import (
	"testing"

	"github.com/codingbrain/clix.go/flag"
	"github.com/stretchr/testify/assert"
)

const (
	cliDef = `---
    cli:
      name: test
      options:
        - name: color
          type: bool
`
)

func runParser(t *testing.T, def string, args ...string) *flag.ParseResult {
	a := assert.New(t)
	if cli, err := flag.DecodeCliDefString(def); a.NoError(err) {
		result := cli.Use(NewExt()).ParseArgs(args...)
		result.Exec()
		return result
	}
	return nil
}

func TestApplyColorOpt(t *testing.T) {
	a := assert.New(t)

	Std.Color = true
	runParser(t, cliDef, "test")
	a.True(Std.Color)

	Std.Color = false
	runParser(t, cliDef, "test")
	a.False(Std.Color)

	runParser(t, cliDef, "test", "--color")
	a.True(Std.Color)

	runParser(t, cliDef, "test", "--no-color")
	a.False(Std.Color)
}

func TestSkipApplyColorOpt(t *testing.T) {
	a := assert.New(t)
	Std.Color = true
	if cli, err := flag.DecodeCliDefString(cliDef); a.NoError(err) {
		result := cli.Use(NewExt().ColorOpt("")).ParseArgs("test")
		result.Exec()
		a.True(Std.Color)
	}
}

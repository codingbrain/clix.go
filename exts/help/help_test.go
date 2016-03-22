package help

import (
	"regexp"
	"testing"

	"github.com/codingbrain/clix.go/args"
	"github.com/stretchr/testify/assert"
)

type testRender struct {
	banner *BannerInfo
	usage  *UsageInfo
	cmds   []*args.Command
	opts   []*args.Option
	args   []*args.Option
	errs   []*ErrInfo

	fwd HelpRender
}

func (r *testRender) RenderBanner(b *BannerInfo) {
	r.banner = b
	if r.fwd != nil {
		r.fwd.RenderBanner(b)
	}
}

func (r *testRender) RenderUsage(u *UsageInfo) {
	r.usage = u
	if r.fwd != nil {
		r.fwd.RenderUsage(u)
	}
}

func (r *testRender) RenderCommands(cmds []*args.Command) {
	r.cmds = cmds
	if r.fwd != nil {
		r.fwd.RenderCommands(cmds)
	}
}

func (r *testRender) RenderOptions(opts []*args.Option) {
	r.opts = opts
	if r.fwd != nil {
		r.fwd.RenderOptions(opts)
	}
}

func (r *testRender) RenderArguments(opts []*args.Option) {
	r.args = opts
	if r.fwd != nil {
		r.fwd.RenderArguments(opts)
	}
}

func (r *testRender) RenderErrors(errs []*ErrInfo) {
	r.errs = errs
	if r.fwd != nil {
		r.fwd.RenderErrors(errs)
	}
}

const (
	testCmdDef1 = `---
    cli:
      name: test
      description: |
        test command
        test command long description
      options:
        - name: o1
          description: |
            option1
            option1 long description
      commands:
        - name: c1
          description: c1 desc
          example: c1 example
          options:
            - name: c1o1
              description: c1o1 desc
          commands:
            - name: c1s1
              description: c1s1 desc
              examples: |
                c1s1 eg line1
                c1s1 eg line2
              arguments:
                - name: c1s1a1
                - name: c1s1a2
        - name: c2
          description: c2 desc
          arguments:
            - name: c2a1
              required: true
            - name: c2a2
              list: true
        - name: c3
          options:
            - name: c3o1
              type: int
            - name: '3'
              type: int
          arguments:
            - name: c3a1
              type: int
    `
)

func runParser(t *testing.T, cmdDef string, cmdArgs ...string) (*testRender, *args.ParseResult) {
	a := assert.New(t)
	if cli, err := args.DecodeCliDefString(cmdDef); a.NoError(err) {
		render := &testRender{fwd: &DefaultRender{}}
		result := cli.Cli.Parser().
			Use(NewExt().UseRender(render).NoExit()).
			ParseArgs(cmdArgs)
		err = result.Exec()
		if a.Error(err) {
			a.Equal(ErrorHelp, err)
		}
		return render, result
	}
	return nil, nil
}

func TestHelpOption(t *testing.T) {
	a := assert.New(t)
	render, res := runParser(t, testCmdDef1, "test", "--help", "c1")
	if a.NotNil(render) {
		if a.Len(res.UnparsedArgs, 1) {
			a.Equal("c1", res.UnparsedArgs[0])
		}
		a.NotNil(render.banner)
		if a.NotNil(render.usage) {
			a.Len(render.usage.Cmds, 1)
			a.Len(render.usage.Args, 1)
		}
		a.Equal([]string{"test"}, render.usage.Cmds)
		if a.Len(render.cmds, 3) {
			a.Equal("c1", render.cmds[0].Name)
			a.Equal("c2", render.cmds[1].Name)
			a.Equal("c3", render.cmds[2].Name)
		}
	}
}

func TestHelpArguments(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c2", "--help")
	if a.NotNil(render) {
		a.NotNil(render.banner)
		if a.NotNil(render.usage) {
			if a.Len(render.usage.Cmds, 2) {
				a.Equal("test", render.usage.Cmds[0])
				a.Equal("c2", render.usage.Cmds[1])
			}
			if a.Len(render.usage.Args, 2) {
				a.Equal("C2A1", render.usage.Args[0])
				a.Equal("[C2A2]", render.usage.Args[1])
			}
		}
		a.Equal([]string{"test", "c2"}, render.usage.Cmds)
		a.Empty(render.cmds)
		if a.Len(render.args, 2) {
			a.Equal("c2a1", render.args[0].Name)
			a.Equal("c2a2", render.args[1].Name)
		}
	}
}

func TestHelpMissingCmd(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "unknown")
	if a.NotNil(render) {
		a.Nil(render.banner)
		if a.NotNil(render.usage) {
			a.Equal([]string{"test"}, render.usage.Cmds)
		}
		if a.Len(render.errs, 1) {
			a.Regexp(regexp.MustCompile(`unknown command: unknown$`), render.errs[0].Msg)
		}
	}
}

func TestHelpNoDef(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c2", "--unknown", "c2a1")
	if a.NotNil(render) {
		a.Nil(render.banner)
		a.Nil(render.usage)
		if a.Len(render.errs, 1) {
			a.Regexp(regexp.MustCompile(`unknown option: unknown$`), render.errs[0].Msg)
		}
	}
}

func TestHelpOptNoVal(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c3", "--c3o1", "-3")
	if a.NotNil(render) {
		a.Nil(render.banner)
		a.Nil(render.usage)
		if a.Len(render.errs, 2) {
			a.Regexp(regexp.MustCompile(`expect value C3O1 after --c3o1$`), render.errs[0].Msg)
			a.Regexp(regexp.MustCompile(`expect value 3 after -3$`), render.errs[1].Msg)
		}
	}
}

func TestHelpArgNoVal(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c2")
	if a.NotNil(render) {
		a.Nil(render.banner)
		a.Nil(render.usage)
		if a.Len(render.errs, 1) {
			a.Regexp(regexp.MustCompile(`expect argument C2A1$`), render.errs[0].Msg)
		}
	}
}

func TestHelpOptBadVal(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c3", "--c3o1=abc", "-3", "C")
	if a.NotNil(render) {
		a.Nil(render.banner)
		a.Nil(render.usage)
		if a.Len(render.errs, 2) {
			a.Regexp(regexp.MustCompile(`invalid value for --c3o1: abc$`), render.errs[0].Msg)
			a.Regexp(regexp.MustCompile(`invalid value for -3: C$`), render.errs[1].Msg)
		}
	}
}

func TestHelpArgBadVal(t *testing.T) {
	a := assert.New(t)
	render, _ := runParser(t, testCmdDef1, "test", "c3", "ABC")
	if a.NotNil(render) {
		a.Nil(render.banner)
		a.Nil(render.usage)
		if a.Len(render.errs, 1) {
			a.Regexp(regexp.MustCompile(`invalid value for argument C3A1: ABC$`), render.errs[0].Msg)
		}
	}
}

func TestCustomHelpOptions(t *testing.T) {
	a := assert.New(t)
	if cli, err := args.DecodeCliDefString(testCmdDef1); a.NoError(err) {
		render := &testRender{fwd: &DefaultRender{}}
		err = cli.Cli.Parser().
			Use(NewExt().OptNames("sos", "!").UseRender(render).NoExit()).
			ParseArgs([]string{"test", "--sos"}).Exec()
		if a.Error(err) {
			a.Equal(ErrorHelp, err)
		}
		err = cli.Cli.Parser().
			Use(NewExt().OptNames("sos", "!", "$").UseRender(render).NoExit()).
			ParseArgs([]string{"test", "-!"}).Exec()
		if a.Error(err) {
			a.Equal(ErrorHelp, err)
		}
		err = cli.Cli.Parser().
			Use(NewExt().OptNames("sos", "!", "$").UseRender(render).NoExit()).
			ParseArgs([]string{"test", "-$"}).Exec()
		if a.Error(err) {
			a.Equal(ErrorHelp, err)
		}
		result := cli.Cli.Parser().
			Use(NewExt().OptNames("sos", "!", "$").UseRender(render).NoExit()).
			ParseArgs([]string{"test", "--help"})
		if a.Error(result.Exec()) {
			a.Equal(ErrorHelp, err)
			a.NotEmpty(result.CmdStack[0].Errs)
		}
	}
}

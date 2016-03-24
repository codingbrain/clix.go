package args

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cli = loadCli()

func loadCli() *CliDef {
	cli, err := DecodeCliDefFile("test.yml")
	if err != nil {
		panic(err)
	}
	return cli
}

func TestShortFlags(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "up", "-a", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		a.Equal("cli", r.Program)

		cs := r.CmdStack[0]
		a.Equal("test", cs.Cmd.Name)
		a.Empty(cs.Args)
		a.Zero(cs.ParsedArgC)
		a.Equal("127.0.0.1:8080", cs.Vars["server"].(string))
		a.Empty(cs.Errs)

		cs = r.CmdStack[1]
		a.Equal("up", cs.Cmd.Name)
		a.Len(cs.Args, 1)
		a.Equal(1, cs.ParsedArgC)
		a.True(cs.Vars["flaga"].(bool))
		a.False(cs.Vars["flagb"].(bool))
	}
	r = cli.ParseArgs("cli", "up", "-ab", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("up", cs.Cmd.Name)
		a.Len(cs.Args, 1)
		a.Equal(1, cs.ParsedArgC)
		a.True(cs.Vars["flaga"].(bool))
		a.True(cs.Vars["flagb"].(bool))
	}
	r = cli.ParseArgs("cli", "up", "-a", "-b", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.True(cs.Vars["flaga"].(bool))
		a.True(cs.Vars["flagb"].(bool))
	}
}

func TestLongFlags(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "--server=123", "up", "--flaga", "--no-flagb")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[0]
		a.Equal("test", cs.Cmd.Name)
		a.Empty(cs.Args)
		a.Zero(cs.ParsedArgC)
		a.Equal("123", cs.Vars["server"].(string))
		a.Empty(cs.Errs)

		cs = r.CmdStack[1]
		a.Equal("up", cs.Cmd.Name)
		a.Len(cs.Args, 1)
		a.Zero(cs.ParsedArgC)
		a.True(cs.Vars["flaga"].(bool))
		a.False(cs.Vars["flagb"].(bool))
		if a.Len(cs.Errs, 1) {
			a.Equal("object", cs.Errs[0].Name)
			a.Equal(VarErrNoVal, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Value)
		}
	}
	r = cli.ParseArgs("cli", "down", "-fVAL")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("down", cs.Cmd.Name)
		a.Empty(cs.Args)
		a.Zero(cs.ParsedArgC)
		a.Equal("VAL", cs.Vars["flag"].(string))
		a.True(cs.Vars["wait"].(bool))
	}
	r = cli.ParseArgs("cli", "down", "-wf", "VAL", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("down", cs.Cmd.Name)
		a.Len(cs.Args, 1)
		a.Zero(cs.ParsedArgC)
		a.True(cs.Vars["wait"].(bool))
		a.Equal("VAL", cs.Vars["flag"].(string))
	}
	r = cli.ParseArgs("cli", "down", "-f", "VAL", "--no-wait")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("down", cs.Cmd.Name)
		a.False(cs.Vars["wait"].(bool))
		a.Equal("VAL", cs.Vars["flag"].(string))
	}
}

func TestDefaults(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "defs")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("defs", cs.Cmd.Name)
		a.Len(cs.Args, 5)
		a.Equal("", cs.Vars["str"])
		a.Equal(int64(0), cs.Vars["int"])
		a.Equal(float64(0), cs.Vars["num"])
		a.Equal(map[string]interface{}{}, cs.Vars["dict"])
		a.Equal("", cs.Vars["slice"]) // args should never be list
	}
	r = cli.ParseArgs("cli", "d")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("defs", cs.Cmd.Name)
		a.Empty(cs.Errs)
	}
}

func TestListOptions(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "list", "--items=164", "--items=682", "--adds=6")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("list", cs.Cmd.Name)
		a.Empty(cs.Errs)
		values, ok := cs.Vars["items"].([]interface{})
		if a.True(ok) && a.Len(values, 2) {
			a.Equal(int64(164), values[0])
			a.Equal(int64(682), values[1])
		}
		values, ok = cs.Vars["adds"].([]interface{})
		if a.True(ok) && a.Len(values, 2) {
			a.Equal(float64(3.14), values[0])
			a.Equal(float64(6), values[1])
		}
	}
}

func TestMapOptions(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "map", "--kv=b=b1", "--kv=c=c1", "--no-defs=x=x")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("map", cs.Cmd.Name)
		a.Empty(cs.Errs)
		dict, ok := cs.Vars["kv"].(map[string]interface{})
		if a.True(ok) && a.Len(dict, 3) {
			a.Equal("a1", dict["a"])
			a.Equal("b1", dict["b"])
			a.Equal("c1", dict["c"])
		}
		dict, ok = cs.Vars["no-defs"].(map[string]interface{})
		if a.True(ok) && a.Len(dict, 1) {
			a.Equal("x", dict["x"])
		}
	}
	r = cli.ParseArgs("cli", "map", "--kv=a=b1", "--kv=c=c1", "--no-defs=x")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		a.Equal("map", cs.Cmd.Name)
		a.Empty(cs.Errs)
		dict, ok := cs.Vars["kv"].(map[string]interface{})
		if a.True(ok) && a.Len(dict, 2) {
			a.Equal("b1", dict["a"])
			a.Equal("c1", dict["c"])
		}
		dict, ok = cs.Vars["no-defs"].(map[string]interface{})
		if a.True(ok) && a.Len(dict, 1) {
			a.Equal(true, dict["x"])
		}
	}
}

func TestMissingRequired(t *testing.T) {
	a := assert.New(t)
	// required opt
	r := cli.ParseArgs("cli", "reqs", "a", "b")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 1) {
			a.Equal("req1", cs.Errs[0].Name)
			a.Equal(VarErrNoVal, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Value)
		}
	}
	// required arg
	r = cli.ParseArgs("cli", "up")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 1) {
			a.Equal("object", cs.Errs[0].Name)
			a.Equal(VarErrNoVal, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Value)
		}
	}
}

func TestMissingValue(t *testing.T) {
	a := assert.New(t)
	// missing long flag value
	r := cli.ParseArgs("cli", "reqs", "--req1", "a")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 1) {
			a.Equal("req1", cs.Errs[0].Name)
			a.Equal(VarErrNoVal, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Value)
		}
	}
	// missing short flag value
	r = cli.ParseArgs("cli", "down", "-f")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 1) {
			a.Equal("f", cs.Errs[0].Name)
			a.NotNil(cs.Errs[0].Def)
			a.Equal(VarErrNoVal, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Value)
		}
	}
}

func TestBadValue(t *testing.T) {
	a := assert.New(t)
	// missing long flag value
	r := cli.ParseArgs("cli", "defs", "str", "not-int", "not-num", "=")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 3) {
			a.Equal("int", cs.Errs[0].Name)
			a.Equal(VarErrBadVal, cs.Errs[0].ErrType)
			a.NotNil(cs.Errs[0].Def)
			if a.NotNil(cs.Errs[0].Value) {
				a.Equal("not-int", *cs.Errs[0].Value)
			}
			a.Equal("num", cs.Errs[1].Name)
			a.Equal(VarErrBadVal, cs.Errs[1].ErrType)
			a.NotNil(cs.Errs[1].Def)
			if a.NotNil(cs.Errs[1].Value) {
				a.Equal("not-num", *cs.Errs[1].Value)
			}
			a.Equal("dict", cs.Errs[2].Name)
			a.Equal(VarErrBadVal, cs.Errs[2].ErrType)
			a.NotNil(cs.Errs[2].Def)
			if a.NotNil(cs.Errs[2].Value) {
				a.Equal("=", *cs.Errs[2].Value)
			}
		}
	}
}

func TestInvalidOpt(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "up", "--unknown=1", "-x", "--no-unknown", "obj")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 3) {
			a.Equal("unknown", cs.Errs[0].Name)
			a.Equal(VarErrNoDef, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Def)
			a.Nil(cs.Errs[0].Value)
			a.Equal("x", cs.Errs[1].Name)
			a.Equal(VarErrNoDef, cs.Errs[1].ErrType)
			a.Nil(cs.Errs[1].Def)
			a.Nil(cs.Errs[1].Value)
			a.Equal("no-unknown", cs.Errs[2].Name)
			a.Equal(VarErrNoDef, cs.Errs[2].ErrType)
			a.Nil(cs.Errs[2].Def)
			a.Nil(cs.Errs[2].Value)
		}
	}

	r = cli.ParseArgs("cli", "up", "--=1", "obj")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.True(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 1) {
			a.Equal("--=1", cs.Errs[0].Name)
			a.Equal(VarErrNoDef, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Def)
			a.Nil(cs.Errs[0].Value)
		}
	}
}

func TestMissingCmd(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "not-exist", "a", "b")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 1) && a.True(r.MissingCmd) {
		a.True(r.HasErrors())
		a.Equal([]string{"not-exist", "a", "b"}, r.UnparsedArgs)
	}
	r = cli.Parser().ParseArgs("cli", "not-exist", "a", "--", "b")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 1) && a.True(r.MissingCmd) {
		a.True(r.HasErrors())
		a.Equal([]string{"not-exist", "a", "b"}, r.UnparsedArgs)
	}
}

func TestStopParsing(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "up", "--", "b", "c")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[1]
		if a.Empty(cs.Errs) {
			a.Equal([]string{"b", "c"}, cs.Args)
			a.Equal(1, cs.ParsedArgC)
			a.Equal("b", cs.Vars["object"].(string))
		}
	}
}

func TestNoSubCmd(t *testing.T) {
	a := assert.New(t)
	r := cli.ParseArgs("cli", "-sA")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 1) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[0]
		if a.Empty(cs.Errs) {
			a.Equal("A", cs.Vars["server"].(string))
		}
	}
	r = cli.ParseArgs("cli")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 1) && a.False(r.MissingCmd) {
		a.False(r.HasErrors())
		cs := r.CmdStack[0]
		a.Empty(cs.Errs)
	}
	r = cli.ParseArgs()
	a.Error(r.Error)
	a.True(r.HasErrors())
	a.Empty(r.CmdStack)
	a.False(r.MissingCmd)
	a.Empty(r.Program)
}

type testStartCmdExt struct {
	t    *testing.T
	cmds []string
}

func (x *testStartCmdExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtStartCmd, event)
	ctx.SetVar("touched", "yes")
	if a.NotNil(ctx.CurrentCmd()) {
		x.cmds = append(x.cmds, ctx.CurrentCmd().Cmd.Name)
		a.Len(ctx.CmdStack(), len(x.cmds))
		ctx.SetVarAt(0, "cmd", ctx.CurrentCmd().Cmd.Name)
	}
	a.Nil(ctx.CmdAt(len(ctx.CmdStack())))
}

func TestParseExtStartCmd(t *testing.T) {
	a := assert.New(t)
	ext := &testStartCmdExt{t: t}
	r := cli.Parser().AddParseExt(EvtStartCmd, ext).ParseArgs("cli", "up", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.Len(ext.cmds, 2) {
		a.Equal("test", ext.cmds[0])
		a.Equal("up", ext.cmds[1])
		if a.Contains(r.CmdStack[0].Vars, "cmd") {
			a.Equal("up", r.CmdStack[0].Vars["cmd"])
		}
		a.Contains(r.CmdStack[0].Vars, "touched")
		a.Contains(r.CmdStack[1].Vars, "touched")
	}
}

type assignment struct {
	at    int
	opt   *Option
	name  string
	value *string
	not   bool
}

type testAssignOptExt struct {
	t       *testing.T
	assigns []*assignment
}

func (x *testAssignOptExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtAssignOpt, event)
	x.assigns = append(x.assigns, &assignment{
		at:    ctx.OptionAt,
		opt:   ctx.Option,
		name:  ctx.Name,
		value: ctx.Value,
		not:   ctx.Not,
	})
}

func TestParseExtAssignOpt(t *testing.T) {
	a := assert.New(t)
	ext := &testAssignOptExt{t: t}
	r := cli.Parser().AddParseExt(EvtAssignOpt, ext).ParseArgs("cli", "up", "--no-flaga", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.Len(ext.assigns, 2) {
		a.Equal(1, ext.assigns[0].at)
		a.NotNil(ext.assigns[0].opt)
		a.Equal("flaga", ext.assigns[0].name)
		a.NotNil(ext.assigns[0].value)
		a.Equal("true", *ext.assigns[0].value)
		a.True(ext.assigns[0].not)

		a.Equal(1, ext.assigns[1].at)
		a.NotNil(ext.assigns[1].opt)
		a.Equal("object", ext.assigns[1].name)
		a.NotNil(ext.assigns[1].value)
		a.Equal("a1", *ext.assigns[1].value)
		a.False(ext.assigns[1].not)
	}
}

type testAssignOptSkipValueExt struct {
	t *testing.T
}

func (x *testAssignOptSkipValueExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtAssignOpt, event)
	ctx.Value = nil
}

func TestParseExtAssignOptSkipValue(t *testing.T) {
	a := assert.New(t)
	ext := &testAssignOptSkipValueExt{t: t}
	r := cli.Parser().AddParseExt(EvtAssignOpt, ext).ParseArgs("cli", "up", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) {
		cs := r.CmdStack[1]
		a.NotContains(cs.Vars, "object")
	}
}

type resolveOpt struct {
	name  string
	value *string
}

type testResolveOptExt struct {
	t *testing.T
	r []*resolveOpt
}

func (x *testResolveOptExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtResolveOpt, event)
	x.r = append(x.r, &resolveOpt{name: ctx.Name, value: ctx.Value})
}

func TestParseExtResolveOpt(t *testing.T) {
	a := assert.New(t)
	ext := &testResolveOptExt{t: t}
	r := cli.Parser().AddParseExt(EvtResolveOpt, ext).ParseArgs("cli", "up", "--unknown1=1", "--unknown2", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) && a.Len(ext.r, 2) {
		cs := r.CmdStack[1]
		if a.Len(cs.Errs, 2) {
			a.Equal("unknown1", cs.Errs[0].Name)
			a.Equal(VarErrNoDef, cs.Errs[0].ErrType)
			a.Nil(cs.Errs[0].Def)
			a.Nil(cs.Errs[0].Value)
			a.Equal("unknown2", cs.Errs[1].Name)
			a.Equal(VarErrNoDef, cs.Errs[1].ErrType)
			a.Nil(cs.Errs[1].Def)
			a.Nil(cs.Errs[1].Value)
		}
		a.Equal("unknown1", ext.r[0].name)
		a.NotNil(ext.r[0].value)
		a.Equal("1", *ext.r[0].value)
		a.Equal("unknown2", ext.r[1].name)
		a.Nil(ext.r[1].value)
	}
}

type testResolveOptSkipExt struct {
	t *testing.T
}

func (x *testResolveOptSkipExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtResolveOpt, event)
	ctx.Ignore = true
}

func TestParseExtResolveOptSkip(t *testing.T) {
	a := assert.New(t)
	ext := &testResolveOptSkipExt{t: t}
	r := cli.Parser().AddParseExt(EvtResolveOpt, ext).ParseArgs("cli", "up", "--unknown1=1", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) {
		cs := r.CmdStack[1]
		a.Empty(cs.Errs)
	}
}

type testDoneExt struct {
	t       *testing.T
	touched bool
}

func (x *testDoneExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtStartCmd, event)
	x.touched = true
	ctx.Done()
}

func TestParseExtDone(t *testing.T) {
	a := assert.New(t)
	x1 := &testDoneExt{t: t}
	x2 := &testDoneExt{t: t}
	r := cli.Parser().
		AddParseExt(EvtStartCmd, x1).
		AddParseExt(EvtStartCmd, x2).
		ParseArgs("cli", "up", "a1")
	if a.NoError(r.Error) && a.Len(r.CmdStack, 2) {
		a.True(x1.touched)
		a.False(x2.touched)
	}
}

type testAbortExt struct {
	t       *testing.T
	touched bool
}

func (x *testAbortExt) HandleParseEvent(event string, ctx *ParseContext) {
	a := assert.New(x.t)
	a.Equal(EvtResolveOpt, event)
	x.touched = true
	ctx.Abort(errors.New("aborted: " + ctx.Name))
}

func TestParseExtAbort(t *testing.T) {
	a := assert.New(t)
	x1 := &testAbortExt{t: t}
	x2 := &testAbortExt{t: t}
	r := cli.Parser().
		AddParseExt(EvtResolveOpt, x1).
		AddParseExt(EvtResolveOpt, x2).
		ParseArgs("cli", "--unknown", "up")
	if a.Error(r.Error) {
		a.Equal("aborted: unknown", r.Error.Error())
		a.True(x1.touched)
		a.False(x2.touched)
		a.Len(r.UnparsedArgs, 1)
		a.Equal("up", r.UnparsedArgs[0])
	}
}

type testExecExt struct {
	t *testing.T
	r *ParseResult
}

func (x *testExecExt) ExecuteCmd(ctx *ExecContext) {
	a := assert.New(x.t)
	x.r = ctx.Result
	a.False(ctx.HasErrors())
	a.NotNil(ctx.Cmd())
}

func (x *testExecExt) RegisterExt(parser *Parser) {
	parser.AddExecExt(x)
}

func TestExecExt(t *testing.T) {
	a := assert.New(t)
	x := &testExecExt{t: t}
	err := cli.Use(x).ParseArgs("cli", "up", "a1").Exec()
	if a.NoError(err) {
		a.NotNil(x.r)
	}
}

type testExecDoneExt struct {
	t *testing.T
	r *ParseResult
}

func (x *testExecDoneExt) ExecuteCmd(ctx *ExecContext) {
	a := assert.New(x.t)
	x.r = ctx.Result
	a.False(ctx.HasErrors())
	a.NotNil(ctx.Cmd())
	a.NoError(ctx.Result.Error)
	ctx.Done(errors.New("done"))
}

func TestExecDoneExt(t *testing.T) {
	a := assert.New(t)
	x1 := &testExecDoneExt{t: t}
	x2 := &testExecDoneExt{t: t}
	err := cli.Parser().
		AddExecExt(x1).
		AddExecExt(x2).
		ParseArgs("cli", "up", "a1").
		Exec()
	if a.Error(err) {
		a.NotNil(x1.r)
		a.Nil(x2.r)
		a.Equal("done", err.Error())
	}
}

package args

const (
	EvtStartCmd   = "cmd.start"
	EvtAssignOpt  = "opt.assign"
	EvtAssigned   = "var.assigned"
	EvtResolveOpt = "opt.resolve"
)

type ParseContext struct {
	OptionAt int
	Option   *Option
	Name     string
	Value    *string
	Not      bool
	Assigned interface{}
	Ignore   bool

	parser   *Parser
	stopped  bool
	abortErr error
}

func (c *ParseContext) Done() {
	c.stopped = true
}

func (c *ParseContext) Abort(err error) {
	c.stopped = true
	c.abortErr = err
}

func (c *ParseContext) CmdStack() []*ParsedCmd {
	return c.parser.result.CmdStack
}

func (c *ParseContext) CurrentCmd() *ParsedCmd {
	return c.parser.currCmd
}

func (c *ParseContext) CmdAt(at int) *ParsedCmd {
	if at >= 0 && at < len(c.parser.result.CmdStack) {
		return c.parser.result.CmdStack[at]
	}
	return nil
}

func (c *ParseContext) SetVarAt(at int, name string, val interface{}) *ParseContext {
	if pcmd := c.CmdAt(at); pcmd != nil {
		pcmd.Vars[name] = val
	}
	return c
}

func (c *ParseContext) SetVar(name string, val interface{}) *ParseContext {
	c.CurrentCmd().Vars[name] = val
	return c
}

type ExecContext struct {
	Result *ParseResult

	completed bool
	err       error
}

func (c *ExecContext) Cmd() *ParsedCmd {
	if l := len(c.Result.CmdStack); l == 0 {
		return nil
	} else {
		return c.Result.CmdStack[l-1]
	}
}

func (c *ExecContext) CmdAt(at int) *ParsedCmd {
	if at >= 0 && at < len(c.Result.CmdStack) {
		return c.Result.CmdStack[at]
	}
	return nil
}

func (c *ExecContext) HasErrors() bool {
	return c.err != nil || c.Result.HasErrors()
}

func (c *ExecContext) Done(err error) {
	c.completed = true
	c.err = err
}

type ParseExt interface {
	HandleParseEvent(event string, context *ParseContext)
}

type ExecExt interface {
	ExecuteCmd(context *ExecContext)
}

type ExtRegistrar interface {
	RegisterExt(parser *Parser)
}

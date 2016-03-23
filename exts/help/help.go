package help

import (
	"errors"
	"os"
	"strings"

	"github.com/codingbrain/clix.go/args"
)

var (
	// DefaultLong defines the default long option for help
	DefaultLong = "help"
	// DefaultAlias defines the default alias options for help
	DefaultAlias = []string{"h", "?"}

	// ErrorHelp is used as error when help is displayed
	ErrorHelp = errors.New("help requested")
)

// BannerInfo defines the information to be displayed as banner
type BannerInfo struct {
	// Banner is the string to be displayed
	Banner []string
}

// ErrInfo contains pre-formatted error message and origin of the error
type ErrInfo struct {
	// pre-formatted message
	Msg string
	// present with cmd string which is invalid
	Cmd string
	// present indicate a parsing error
	Var *args.VarError
}

// UsageInfo defines the information to be displayed in usage line
type UsageInfo struct {
	Cmds []string
	Opts []string
	Args []string
	Tail []string
}

// HelpRender defines the interface that a render to implement to show help
type HelpRender interface {
	// RenderBanner displays banner
	RenderBanner(*BannerInfo)
	// RenderUsage displays usage line
	RenderUsage(*UsageInfo)
	// RenderCommands displays subcommands and details
	RenderCommands([]*args.Command)
	// RenderArguments displays arguments and details
	RenderArguments([]*args.Option)
	// RenderOptions displays options and details
	RenderOptions([]*args.Option)
	// RenderErrors displays error messages
	RenderErrors([]*ErrInfo)
}

// HelpExt defines the help extension which must be hooked up to
// - EvtResoveOpt
// - Execution
type HelpExt struct {
	Long     string
	Alias    []string
	Render   HelpRender
	ExitCode int
	AnyError bool

	helpCmdAt int
}

// NewExt creates help extension
func NewExt() *HelpExt {
	return &HelpExt{
		Long:      DefaultLong,
		Alias:     DefaultAlias,
		Render:    &DefaultRender{},
		ExitCode:  2,
		AnyError:  true,
		helpCmdAt: -1,
	}
}

// OptNames overrides long and alias help options
func (x *HelpExt) OptNames(long string, alias ...string) *HelpExt {
	x.Long = long
	x.Alias = alias
	return x
}

// UseRender sets the render for help information
func (x *HelpExt) UseRender(render HelpRender) *HelpExt {
	x.Render = render
	return x
}

// NoExit prevents the extension invoke os.Exit(x.ExitCode)
// and returns ErrorHelp in result only
func (x *HelpExt) NoExit() *HelpExt {
	x.ExitCode = -1
	return x
}

// ExitWith specifies the exit code to use when exit after help displayed
func (x *HelpExt) ExitWith(code int) *HelpExt {
	x.ExitCode = code
	return x
}

// ExecuteCmd implements execution extension
func (x *HelpExt) ExecuteCmd(ctx *args.ExecContext) {
	if err := ctx.Result.Error; err == ErrorHelp && x.helpCmdAt >= 0 {
		x.displayHelp(ctx.Result.CmdStack, x.helpCmdAt, true)
		x.exit()
		return
	} else if err != nil {
		if err != ErrorHelp && x.AnyError {
			x.displayErrors([]*ErrInfo{&ErrInfo{Msg: err.Error()}})
			x.exit()
		}
		return
	}

	if ctx.Result.MissingCmd {
		x.displayErrors([]*ErrInfo{&ErrInfo{Cmd: ctx.Result.UnparsedArgs[0]}})
		x.displayHelp(ctx.Result.CmdStack, len(ctx.Result.CmdStack)-1, false)
	} else {
		errs := make([]*ErrInfo, 0, 0)
		for _, pcmd := range ctx.Result.CmdStack {
			for _, e := range pcmd.Errs {
				errs = append(errs, &ErrInfo{Var: e})
			}
		}
		if len(errs) == 0 {
			return
		}
		x.displayErrors(errs)
	}
	ctx.Done(ErrorHelp)
	x.exit()
}

// HandleParseEvent implements parse extension
func (x *HelpExt) HandleParseEvent(event string, ctx *args.ParseContext) {
	if event != args.EvtResolveOpt || x.helpCmdAt >= 0 || ctx.Name == "" {
		return
	}
	if ctx.Name != x.Long {
		found := false
		for _, a := range x.Alias {
			if ctx.Name == a {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	x.helpCmdAt = len(ctx.CmdStack()) - 1
	ctx.Ignore = true
	ctx.Abort(ErrorHelp)
}

// RegisterExt implements ExtRegistrar
func (x *HelpExt) RegisterExt(parser *args.Parser) {
	parser.AddParseExt(args.EvtResolveOpt, x)
	parser.AddExecExt(x)
	x.helpCmdAt = -1
}

// RenderBanner self implements HelpRender
func (x *HelpExt) RenderBanner(info *BannerInfo) {
	if x.Render != nil {
		x.Render.RenderBanner(info)
	}
}

// RenderUsage self implements HelpRender
func (x *HelpExt) RenderUsage(info *UsageInfo) {
	if x.Render != nil {
		x.Render.RenderUsage(info)
	}
}

// RenderCommands self implements HelpRender
func (x *HelpExt) RenderCommands(cmds []*args.Command) {
	if x.Render != nil {
		x.Render.RenderCommands(cmds)
	}
}

// RenderArguments self implements HelpRender
func (x *HelpExt) RenderArguments(opts []*args.Option) {
	if x.Render != nil {
		x.Render.RenderArguments(opts)
	}
}

// RenderOptions self implements HelpRender
func (x *HelpExt) RenderOptions(opts []*args.Option) {
	if x.Render != nil {
		x.Render.RenderOptions(opts)
	}
}

// RenderErrors self implements HelpRender
func (x *HelpExt) RenderErrors(errs []*ErrInfo) {
	if x.Render != nil {
		x.Render.RenderErrors(errs)
	}
}

func (x *HelpExt) displayHelp(stack []*args.ParsedCmd, at int, banner bool) {
	if banner {
		pcmd := stack[0]
		if pcmd.Cmd.Desc != "" {
			x.RenderBanner(&BannerInfo{Banner: []string{pcmd.Cmd.Desc}})
		}
	}

	usage := &UsageInfo{}
	optCount := 0
	for i := 0; i <= at; i++ {
		pcmd := stack[i]
		usage.Cmds = append(usage.Cmds, pcmd.Cmd.Name)
		optCount += len(pcmd.Cmd.Options)
	}
	if optCount > 0 {
		usage.Opts = []string{"[OPTIONS]"}
	}
	pcmd := stack[at]
	if len(pcmd.Cmd.Commands) > 0 {
		usage.Args = []string{"SUBCOMMAND"}
	} else {
		for _, arg := range pcmd.Cmd.Arguments {
			name := strings.ToUpper(arg.ValueName())
			if !arg.Required {
				name = "[" + name + "]"
			}
			usage.Args = append(usage.Args, name)
		}
	}
	usage.Tail = []string{"..."}

	x.RenderUsage(usage)

	if len(pcmd.Cmd.Commands) > 0 {
		x.RenderCommands(pcmd.Cmd.Commands)
	} else if len(pcmd.Cmd.Arguments) > 0 {
		x.RenderArguments(pcmd.Cmd.Arguments)
	}

	var opts []*args.Option
	for i := 0; i <= at; i++ {
		pcmd := stack[i]
		opts = append(opts, pcmd.Cmd.Options...)
	}
	if len(opts) > 0 {
		x.RenderOptions(opts)
	}
}

func (x *HelpExt) displayErrors(errs []*ErrInfo) {
	for _, err := range errs {
		if err.Cmd != "" {
			err.Msg = "unknown command: " + err.Cmd
		} else if err.Var != nil {
			switch err.Var.ErrType {
			case args.VarErrNoDef:
				err.Msg = "unknown option: " + err.Var.Name
			case args.VarErrNoVal:
				if err.Var.Def.IsArg {
					err.Msg = "expect argument " + strings.ToUpper(err.Var.Def.ValueName())
				} else {
					err.Msg = "expect value " + strings.ToUpper(err.Var.Def.ValueName()) + " after "
					if len(err.Var.Name) > 1 {
						err.Msg += "--" + err.Var.Name
					} else {
						err.Msg += "-" + err.Var.Name
					}
				}
			case args.VarErrBadVal:
				if err.Var.Def.IsArg {
					err.Msg = "invalid value for argument " + strings.ToUpper(err.Var.Def.ValueName())
				} else {
					err.Msg = "invalid value for "
					if len(err.Var.Name) > 1 {
						err.Msg += "--" + err.Var.Name
					} else {
						err.Msg += "-" + err.Var.Name
					}
				}
				if err.Var.Value != nil {
					err.Msg += ": " + *err.Var.Value
				}
			}
		}
	}
	if len(errs) > 0 {
		x.RenderErrors(errs)
	}
}

func (x *HelpExt) exit() {
	if x.ExitCode >= 0 {
		os.Exit(x.ExitCode)
	}
}

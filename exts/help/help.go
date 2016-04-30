package help

import (
	"errors"
	"os"
	"strings"

	"github.com/codingbrain/clix.go/flag"
)

const (
	TagVar = "help-var"
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
	Var *flag.VarError
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
	// RenderStart indicates the start of render process
	RenderStart()
	// RenderComplete indicates the end of render process
	RenderComplete()
	// RenderBanner displays banner
	RenderBanner(*BannerInfo)
	// RenderUsage displays usage line
	RenderUsage(*UsageInfo)
	// RenderCommands displays subcommands and details
	RenderCommands([]*flag.Command)
	// RenderArguments displays arguments and details
	RenderArguments([]*flag.Option)
	// RenderOptions displays options and details
	RenderOptions([]*flag.Option)
	// RenderErrors displays error messages
	RenderErrors([]*ErrInfo)
}

// HelpExt defines the help extension which must be hooked up to
// - EvtResoveOpt
// - Execution
type HelpExt struct {
	Long         string
	Alias        []string
	Render       HelpRender
	HelpExitCode int
	ErrExitCode  int

	helpCmdAt int
}

// NewExt creates help extension
func NewExt() *HelpExt {
	return &HelpExt{
		Long:         DefaultLong,
		Alias:        DefaultAlias,
		Render:       &DefaultRender{},
		HelpExitCode: 2,
		ErrExitCode:  1,
		helpCmdAt:    -1,
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
	x.HelpExitCode = -1
	x.ErrExitCode = -1
	return x
}

// ExitWith specifies the exit code to use when exit after help displayed
func (x *HelpExt) ExitWith(helpCode, errCode int) *HelpExt {
	x.HelpExitCode = helpCode
	x.ErrExitCode = errCode
	return x
}

// ExecuteCmd implements execution extension
func (x *HelpExt) ExecuteCmd(ctx *flag.ExecContext) {
	err := ctx.Result.Error
	if err == ErrorHelp && x.helpCmdAt >= 0 {
		x.RenderStart()
		x.displayHelp(ctx.Result.CmdStack, x.helpCmdAt, true)
		x.RenderComplete()
		exit(x.HelpExitCode)
		return
	}
	if err != nil {
		if err != ErrorHelp && x.ErrExitCode >= 0 {
			x.RenderStart()
			x.displayErrors([]*ErrInfo{&ErrInfo{Msg: err.Error()}})
			x.RenderComplete()
			exit(x.ErrExitCode)
		}
		return
	}

	x.RenderStart()
	if ctx.Result.MissingCmd {
		x.displayErrors([]*ErrInfo{&ErrInfo{Cmd: ctx.Result.UnparsedArgs[0]}})
	} else if !ctx.Result.ExpectCmd {
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
	x.displayHelp(ctx.Result.CmdStack, len(ctx.Result.CmdStack)-1, ctx.Result.ExpectCmd)
	x.RenderComplete()

	ctx.Done(ErrorHelp)
	exit(x.HelpExitCode)
}

// HandleParseEvent implements parse extension
func (x *HelpExt) HandleParseEvent(event string, ctx *flag.ParseContext) {
	if event != flag.EvtResolveOpt || x.helpCmdAt >= 0 || ctx.Name == "" {
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
func (x *HelpExt) RegisterExt(parser *flag.Parser) {
	parser.AddParseExt(flag.EvtResolveOpt, x)
	parser.AddExecExt(x)
	x.helpCmdAt = -1
}

// RenderStart self implements HelpRender
func (x *HelpExt) RenderStart() {
	if x.Render != nil {
		x.Render.RenderStart()
	}
}

// RenderComplete self implements HelpRender
func (x *HelpExt) RenderComplete() {
	if x.Render != nil {
		x.Render.RenderComplete()
	}
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
func (x *HelpExt) RenderCommands(cmds []*flag.Command) {
	if x.Render != nil {
		x.Render.RenderCommands(cmds)
	}
}

// RenderArguments self implements HelpRender
func (x *HelpExt) RenderArguments(opts []*flag.Option) {
	if x.Render != nil {
		x.Render.RenderArguments(opts)
	}
}

// RenderOptions self implements HelpRender
func (x *HelpExt) RenderOptions(opts []*flag.Option) {
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

func (x *HelpExt) displayHelp(stack []*flag.ParsedCmd, at int, banner bool) {
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
			name := ArgDisplayName(arg)
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

	var opts []*flag.Option
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
			case flag.VarErrNoDef:
				err.Msg = "unknown option: " + err.Var.Name
			case flag.VarErrNoVal:
				if err.Var.Def.IsArg {
					err.Msg = "require argument " + ArgDisplayName(err.Var.Def)
				} else {
					err.Msg = "require option " + OptName(err.Var.Name)
				}
			case flag.VarErrBadVal:
				if err.Var.Def.IsArg {
					err.Msg = "invalid value for argument " + ArgDisplayName(err.Var.Def)
				} else {
					err.Msg = "invalid value for " + OptName(err.Var.Name)
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

func OptName(name string) string {
	if len(name) > 1 {
		return "--" + name
	}
	return "-" + name
}

func OptVarName(opt *flag.Option) string {
	if v, exist := opt.TagString(TagVar); exist && v != "" {
		return v
	}
	return strings.ToUpper(opt.Name)
}

func ArgDisplayName(arg *flag.Option) string {
	if v, exist := arg.TagString(TagVar); exist && v != "" {
		return v
	}
	name := []string{arg.Name}
	name = append(name, arg.Alias...)
	return strings.ToUpper(strings.Join(name, "|"))
}

func OptionDisplayName(opt *flag.Option) string {
	if opt.IsArg {
		return ArgDisplayName(opt)
	}
	return OptName(opt.Name)
}

func exit(code int) {
	if code >= 0 {
		os.Exit(code)
	}
}

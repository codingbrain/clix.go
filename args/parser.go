package args

import (
	"os"
	"reflect"
	"strings"
)

const (
	// VarErrNoDef means option is not defined
	VarErrNoDef = 0
	// VarErrNoVal means option is expecting a value but not specified
	VarErrNoVal = 1
	// VarErrBadVal means the value is invalid (e.g. unable to parse)
	VarErrBadVal = 2

	statePre    = "p"
	stateCmd    = "c"
	stateVal    = "v"
	stateErr    = "e"
	stateErrEnd = "x"
	stateEnd    = "d"
)

// VarError records any errors during parsing process
type VarError struct {
	Name    string
	Value   *string
	Def     *Option
	ErrType int
}

// ParsedCmd represent a Command which is being parsed or parsed in stack
type ParsedCmd struct {
	Cmd        *Command
	Args       []string
	ParsedArgC int
	Vars       map[string]interface{}
	Errs       []*VarError
}

// ParseResult represent the result of parsing process
type ParseResult struct {
	Program      string
	CmdStack     []*ParsedCmd
	UnparsedArgs []string
	MissingCmd   bool
	Error        error

	exts []ExecExt
}

// Parser is the state machine for parsing args
type Parser struct {
	rootCmd *Command
	result  ParseResult
	currCmd *ParsedCmd
	state   string

	// the saved context for an option waiting a value
	option   *Option
	optName  string
	stackPos int

	// extensions
	exts map[string][]ParseExt
}

func newParsedCmd(cmd *Command) *ParsedCmd {
	pcmd := &ParsedCmd{Cmd: cmd}
	pcmd.Vars = make(map[string]interface{})
	cmd.DefaultVars(pcmd.Vars)
	return pcmd
}

func (pcmd *ParsedCmd) varError(err *VarError) {
	for _, e := range pcmd.Errs {
		if e.Name == err.Name &&
			e.Def == err.Def &&
			e.ErrType == err.ErrType {
			return
		}
	}
	pcmd.Errs = append(pcmd.Errs, err)
}

func (pcmd *ParsedCmd) varNoDef(name string) {
	pcmd.varError(&VarError{Name: name, ErrType: VarErrNoDef})
}

func (pcmd *ParsedCmd) varNoVal(name string, opt *Option) {
	pcmd.varError(&VarError{Name: name, Def: opt, ErrType: VarErrNoVal})
}

func (pcmd *ParsedCmd) varBadVal(opt *Option, val *string) {
	pcmd.varError(&VarError{Name: opt.Name, Def: opt, Value: val, ErrType: VarErrBadVal})
}

func (pcmd *ParsedCmd) verifyRequiredOpts() {
	for _, opt := range pcmd.Cmd.Options {
		if !opt.Required {
			continue
		}
		if _, exists := pcmd.Vars[opt.Name]; !exists {
			pcmd.varNoVal(opt.Name, opt)
		}
	}
}

func (pcmd *ParsedCmd) verifyRequiredArgs() {
	for i, arg := range pcmd.Cmd.Arguments {
		if i < len(pcmd.Args) {
			continue
		}
		if arg.Required {
			pcmd.varNoVal(arg.Name, arg)
			pcmd.Args = append(pcmd.Args, "")
		} else {
			pcmd.Args = append(pcmd.Args, arg.DefaultAsString())
		}
	}
}

func (pcmd *ParsedCmd) hasSubCommands() bool {
	return len(pcmd.Cmd.Commands) > 0
}

func (pcmd *ParsedCmd) startSubCommand(name string) *ParsedCmd {
	if cmd := pcmd.Cmd.FindCommand(name); cmd != nil {
		return newParsedCmd(cmd)
	}
	return nil
}

func (pcmd *ParsedCmd) assignOption(opt *Option, val string, valNot bool) (parsedVal interface{}, err error) {
	if parsedVal, err = opt.ParseStrVal(val); err != nil {
		pcmd.varBadVal(opt, &val)
		return
	} else if opt.ValueKind == reflect.Map {
		var destMap map[string]interface{}
		if dictVal, exists := pcmd.Vars[opt.Name]; exists {
			if dict, ok := dictVal.(map[string]interface{}); ok {
				destMap = dict
			}
		}
		if destMap == nil {
			destMap = make(map[string]interface{})
		}
		parsedMap := parsedVal.(map[string]interface{})
		for k, v := range parsedMap {
			destMap[k] = v
		}
		parsedVal = destMap
	} else {
		if opt.ValueKind == reflect.Bool && valNot {
			parsedVal = !parsedVal.(bool)
		}
		if opt.List {
			var destList *[]interface{}
			if listVal, exists := pcmd.Vars[opt.Name]; exists {
				if list, ok := listVal.([]interface{}); ok {
					destList = &list
				}
			}
			if destList == nil {
				destList = &[]interface{}{}
			}
			parsedVal = append(*destList, parsedVal)
		}
	}
	pcmd.Vars[opt.Name] = parsedVal
	return
}

// Parser creates a parser with current command as root command
func (cmd *Command) Parser() *Parser {
	return &Parser{
		rootCmd: cmd,
		state:   statePre,
		exts:    make(map[string][]ParseExt),
	}
}

func (p *Parser) startParsing(program string) {
	p.result.Program = program
	p.state = stateCmd
	p.pushCommand(newParsedCmd(p.rootCmd))
}

func (p *Parser) findOption(name string) (*Option, int) {
	for i := len(p.result.CmdStack) - 1; i >= 0; i-- {
		if opt := p.result.CmdStack[i].Cmd.FindOption(name); opt != nil {
			return opt, i
		}
	}
	return nil, -1
}

func (p *Parser) stackAt(at int) *ParsedCmd {
	return p.result.CmdStack[at]
}

func (p *Parser) invokeExts(event string, ctx *ParseContext) {
	ctx.parser = p
	for _, ext := range p.exts[event] {
		if ctx.stopped {
			break
		}
		if ext != nil {
			ext.HandleParseEvent(event, ctx)
		}
	}
	if ctx.abortErr != nil {
		p.result.Error = ctx.abortErr
		p.state = stateErrEnd
	}
}

func (p *Parser) pushCommand(pcmd *ParsedCmd) {
	p.result.CmdStack = append(p.result.CmdStack, pcmd)
	p.currCmd = pcmd
	p.invokeExts(EvtStartCmd, &ParseContext{})
	for k, v := range pcmd.Vars {
		ctx := &ParseContext{
			OptionAt: len(p.result.CmdStack) - 1,
			Option:   pcmd.Cmd.FindOptArg(k),
			Name:     k,
			Assigned: v,
		}
		p.invokeExts(EvtAssigned, ctx)
	}
}

func (p *Parser) assignOption(at int, opt *Option, val string, valNot bool) {
	ctx := &ParseContext{
		OptionAt: at,
		Option:   opt,
		Name:     opt.Name,
		Value:    &val,
		Not:      valNot,
	}
	p.invokeExts(EvtAssignOpt, ctx)
	if ctx.Value != nil {
		if val, err := p.stackAt(at).assignOption(opt, *ctx.Value, ctx.Not); err == nil {
			ctx.Assigned = val
			p.invokeExts(EvtAssigned, ctx)
		}
	}
}

func (p *Parser) pushArg(arg string) {
	pcmd := p.currCmd
	at := len(pcmd.Args)
	pcmd.Args = append(pcmd.Args, arg)
	if at < len(pcmd.Cmd.Arguments) {
		pcmd.ParsedArgC++
		p.assignOption(len(p.result.CmdStack)-1, pcmd.Cmd.Arguments[at], arg, false)
	}
}

func (p *Parser) resolveUnknownOption(name string, val *string) {
	ctx := &ParseContext{Name: name, Value: val}
	p.invokeExts(EvtResolveOpt, ctx)
	if !ctx.Ignore {
		p.currCmd.varNoDef(name)
	}
}

func (p *Parser) resolveUnknownCommand(cmd string) {
	p.result.MissingCmd = true
	p.result.UnparsedArgs = []string{cmd}
	p.state = stateErr
}

func (p *Parser) parseArg(arg string) {
	switch p.state {
	case statePre:
		p.startParsing(arg)
	case stateCmd:
		if arg == "--" {
			p.state = stateEnd
		} else if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			var val *string
			if pos := strings.IndexByte(name, '='); pos == 0 {
				p.currCmd.varNoDef(arg)
				return
			} else if pos > 0 {
				valStr := name[pos+1:]
				val = &valStr
				name = name[0:pos]
			}

			opt, at := p.findOption(name)
			valNot := false
			if opt == nil {
				// if prefixed with "--no-", try to find a bool option
				if strings.HasPrefix(arg, "--no-") {
					name = name[3:]
					opt, at = p.findOption(name)
					if opt != nil && opt.ValueKind == reflect.Bool {
						valNot = true
					} else {
						name = arg[2:]
						opt = nil
					}
				}
			}
			if opt == nil {
				p.resolveUnknownOption(name, val)
			} else if val != nil {
				// option with a value --flag=VALUE
				p.assignOption(at, opt, *val, valNot)
			} else if opt.ValueKind == reflect.Bool {
				// bool option without a value --flag or --no-flag (valNot=true)
				p.assignOption(at, opt, "true", valNot)
			} else {
				// non-bool long option always require a value --flag=VALUE
				p.stackAt(at).varNoVal(name, opt)
			}
		} else if strings.HasPrefix(arg, "-") {
			for i, nameRune := range arg[1:] {
				name := string(nameRune)
				var val *string
				if i+1 < len(arg)-1 {
					valStr := arg[i+2:]
					val = &valStr
				}
				opt, at := p.findOption(name)
				if opt == nil {
					p.resolveUnknownOption(name, val)
				} else if opt.ValueKind == reflect.Bool {
					// for bool, -f indicate true
					p.assignOption(at, opt, "true", false)
				} else if val != nil {
					// for non-bool, -fVALUE, the rest is value
					p.assignOption(at, opt, *val, false)
					break
				} else {
					// for non-bool, -f VALUE is expected
					p.optName = name
					p.option = opt
					p.stackPos = at
					p.state = stateVal
				}
			}
		} else if p.currCmd.hasSubCommands() {
			pcmd := p.currCmd.startSubCommand(arg)
			if pcmd != nil {
				p.pushCommand(pcmd)
			} else {
				p.resolveUnknownCommand(arg)
			}
		} else {
			p.pushArg(arg)
		}
	case stateVal:
		p.assignOption(p.stackPos, p.option, arg, false)
		p.state = stateCmd
	case stateEnd:
		p.pushArg(arg)
		p.result.UnparsedArgs = append(p.result.UnparsedArgs, arg)
	case stateErr:
		if arg == "--" {
			p.state = stateErrEnd
		} else {
			p.result.UnparsedArgs = append(p.result.UnparsedArgs, arg)
		}
	case stateErrEnd:
		p.result.UnparsedArgs = append(p.result.UnparsedArgs, arg)
	}
}

func (p *Parser) parseEnd() error {
	if p.state == statePre {
		// nothing parsed
		return ErrArgsTooFew
	}
	if p.state == stateVal {
		p.stackAt(p.stackPos).varNoVal(p.optName, p.option)
	}
	for _, pcmd := range p.result.CmdStack {
		pcmd.verifyRequiredOpts()
	}
	if p.state == stateCmd || p.state == stateEnd {
		p.currCmd.verifyRequiredArgs()
	}
	return p.result.Error
}

// ParseArgs parses a slice of arguments including args[0]
func (p *Parser) ParseArgs(args []string) *ParseResult {
	for _, arg := range args {
		p.parseArg(arg)
	}
	p.result.Error = p.parseEnd()
	return &p.result
}

// Parse parses args from os.Args
func (p *Parser) Parse() *ParseResult {
	return p.ParseArgs(os.Args)
}

// Use registers an extension to current parser
func (p *Parser) Use(extReg ExtRegistrar) *Parser {
	extReg.RegisterExt(p)
	return p
}

// AddParseExt registers a parser extension
func (p *Parser) AddParseExt(event string, ext ParseExt) *Parser {
	p.exts[event] = append(p.exts[event], ext)
	return p
}

// AddExecExt registers an execution extension to ParseResult
func (p *Parser) AddExecExt(ext ExecExt) *Parser {
	p.result.AddExt(ext)
	return p
}

// AddExt registers an execution extension
func (r *ParseResult) AddExt(ext ExecExt) *ParseResult {
	r.exts = append(r.exts, ext)
	return r
}

// HasErrors indicates any errors in ParseResult
func (r *ParseResult) HasErrors() bool {
	if r.Error != nil || r.MissingCmd {
		return true
	}
	for _, pcmd := range r.CmdStack {
		if len(pcmd.Errs) > 0 {
			return true
		}
	}
	return false
}

// Exec applies the execution extensions
func (r *ParseResult) Exec() error {
	ctx := &ExecContext{Result: r}
	for _, ext := range r.exts {
		if ctx.completed {
			break
		}
		ext.ExecuteCmd(ctx)
	}
	return r.Error
}

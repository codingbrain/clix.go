package golang

import (
	"github.com/codingbrain/clix.go/flag"
	"github.com/codingbrain/clix.go/gen"
)

const (
	// BackendName is the name of the backend
	BackendName = "clix.go"
)

// Parameter names
const (
	ParamPackage = "package"
	ParamFactory = "factory"
	ParamVar     = "var"
)

// Default values
const (
	DefaultPackage = "main"
	DefaultFactory = "cliDef"
)

const (
	headerFormat = `// THIS FILE IS AUTO-GENERATED, DO NOT EDIT
package %s

import "github.com/codingbrain/clix.go/flag"
`
)

// ClixBackend is golang backend using clix.go
type ClixBackend struct {
	// package name
	Package string
	// factory name
	Factory string
	// exported variable name
	Var string
}

// NewClixBackend is the factory for ClixBackend
func NewClixBackend(params gen.BackendParams) (gen.Backend, error) {
	b := &ClixBackend{}
	b.Package, _ = params[ParamPackage].(string)
	b.Factory, _ = params[ParamFactory].(string)
	b.Var, _ = params[ParamVar].(string)
	return b, nil
}

// GenerateCode implements Backend
func (b *ClixBackend) GenerateCode(def *flag.CliDef, w *gen.Writer) error {
	pkg := b.Package
	if pkg == "" {
		pkg = DefaultPackage
	}
	w.Writeln(headerFormat, pkg)

	factory := b.Factory
	if factory == "" {
		factory = DefaultFactory
	}
	varName := b.Var
	if varName != "" {
		w.Writeln("var %s = %s()", varName, factory)
		w.Writeln("")
	}
	w.Writeln("func %s() *flag.CliDef {", factory)
	w1 := w.Indent()
	w1.Writeln("d := &flag.CliDef{")
	printCommand(w1.Indent(), "Cli: ", def.Cli)
	w1.Writeln("}")
	w1.Writeln("d.Normalize()")
	w1.Writeln("return d")
	w.Writeln("}")
	w.Writeln("")

	return nil
}

func pad(str string, minLen int) string {
	diff := minLen - len(str)
	for i := 0; i < diff; i++ {
		str += " "
	}
	return str
}

func printCommand(w *gen.Writer, prefix string, cmd *flag.Command) {
	padding := 0
	if len(cmd.Alias) > 0 {
		padding = 1
	}
	if cmd.Example != "" {
		padding = 3
	}
	padding += 6

	w.Writeln(prefix + "&flag.Command{")
	w1 := w.Indent()
	w1.Writeln(pad("Name:", padding)+"%#v,", cmd.Name)
	if len(cmd.Alias) > 0 {
		w1.Writeln(pad("Alias:", padding)+"%#v,", cmd.Alias)
	}
	if cmd.Desc != "" {
		w1.Writeln(pad("Desc:", padding)+"%#v,", cmd.Desc)
	}
	if cmd.Example != "" {
		w1.Writeln(pad("Example:", padding)+"%#v,", cmd.Example)
	}
	// TODO Tags
	if len(cmd.Options) > 0 {
		printOptions(w1, "Options: ", cmd.Options)
	}
	if len(cmd.Arguments) > 0 {
		printOptions(w1, "Arguments: ", cmd.Arguments)
	}
	if len(cmd.Commands) > 0 {
		w1.Writeln("Commands: []*flag.Command{")
		w2 := w1.Indent()
		for _, c := range cmd.Commands {
			printCommand(w2, "", c)
		}
		w1.Writeln("},")
	}
	w.Writeln("},")
}

func printOptions(w *gen.Writer, prefix string, opts []*flag.Option) {
	w.Writeln(prefix + "[]*flag.Option{")
	w1 := w.Indent()
	for _, opt := range opts {
		padding := 0
		if len(opt.Alias) > 0 {
			padding = 1
		}
		if opt.Example != "" || opt.Default != nil {
			padding = 3
		}
		if opt.Required {
			padding = 4
		}
		padding += 6
		w1.Writeln("&flag.Option{")
		w2 := w1.Indent()
		w2.Writeln(pad("Name:", padding)+"%#v,", opt.Name)
		if len(opt.Alias) > 0 {
			w2.Writeln(pad("Alias:", padding)+"%#v,", opt.Alias)
		}
		if opt.Desc != "" {
			w2.Writeln(pad("Desc:", padding)+"%#v,", opt.Desc)
		}
		if opt.Example != "" {
			w2.Writeln(pad("Example:", padding)+"%#v,", opt.Example)
		}
		if opt.Type != "" {
			w2.Writeln(pad("Type:", padding)+"%#v,", opt.Type)
		}
		if opt.List {
			w2.Writeln(pad("List:", padding) + "true,")
		}
		if opt.Required {
			w2.Writeln(pad("Required:", padding) + "true,")
		}
		if opt.Default != nil {
			w2.Writeln(pad("Default:", padding)+"%#v,", opt.Default)
		}
		// TODO Tags
		w1.Writeln("},")
	}
	w.Writeln("},")
}

func init() {
	gen.BackendFactories[BackendName] = NewClixBackend
}

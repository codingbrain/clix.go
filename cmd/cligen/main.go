package main

import (
	"fmt"
	"sort"

	"github.com/codingbrain/clix.go/exts/bind"
	"github.com/codingbrain/clix.go/exts/help"
	"github.com/codingbrain/clix.go/flag"
	"github.com/codingbrain/clix.go/gen"

	_ "github.com/codingbrain/clix.go/gen/golang"
)

type genCmd struct {
	DefFile string `n:"def-file"`
	Output  string
	Backend string
	Params  map[string]interface{} `n:"define"`
}

func (c *genCmd) Execute([]string) error {
	if c.Params == nil {
		c.Params = make(map[string]interface{})
	}

	backend, err := gen.CreateBackend(c.Backend, c.Params)
	if err != nil {
		return err
	}
	if backend == nil {
		return fmt.Errorf("backend not found: %s", c.Backend)
	}

	def, err := flag.DecodeCliDefFile(c.DefFile)
	if err != nil {
		return err
	}

	w, err := gen.NewFileWriter(c.Output)
	if err != nil {
		return err
	}
	defer w.Close()

	return backend.GenerateCode(def, w)
}

type backendsCmd struct {
}

func (c *backendsCmd) Execute([]string) error {
	for _, name := range backendNames() {
		fmt.Println(name)
	}
	return nil
}

func backendNames() []string {
	names := make([]string, 0, len(gen.BackendFactories))
	for name := range gen.BackendFactories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func main() {
	cli := &flag.CliDef{
		Cli: &flag.Command{
			Name: "cligen",
			Desc: "Generate CLI code from definition file",
			Commands: []*flag.Command{
				&flag.Command{
					Name: "gen",
					Desc: "Generate code",
					Options: []*flag.Option{
						&flag.Option{
							Name:     "def-file",
							Alias:    []string{"f"},
							Desc:     "Commands definition file",
							Required: true,
						},
						&flag.Option{
							Name:  "output",
							Alias: []string{"o"},
							Desc:  "Output file",
						},
						&flag.Option{
							Name:    "backend",
							Alias:   []string{"b"},
							Desc:    "Specify the backend, use backends to list all backends",
							Default: backendNames()[0],
						},
						&flag.Option{
							Name:  "define",
							Alias: []string{"D"},
							Desc:  "Define backend specific parameters",
							Type:  "dict",
						},
					},
				},
				&flag.Command{
					Name: "backends",
					Desc: "List supported backends",
				},
			},
		},
	}
	cli.Normalize()
	cli.Use(
		bind.NewExt().
			Bind(&genCmd{}, "gen").
			Bind(&backendsCmd{}, "backends")).
		Use(help.NewExt()).
		Parse().Exec()
}

package bind

import (
	"errors"
	"testing"

	"github.com/codingbrain/clix.go/args"
	"github.com/stretchr/testify/assert"
)

var cli = createCli()

func createCli() *args.CliDef {
	cli, err := args.DecodeCliDefString(`---
cli:
    name: test
    options:
        - name: bool1
          alias: [b1]
          type: boolean
          default: true
          tags:
              bind: bool-1
        - name: bool2
          type: boolean
          default: true
        - name: str1
          type: string
          default: str1
        - name: str2
          type: string
          default: str2
        - name: int1
          type: integer
          default: 1
        - name: int2
          type: integer
          default: 2
        - name: num1
          type: number
          default: 1.1
        - name: num2
          type: number
          default: 2.1
        - name: num3
          type: number
          default: 3.1
        - name: list1
          type: string
          list: true
          default: [a1]
        - name: dict1
          type: dict
          default:
              a: a1
    arguments:
        - name: arg1
          type: string
          tags:
              bind: false
        - name: arg2
          type: string
          tags:
              bind: '-'
        - name: arg3
          type: string
          tags:
              bind: arg
`)
	if err != nil {
		panic(err)
	}
	return cli
}

type testBind1 struct {
	Bool_1  bool
	Bool2   bool
	String1 string `n:"str1,x"`
	Str2    string `json:",x"`
	Int1    int
	Int2    uint
	Num1    float32
	Num2    float64
	Num3    complex64
	List1   []string
	Dict1   map[string]interface{}
	Arg1    string
	Arg2    string
	Arg     string
}

func TestStructBind(t *testing.T) {
	a := assert.New(t)
	s := &testBind1{}
	err := cli.Cli.Parser().
		Use(NewExt().Bind(s)).
		ParseArgs([]string{"test", "a1", "a2", "a3"}).
		Exec()
	if a.NoError(err) {
		a.Equal(true, s.Bool_1)
		a.Equal(true, s.Bool2)
		a.Equal("str1", s.String1)
		a.Equal("str2", s.Str2)
		a.EqualValues(1, s.Int1)
		a.EqualValues(2, s.Int2)
		a.EqualValues(1.1, s.Num1)
		a.EqualValues(2.1, s.Num2)
		a.EqualValues(complex(3.1, 0), s.Num3)
		a.Len(s.List1, 1)
		a.Equal("a1", s.List1[0])
		a.Len(s.Dict1, 1)
		a.Contains(s.Dict1, "a")
		a.Equal("a1", s.Dict1["a"])
		a.Empty(s.Arg1)
		a.Empty(s.Arg2)
		a.Equal("a3", s.Arg)
	}
}

type testBind2 struct {
	Str2  *string
	List1 *[]string
}

type testBind3 struct {
	List1 []*string
}

func TestStructBindPtr(t *testing.T) {
	a := assert.New(t)
	s := &testBind2{}
	err := cli.Cli.Parser().
		Use(NewExt().Bind(s)).
		ParseArgs([]string{"test"}).
		Exec()
	if a.NoError(err) {
		if a.NotNil(s.Str2) {
			a.Equal("str2", *s.Str2)
		}
		if a.NotNil(s.List1) {
			a.Len(*s.List1, 1)
			a.Equal("a1", (*s.List1)[0])
		}
	}

	s1 := &testBind3{}
	err = cli.Cli.Parser().
		Use(NewExt().Bind(s1)).
		ParseArgs([]string{"test"}).
		Exec()
	if a.NoError(err) {
		a.NotEmpty(s1.List1)
		a.NotNil(s1.List1[0])
		a.Equal("a1", *s1.List1[0])
	}
}

type testBindBytes struct {
	Str2 *[]byte
}

func TestStructBindBytes(t *testing.T) {
	a := assert.New(t)
	s := &testBindBytes{}
	err := cli.Cli.Parser().
		Use(NewExt().Bind(s)).
		ParseArgs([]string{"test"}).
		Exec()
	if a.NoError(err) {
		if a.NotNil(s.Str2) {
			a.Equal("str2", string(*s.Str2))
		}
	}
}

type testBindSubCmd struct {
	Opt string
}

func TestBindSubCmd(t *testing.T) {
	a := assert.New(t)
	cli, err := args.DecodeCliDefString(`---
        cli:
            name: test
            commands:
                - name: cmd
                  options:
                      - name: opt
                        type: string
    `)
	if a.NoError(err) {
		s := &testBindSubCmd{}
		err = cli.Cli.Parser().
			Use(NewExt().Bind(s, "cmd")).
			ParseArgs([]string{"test", "cmd", "--opt=ok"}).
			Exec()
		if a.NoError(err) {
			a.Equal("ok", s.Opt)
		}
	}
}

type testBindExec struct {
	Opt  string
	val  string
	args []string
}

func (t *testBindExec) Execute(args []string) error {
	t.val = t.Opt
	t.args = args
	return errors.New("ok")
}

func TestBindExec(t *testing.T) {
	a := assert.New(t)
	cli, err := args.DecodeCliDefString(`---
        cli:
            name: test
            options:
                - name: opt
                  type: string
    `)
	if a.NoError(err) {
		s := &testBindExec{}
		err = cli.Cli.Parser().
			Use(NewExt().Bind(s)).
			ParseArgs([]string{"test", "--opt=opt", "a1", "a2"}).
			Exec()
		if a.Error(err) {
			a.Equal("opt", s.val)
			if a.Len(s.args, 2) {
				a.Equal("a1", s.args[0])
				a.Equal("a2", s.args[1])
			}
			a.Equal("ok", err.Error())
		}
	}
}

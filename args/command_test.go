package args

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionType(t *testing.T) {
	a := assert.New(t)
	_, err := DecodeCmdsString(`---
        name: cmd
        options:
          - name: opt
            type: wrong-type
    `)
	a.Error(err)
}

func TestNames(t *testing.T) {
	a := assert.New(t)
	_, err := DecodeCmdsString(`---
        name: cmd
        options:
            - type: string
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        options:
            - name: a
              alias: [a1]
              type: string
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        options:
            - name: a
              type: string
            - name: a
              type: string
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        arguments:
            - name: a
              type: string
            - name: a
              type: string
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        options:
            - name: a
              type: string
        arguments:
            - name: a
              type: string
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        commands:
            - name: sub
            - name: sub
    `)
	a.Error(err)

	_, err = DecodeCmdsString(`---
        name: cmd
        commands:
            - description: something
    `)
	a.Error(err)
}

func TestTags(t *testing.T) {
	a := assert.New(t)
	cmd, err := DecodeCmdsString(`---
        name: cmd
        options:
            - name: opt
              type: string
              tags:
                  str: str
                  b: true
            - name: notag
              type: string
        tags:
            str: str
            b: true
    `)
	if a.NoError(err) && a.NotNil(cmd) {
		if str, ok := cmd.TagString("str"); a.True(ok) {
			a.Equal("str", str)
		}
		if b, ok := cmd.TagBool("b"); a.True(ok) {
			a.True(b)
		}
		opt := cmd.FindOption("opt")
		if str, ok := opt.TagString("str"); a.True(ok) {
			a.Equal("str", str)
		}
		if b, ok := opt.TagBool("b"); a.True(ok) {
			a.True(b)
		}
		_, ok := cmd.TagString("non-exist")
		a.False(ok)
		_, ok = cmd.TagBool("non-exist")
		a.False(ok)
		_, ok = opt.TagString("non-exist")
		a.False(ok)
		_, ok = opt.TagBool("non-exist")
		a.False(ok)
		_, ok = opt.TagBool("str")
		a.False(ok)
		_, ok = opt.TagString("b")
		a.False(ok)

		opt = cmd.FindOption("notag")
		_, ok = opt.TagString("str")
		a.False(ok)
		_, ok = opt.TagBool("b")
		a.False(ok)
	}
}

func TestNotList(t *testing.T) {
	a := assert.New(t)
	cmd, err := DecodeCmdsString(`---
        name: cmd
        options:
          - name: l1
            type: dict
            list: true
        arguments:
          - name: arg
            type: string
            list: true
    `)
	if a.NoError(err) && a.NotNil(cmd) {
		arg := cmd.FindArgument("arg")
		a.NotNil(arg)
		a.True(arg.IsArg)
		a.False(arg.List)
		a.Equal(1, arg.Position)

		arg = cmd.FindOption("l1")
		a.NotNil(arg)
		a.False(arg.List)
	}
}

func TestDefValuesMap(t *testing.T) {
	a := assert.New(t)
	cmd, err := DecodeCmdsString(`---
        name: cmd
        options:
        - name: dict1
          type: dict
          default:
            a: 1
            b: true
    `)
	if a.NoError(err) {
		dict, ok := cmd.DefVars["dict1"].(map[string]interface{})
		a.True(ok)
		a.EqualValues(1, dict["a"])
		a.True(dict["b"].(bool))
	}
}

func TestDefValues(t *testing.T) {
	a := assert.New(t)
	cmd, err := DecodeCmdsString(`---
        name: cmd
        options:
          - name: str-int
            type: string
            default: 1
          - name: str-float
            type: string
            default: 1.1
          - name: str-bool
            type: string
            default: true
          - name: int-str
            type: integer
            default: "1"
          - name: int-int
            type: integer
            default: 2
          - name: float-str
            type: number
            default: "1.1"
          - name: float-int
            type: number
            default: 1
          - name: float-float
            type: number
            default: 2.1
          - name: bool-int
            type: bool
            default: 1
          - name: bool-str
            type: bool
            default: "true"
          - name: bool-bool
            type: bool
            default: true
          - name: dict-str
            type: dict
            default: a=a1
          - name: dict1
            type: dict
            default:
              a: 1
              b: true
          - name: list-str
            type: string
            list: true
            default: a
          - name: list1
            type: string
            list: true
            default:
                - 1.1
                - 1
                - true
    `)
	if a.NoError(err) && a.NotNil(cmd) {
		a.Equal("1", cmd.DefVars["str-int"])
		a.Equal("1.1", cmd.DefVars["str-float"])
		a.Equal("true", cmd.DefVars["str-bool"])
		a.EqualValues(1, cmd.DefVars["int-str"])
		a.EqualValues(2, cmd.DefVars["int-int"])
		a.EqualValues(1.1, cmd.DefVars["float-str"])
		a.EqualValues(1.0, cmd.DefVars["float-int"])
		a.EqualValues(2.1, cmd.DefVars["float-float"])
		a.True(cmd.DefVars["bool-int"].(bool))
		a.True(cmd.DefVars["bool-str"].(bool))
		a.True(cmd.DefVars["bool-bool"].(bool))
		if dict, ok := cmd.DefVars["dict-str"].(map[string]interface{}); a.True(ok) {
			a.Equal("a1", dict["a"])
		}
		if dict, ok := cmd.DefVars["dict1"].(map[string]interface{}); a.True(ok) {
			a.EqualValues(1, dict["a"])
			a.True(dict["b"].(bool))
		}
		if slice, ok := cmd.DefVars["list-str"].([]interface{}); a.True(ok) {
			a.Len(slice, 1)
			a.Equal("a", slice[0])
		}
		if slice, ok := cmd.DefVars["list1"].([]interface{}); a.True(ok) {
			a.Len(slice, 3)
			a.Equal("1.1", slice[0])
			a.Equal("1", slice[1])
			a.Equal("true", slice[2])
		}

		a.Equal("1", cmd.FindOption("str-int").DefaultAsString())
		a.Equal("2", cmd.FindOption("int-int").DefaultAsString())
		a.Equal("2.1", cmd.FindOption("float-float").DefaultAsString())
		a.Equal("true", cmd.FindOption("bool-bool").DefaultAsString())
		a.Empty(cmd.FindOption("list1").DefaultAsString())
		a.Empty(cmd.FindOption("dict1").DefaultAsString())
	}
}

func TestEmptyList(t *testing.T) {
	a := assert.New(t)
	cmd, err := DecodeCmdsString(`---
        name: cmd
        options:
          - name: l1
            type: string
            list: true
    `)
	if a.NoError(err) && a.NotNil(cmd) {
		a.Equal([]interface{}{}, cmd.DefVars["l1"])
	}
}

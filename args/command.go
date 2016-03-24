package args

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/codingbrain/clix.go/clix"
)

type Option struct {
	Name     string                 `yaml:"name,omitempty"`
	Alias    []string               `yaml:"alias,omitempty"`
	Desc     string                 `yaml:"description,omitempty"`
	Example  string                 `yaml:"example,omitempty"`
	Type     string                 `yaml:"type,omitempty"`
	Required bool                   `yaml:"required,omitempty"`
	Default  interface{}            `yaml:"default,omitempty"`
	List     bool                   `yaml:"list,omitempty"`
	Tags     map[string]interface{} `yaml:"tags,omitempty"`

	IsArg     bool         `yaml:"-"`
	Position  int          `yaml:"-"`
	SubType   string       `yaml:"-"`
	ValueKind reflect.Kind `yaml:"-"`
}

type Command struct {
	Name      string                 `yaml:"name"`
	Alias     []string               `yaml:"alias,omitempty"`
	Desc      string                 `yaml:"description,omitempty"`
	Example   string                 `yaml:"example,omitempty"`
	Options   []*Option              `yaml:"options,omitempty"`
	Arguments []*Option              `yaml:"arguments,omitempty"`
	Commands  []*Command             `yaml:"commands,omitempty"`
	Tags      map[string]interface{} `yaml:"tags,omitempty"`

	OptMap  map[string]*Option     `yaml:"-"`
	ArgMap  map[string]*Option     `yaml:"-"`
	CmdMap  map[string]*Command    `yaml:"-"`
	DefVars map[string]interface{} `yaml:"-"`
}

type CmdDefError struct {
	Command string
	Message string
}

func (e *CmdDefError) Error() string {
	return "Command Definition Error: " + e.Command + ": " + e.Message
}

func tagString(tags map[string]interface{}, name string) (string, bool) {
	if tags == nil {
		return "", false
	}
	if val, exists := tags[name]; !exists {
		return "", false
	} else if str, ok := val.(string); ok {
		return str, true
	}
	return "", false
}

func tagBool(tags map[string]interface{}, name string) (bool, bool) {
	if tags == nil {
		return false, false
	}
	if val, exists := tags[name]; !exists {
		return false, false
	} else if boolVal, ok := val.(bool); ok {
		return boolVal, true
	}
	return false, false
}

func (opt *Option) TagString(name string) (string, bool) {
	return tagString(opt.Tags, name)
}

func (opt *Option) TagBool(name string) (bool, bool) {
	return tagBool(opt.Tags, name)
}

func (opt *Option) ParseStrVal(val string) (interface{}, error) {
	switch opt.ValueKind {
	case reflect.String:
		return val, nil
	case reflect.Bool:
		return strconv.ParseBool(val)
	case reflect.Int64:
		return strconv.ParseInt(val, 0, 64)
	case reflect.Float64:
		return strconv.ParseFloat(val, 64)
	case reflect.Map:
		if pos := strings.IndexByte(val, '='); pos == 0 {
			return nil, errors.New(errMsgNameEmpty)
		} else if pos > 0 {
			return map[string]interface{}{val[0:pos]: val[pos+1:]}, nil
		} else {
			return map[string]interface{}{val: true}, nil
		}
	}
	panic(errMsgInvalidType + opt.ValueKind.String())
}

func (opt *Option) DefaultAsString() string {
	if opt.Default == nil || opt.List || opt.ValueKind == reflect.Map {
		return ""
	}
	return fmt.Sprintf("%v", opt.Default)
}

func (opt *Option) ExpectValue() bool {
	return opt.IsArg || opt.ValueKind != reflect.Bool
}

func (opt *Option) defError(cmdPath, msg string) *CmdDefError {
	return &CmdDefError{cmdPath + "[" + opt.Name + "]", msg}
}

func (opt *Option) normalizeType(cmdPath string) error {
	if pos := strings.IndexByte(opt.Type, '/'); pos > 0 {
		opt.SubType = opt.Type[pos+1:]
		opt.Type = opt.Type[0:pos]
	}
	switch opt.Type {
	case "string", "str", "text", "":
		opt.ValueKind = reflect.String
	case "integer", "int":
		opt.ValueKind = reflect.Int64
	case "number":
		opt.ValueKind = reflect.Float64
	case "boolean", "bool":
		opt.ValueKind = reflect.Bool
	case "map", "dict":
		opt.ValueKind = reflect.Map
		opt.List = false
	default:
		return opt.defError(cmdPath, errMsgInvalidType+opt.Type)
	}
	return nil
}

func (opt *Option) normalizeAsOption(cmdPath string) error {
	if opt.Name == "" {
		return opt.defError(cmdPath, errMsgNameEmpty)
	}
	if len(opt.Name) == 1 {
		for _, a := range opt.Alias {
			if len(a) > 1 {
				return opt.defError(cmdPath, errMsgNameTooShort)
			}
		}
	}
	if err := opt.normalizeType(cmdPath); err != nil {
		return err
	}
	return nil
}

func (opt *Option) normalizeAsArgument(cmdPath string, position int) error {
	if err := opt.normalizeType(cmdPath); err != nil {
		return err
	}
	opt.IsArg = true
	opt.List = false            // for arguments, list should always be false
	opt.Position = position + 1 // position starts from 1
	return nil
}

func scalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	}
	return false
}

func intVal(val interface{}) (int64, bool) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	}
	return 0, false
}

func uintVal(val interface{}) (uint64, bool) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint(), true
	}
	return 0, false
}

func floatVal(val interface{}) (float64, bool) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	}
	return 0, false
}

func parseNotSlice(kind reflect.Kind, val interface{}) (interface{}, error) {
	// same scalar type can be passed through, map should be handled specially
	// in some cases, the map type is like map[interface{}]interface{}
	// what is wanted here is map[string]interface{}
	if k := reflect.ValueOf(val).Kind(); k == kind && k != reflect.Map {
		return val, nil
	}
	switch kind {
	case reflect.String:
		if k := reflect.ValueOf(val).Kind(); scalarKind(k) {
			return fmt.Sprintf("%v", val), nil
		}
	case reflect.Bool:
		if str, ok := val.(string); ok {
			return strconv.ParseBool(str)
		} else if intVal, ok := intVal(val); ok {
			return intVal != 0, nil
		} else if uintVal, ok := uintVal(val); ok {
			return uintVal != 0, nil
		}
	case reflect.Int64:
		if str, ok := val.(string); ok {
			return strconv.ParseInt(str, 0, 64)
		} else if intVal, ok := intVal(val); ok {
			return intVal, nil
		} else if uintVal, ok := uintVal(val); ok {
			return int64(uintVal), nil
		}
	case reflect.Float64:
		if str, ok := val.(string); ok {
			return strconv.ParseFloat(str, 64)
		} else if floatVal, ok := floatVal(val); ok {
			return floatVal, nil
		} else if intVal, ok := intVal(val); ok {
			return float64(intVal), nil
		} else if uintVal, ok := uintVal(val); ok {
			return float64(int64(uintVal)), nil
		}
	case reflect.Map:
		if str, ok := val.(string); ok {
			if pos := strings.IndexByte(str, '='); pos == 0 {
				return nil, errors.New(errMsgNameEmpty)
			} else if pos > 0 {
				return map[string]interface{}{str[0:pos]: str[pos+1:]}, nil
			} else {
				return map[string]interface{}{str: true}, nil
			}
		} else if sv := reflect.ValueOf(val); sv.Kind() == reflect.Map {
			des := make(map[string]interface{})
			for _, kv := range sv.MapKeys() {
				vv := sv.MapIndex(kv)
				if !kv.CanInterface() || !vv.CanInterface() {
					return nil, errors.New(errMsgInvalidType + sv.Kind().String())
				}
				key := fmt.Sprintf("%v", kv.Interface())
				des[key] = vv.Interface()
			}
			return des, nil
		}
	}
	return nil, errors.New(errMsgInvalidType + reflect.ValueOf(val).Kind().String())
}

func (opt *Option) parseDefaultVal() (interface{}, error) {
	if !opt.List {
		return parseNotSlice(opt.ValueKind, opt.Default)
	}
	rv := reflect.ValueOf(opt.Default)
	if kind := rv.Kind(); scalarKind(kind) {
		parsedVal, err := parseNotSlice(opt.ValueKind, opt.Default)
		if err != nil {
			return nil, err
		}
		return []interface{}{parsedVal}, nil
	} else if kind == reflect.Array || kind == reflect.Slice {
		list := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			src := rv.Index(i)
			if !src.CanInterface() {
				return nil, errors.New(errMsgInvalidType + src.Kind().String())
			}
			parsedVal, err := parseNotSlice(opt.ValueKind, src.Interface())
			if err != nil {
				return nil, err
			}
			list[i] = parsedVal
		}
		return list, nil
	} else {
		return nil, errors.New(errMsgInvalidType + kind.String())
	}
}

func (opt *Option) defaultVar(cmdPath string, vars map[string]interface{}) error {
	var v interface{}
	if opt.Required {
		return nil
	} else if opt.Default != nil {
		val, err := opt.parseDefaultVal()
		if err != nil {
			return opt.defError(cmdPath, "invalid default value: "+err.Error())
		}
		v = val
	} else if opt.List {
		v = []interface{}{}
	} else {
		switch opt.ValueKind {
		case reflect.String:
			v = ""
		case reflect.Bool:
			v = false
		case reflect.Int64:
			v = int64(0)
		case reflect.Float64:
			v = float64(0)
		case reflect.Map:
			v = make(map[string]interface{})
		default:
			panic(errMsgInvalidType + opt.ValueKind.String())
		}
	}
	vars[opt.Name] = v
	return nil
}

func indexOpt(cmdPath string, optMap, map1 map[string]*Option, opt *Option) error {
	names := append([]string{opt.Name}, opt.Alias...)
	for _, name := range names {
		if name == "" {
			continue
		}
		_, exists := optMap[name]
		if !exists {
			_, exists = map1[name]
		}
		if exists {
			return opt.defError(cmdPath, errMsgDupName)
		}
		optMap[name] = opt
	}
	return nil
}

func indexCmd(cmdPath string, cmdMap map[string]*Command, cmd *Command) error {
	names := append([]string{cmd.Name}, cmd.Alias...)
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, exists := cmdMap[name]; exists {
			return &CmdDefError{cmdPath + "/" + cmd.Name, errMsgDupName}
		}
		cmdMap[name] = cmd
	}
	return nil
}

func (cmd *Command) FindCommand(name string) *Command {
	return cmd.CmdMap[name]
}

func (cmd *Command) FindOption(name string) *Option {
	return cmd.OptMap[name]
}

func (cmd *Command) FindArgument(name string) *Option {
	return cmd.ArgMap[name]
}

func (cmd *Command) FindOptArg(name string) *Option {
	opt := cmd.FindOption(name)
	if opt == nil {
		opt = cmd.FindArgument(name)
	}
	return opt
}

func (cmd *Command) TagString(name string) (string, bool) {
	return tagString(cmd.Tags, name)
}

func (cmd *Command) TagBool(name string) (bool, bool) {
	return tagBool(cmd.Tags, name)
}

func (cmd *Command) DefaultVars(vars map[string]interface{}) {
	for k, v := range cmd.DefVars {
		if dict, ok := v.(map[string]interface{}); ok {
			dest := make(map[string]interface{})
			for name, val := range dict {
				dest[name] = val
			}
			vars[k] = dest
		} else if slice, ok := v.([]interface{}); ok {
			vars[k] = append([]interface{}{}, slice...)
		} else {
			vars[k] = v
		}
	}
}

func (cmd *Command) normalizeAsCommand(cmdPath string) error {
	if cmd.Name == "" {
		return &CmdDefError{cmdPath, errMsgNameEmpty}
	}
	if cmdPath != "" {
		cmdPath += "/"
	}
	cmdPath += cmd.Name
	cmd.OptMap = make(map[string]*Option)
	cmd.ArgMap = make(map[string]*Option)
	cmd.CmdMap = make(map[string]*Command)
	cmd.DefVars = make(map[string]interface{})
	errs := &clix.AggregatedError{}
	for _, opt := range cmd.Options {
		if errs.Add(opt.normalizeAsOption(cmdPath)) {
			continue
		}
		if errs.Add(indexOpt(cmdPath, cmd.OptMap, cmd.ArgMap, opt)) {
			continue
		}
		errs.Add(opt.defaultVar(cmdPath, cmd.DefVars))
	}
	for i, arg := range cmd.Arguments {
		if errs.Add(arg.normalizeAsArgument(cmdPath, i)) {
			continue
		}
		if errs.Add(indexOpt(cmdPath, cmd.ArgMap, cmd.OptMap, arg)) {
			continue
		}
		errs.Add(arg.defaultVar(cmdPath, cmd.DefVars))
	}
	for _, sub := range cmd.Commands {
		if errs.Add(sub.normalizeAsCommand(cmdPath)) {
			continue
		}
		errs.Add(indexCmd(cmdPath, cmd.CmdMap, sub))
	}
	return errs.Aggregate()
}

func (cmd *Command) Normalize() error {
	return cmd.normalizeAsCommand("")
}

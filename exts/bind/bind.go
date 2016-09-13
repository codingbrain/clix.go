package bind

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/codingbrain/clix.go/flag"
)

var (
	fieldTags  = []string{"n", "k", "key", "map", "flag", "json", "yaml"}
	optionTag  = "bind"
	execMethod = "Execute"
)

type BindExt struct {
	b map[string]*binding
}

type Executable interface {
	Execute([]string) error
}

type binding struct {
	update  modelUpdateFn
	execCmd execCmdFn
}

type modelUpdateFn func(opt *flag.Option, name string, value interface{})
type execCmdFn func([]string) error
type fieldUpdateFn func(value interface{})
type valueUpdateFn func(v *reflect.Value, value interface{})

func NewExt() *BindExt {
	return &BindExt{b: make(map[string]*binding)}
}

func (x *BindExt) Bind(model interface{}, cmds ...string) *BindExt {
	v := reflect.Indirect(reflect.ValueOf(model))
	switch k := v.Kind(); k {
	case reflect.Struct:
		x.b[cmdsToKey(cmds)] = makeBinding(model, structUpdateFn(&v))
	default:
		panic("Model type not supported: " + k.String())
	}
	return x
}

func (x *BindExt) HandleParseEvent(event string, ctx *flag.ParseContext) {
	if event != flag.EvtAssigned {
		return
	}
	prefix := keyFromStack(ctx.CmdStack())
	for k, b := range x.b {
		if prefix == "" || k == prefix || strings.HasPrefix(k, prefix+" ") {
			b.update(ctx.Option, ctx.Name, ctx.Assigned)
		}
	}
}

func (x *BindExt) ExecuteCmd(ctx *flag.ExecContext) {
	b, exists := x.b[keyFromStack(ctx.Result.CmdStack)]
	if exists && b.execCmd != nil && !ctx.HasErrors() {
		if err := b.execCmd(ctx.Cmd().Args); err != nil {
			ctx.Result.Error = err
		} else {
			ctx.Done(nil)
		}
	}
}

func (x *BindExt) RegisterExt(parser *flag.Parser) {
	parser.AddParseExt(flag.EvtAssigned, x)
	parser.AddExecExt(x)
}

func cmdsToKey(cmds []string) string {
	return strings.Join(cmds, " ")
}

func keyFromStack(cmdStack []*flag.ParsedCmd) string {
	key := ""
	for i, pcmd := range cmdStack {
		// skip root command
		if i > 0 {
			if key != "" {
				key += " "
			}
			key += pcmd.Cmd.Name
		}
	}
	return key
}

func makeBinding(model interface{}, update modelUpdateFn) *binding {
	b := &binding{update: update}
	if executable, ok := model.(Executable); ok {
		b.execCmd = func(args []string) error {
			return executable.Execute(args)
		}
	}
	return b
}

func fieldMappingKey(f reflect.StructField) (key string) {
	if name := f.Name; name[0] >= 'A' && name[0] <= 'Z' {
		for _, tag := range fieldTags {
			vals := strings.Split(f.Tag.Get(tag), ",")
			if len(vals) > 0 && vals[0] != "-" {
				if vals[0] != "" {
					key = vals[0]
				}
				break
			}
		}
		if key == "" {
			key = strings.Replace(strings.ToLower(f.Name), "_", "-", -1)
		}
	}
	return
}

func panicBadKind(kind reflect.Kind) {
	panic("invalid value type: " + kind.String())
}

func panicBadType(val interface{}) {
	panicBadKind(reflect.TypeOf(val).Kind())
}

func sliceUpdater(v *reflect.Value, value interface{}) {
	kind := v.Type().Elem().Kind()
	if kind == reflect.Uint8 {
		if str, ok := value.(string); ok {
			v.SetBytes([]byte(str))
			return
		}
	}
	if t := reflect.TypeOf(value); t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		if t.Elem().Kind() == kind {
			reflect.Copy(*v, reflect.ValueOf(value))
		} else if fn := valueUpdateFactory(v.Type().Elem()); fn != nil {
			vals := reflect.ValueOf(value)
			if v.Cap() >= vals.Len() {
				v.SetLen(vals.Len())
			} else {
				v.Set(reflect.MakeSlice(v.Type(), vals.Len(), vals.Len()))
			}
			for i := 0; i < vals.Len(); i++ {
				src := vals.Index(i)
				if !src.CanInterface() {
					panicBadKind(src.Kind())
				}
				des := v.Index(i)
				fn(&des, src.Interface())
			}
			return
		} else {
			panicBadKind(t.Elem().Kind())
		}
	}
	panicBadType(value)
}

func mapUpdater(v *reflect.Value, value interface{}) {
	if sv := reflect.ValueOf(value); sv.Kind() == reflect.Map {
		des := make(map[string]interface{})
		for _, kv := range sv.MapKeys() {
			vv := sv.MapIndex(kv)
			if !kv.CanInterface() {
				panicBadKind(kv.Kind())
			}
			if !vv.CanInterface() {
				panicBadKind(vv.Kind())
			}
			key := fmt.Sprintf("%v", kv.Interface())
			des[key] = vv.Interface()
		}
		v.Set(reflect.ValueOf(des))
	} else {
		panicBadType(value)
	}
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

func scalarUpdateFactory(t reflect.Type) valueUpdateFn {
	switch t.Kind() {
	case reflect.Bool, reflect.String:
		return func(v *reflect.Value, value interface{}) {
			v.Set(reflect.ValueOf(value))
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return func(v *reflect.Value, value interface{}) {
			if int64Val, ok := intVal(value); ok {
				v.SetInt(int64Val)
			} else if uint64Val, ok := uintVal(value); ok {
				v.SetInt(int64(uint64Val))
			} else {
				panicBadType(value)
			}
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return func(v *reflect.Value, value interface{}) {
			if uint64Val, ok := uintVal(value); ok {
				v.SetUint(uint64Val)
			} else if int64Val, ok := intVal(value); ok {
				v.SetUint(uint64(int64Val))
			} else {
				panicBadType(value)
			}
		}
	case reflect.Float32, reflect.Float64:
		return func(v *reflect.Value, value interface{}) {
			if float64Val, ok := floatVal(value); ok {
				v.SetFloat(float64Val)
			} else if int64Val, ok := intVal(value); ok {
				v.SetFloat(float64(int64Val))
			} else if uint64Val, ok := uintVal(value); ok {
				v.SetFloat(float64(int64(uint64Val)))
			} else {
				panicBadType(value)
			}
		}
	case reflect.Complex64, reflect.Complex128:
		return func(v *reflect.Value, value interface{}) {
			if float64Val, ok := floatVal(value); ok {
				v.SetComplex(complex(float64Val, 0))
			} else if int64Val, ok := intVal(value); ok {
				v.SetComplex(complex(float64(int64Val), 0))
			} else if uint64Val, ok := uintVal(value); ok {
				v.SetComplex(complex(float64(int64(uint64Val)), 0))
			} else {
				panicBadType(value)
			}
		}
	}
	return nil
}

func valueUpdateFactory(t reflect.Type) valueUpdateFn {
	if fn := scalarUpdateFactory(t); fn != nil {
		return fn
	} else {
		switch t.Kind() {
		case reflect.Array, reflect.Slice:
			return sliceUpdater
		case reflect.Map:
			return mapUpdater
		case reflect.Ptr:
			if fn := valueUpdateFactory(t.Elem()); fn != nil {
				return func(v *reflect.Value, value interface{}) {
					ptr := reflect.New(t.Elem())
					val := reflect.Indirect(ptr)
					fn(&val, value)
					v.Set(ptr)
				}
			}
		}
	}
	return nil
}

func fieldUpdateFactory(v *reflect.Value) fieldUpdateFn {
	if fn := valueUpdateFactory(v.Type()); fn != nil {
		return func(value interface{}) {
			fn(v, value)
		}
	} else {
		return func(value interface{}) {
			v.Set(reflect.ValueOf(value))
		}
	}
}

func structUpdateFn(model *reflect.Value) modelUpdateFn {
	t := model.Type()
	mapper := make(map[string]fieldUpdateFn)
	for i := 0; i < t.NumField(); i++ {
		if key := fieldMappingKey(t.Field(i)); key != "" {
			v := model.Field(i)
			mapper[key] = fieldUpdateFactory(&v)
		}
	}
	return func(opt *flag.Option, name string, value interface{}) {
		if opt == nil {
			return
		}
		if val, ok := opt.TagBool(optionTag); ok && !val {
			// bind disabled
			return
		} else if bindKey, ok := opt.TagString(optionTag); ok && bindKey != "" {
			if bindKey == "-" {
				// bind disabled
				return
			}
			name = bindKey
		} else {
			name = opt.Name
		}
		if fn, ok := mapper[name]; ok {
			fn(value)
		}
		return
	}
}

// struct based option parse library.
// use field type and tag to specify property, e.g.:
//
// type option struct {
// 	OptionA string        `usage:"" default:""`
// 	OptionB int           `usage:"an int option" default:"10"`
// 	OptionC time.Duration `usage:"base type is int64, so we need derive here" derive:"time.Duration" default:"1h1m"`
// 	OptionD bool          `usage:"can only be used in command line" nocfg:"true" default:"false"`
// }
package opt

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

const (
	Sep = '.'
)

// Opt is a underlying structer for parser.
type Opt struct {
	f           *flag.FlagSet
	nocfg       map[string]struct{}
	initialized bool
}

// NewOpt create a new option parser context. The argument needs to be
// a struct describing each option.
func NewOpt(s interface{}) *Opt {
	o := &Opt{
		f:     flag.NewFlagSet(path.Base(os.Args[0]), flag.ExitOnError),
		nocfg: make(map[string]struct{}),
	}
	o.init(s, "")
	o.initialized = true
	return o
}

// Parse deal with command line arguments. Most common use is
// Parse(os.Args[1:]).
func (o *Opt) Parse(arg []string) error {
	if !o.initialized {
		return errors.New("not initialized")
	}
	return o.f.Parse(arg)
}

// Defaults print all the options as while as their default value as
// the format of loadable configuration file to stdout.
func (o *Opt) Defaults() string {
	if !o.initialized {
		return "not initialized"
	}

	b := &bytes.Buffer{}

	b.WriteString(fmt.Sprintf(
		"# auto generated configuration file for %s\n\n",
		path.Base(os.Args[0])))

	f := func(f *flag.Flag) {
		if _, ok := o.nocfg[f.Name]; ok {
			return
		}
		b.WriteString(fmt.Sprintf("# %s\n", f.Usage))
		b.WriteString(fmt.Sprintf("%s = %s\n", f.Name, f.DefValue))
	}
	o.f.VisitAll(f)
	return b.String()
}

// Load reads option from config file. The format of this file is:
//   # comment started by hash tag
//   key = value
func (o *Opt) Load(fname string) error {
	if !o.initialized {
		return errors.New("not initialized")
	}

	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		line := strings.TrimLeft(s.Text(), " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.Trim(parts[0], " '\"")
		key = strings.ToLower(key)
		value := strings.Trim(parts[1], " '\"")

		// ignore cli only options
		if _, ok := o.nocfg[key]; ok {
			continue
		}

		v := o.f.Lookup(key)
		if v == nil {
			return fmt.Errorf("flag provided but not defined: %s", key)
		}

		if err := v.Value.Set(value); err != nil {
			return err
		}
	}
	return nil
}

func (o *Opt) init(des interface{}, prefix string) {
	var v reflect.Value
	for {
		var ok bool
		if v, ok = des.(reflect.Value); ok {
			break
		}
		des = reflect.ValueOf(des)
	}

	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("invalid argument")
	}

	for i := 0; i < v.NumField(); i++ {
		item := v.Field(i)
		field := v.Type().Field(i)
		if !item.CanSet() {
			continue
		}

		//fmt.Printf("%#v\n", field)
		name := normalize(field.Name)
		if prefix != "" {
			name = fmt.Sprintf("%s.%s", prefix, name)
		}

		usage := field.Tag.Get("usage")
		derive := field.Tag.Get("derive")
		def := field.Tag.Get("default")
		nocfg := field.Tag.Get("nocfg")

		ptr := unsafe.Pointer(item.UnsafeAddr())
		switch item.Kind() {
		case reflect.Bool:
			o.f.BoolVar((*bool)(ptr), name, item.Bool(), usage)
		case reflect.Int:
			o.f.IntVar((*int)(ptr), name, int(item.Int()), usage)
		case reflect.Int64:
			if derive == "time.Duration" {
				o.f.DurationVar((*time.Duration)(ptr), name, time.Duration(item.Int()), usage)
			} else {
				o.f.Int64Var((*int64)(ptr), name, item.Int(), usage)
			}
		case reflect.Int8, reflect.Int16, reflect.Int32:
			panic("not implemented")
		case reflect.Uint:
			o.f.UintVar((*uint)(ptr), name, uint(item.Uint()), usage)
		case reflect.Uint64:
			o.f.Uint64Var((*uint64)(ptr), name, item.Uint(), usage)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32:
			panic("not implemented")
		case reflect.Float32:
			panic("not implemented")
		case reflect.Float64:
			o.f.Float64Var((*float64)(ptr), name, item.Float(), usage)
		case reflect.String:
			o.f.StringVar((*string)(ptr), name, item.String(), usage)
		case reflect.Struct:
			o.init(item, name)
		default:
		}

		if def != "" {
			if f := o.f.Lookup(name); f != nil {
				f.DefValue = def
				o.f.Set(name, def)
			}
		}
		if nocfg == "true" || nocfg == "1" {
			o.nocfg[name] = struct{}{}
		}
	}
}

func normalize(s string) string {
	if s == "" {
		return s
	}

	var b []rune
	for i, c := range s {
		switch {
		case i == 0:
			b = append(b, c)
		case c == '_' || c == '.':
			b = append(b, Sep)
		case unicode.IsUpper(c):
			b = append(b, Sep)
			b = append(b, c)
		default:
			b = append(b, c)
		}
	}
	return strings.ToLower(string(b))
}

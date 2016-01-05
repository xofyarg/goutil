// Package opt uses struct based option parse library which use field
// type and tag to specify property.
//
// Supported base types:
//     bool, int, int64, uint, uint64, float64, string, struct
//
// Supported inferred types:
//     time.Duration
//
// Supported tags:
//   usage:     Message shows in help and config comment.
//   default:   String represented default value.
//   name:      Symbol to use instead of string inferred from field
//              name. Rule to generate name is adding "Sep" char
//              between CamelCase names or replace "Breaker" with "Sep".
//   cli:       Option can only be used in command line, will not
//              load from/dump into config files.
//
// example usage:
//   type myOption struct {
//       AString string      `usage:"a simple option" default:"hello"`
//       MyIP string         `name:"my.ip" usage:"overwrite default name"`
//       OptionC bool        `usage:"can only be used in command line" cli:"true" default:"false"`
//   }
//
//   c := new(myOption)
//   o, err := opt.New(c)
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   // load option from command line
//   o.Parse(os.Args[1:])
//
//   // or load from config file
//   o.Load("~/.my.conf")
//
//   // dump config to stdout
//   fmt.Println(o.Defaults())
//
//   // access config values
//   fmt.Printf("a.string: %s\n", c.AString)
//
package opt

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

const (
	// Sep is the separator in option name
	Sep = '.'
	// Breaker is a list of chars to break words
	Breaker = "_."
)

// Opt is a underlying structure for parser.
type Opt struct {
	f           *flag.FlagSet
	cli         map[string]struct{}
	initialized bool
}

// New create a new option parser context. The argument needs to be
// a struct describing each option.
func New(s interface{}) (*Opt, error) {
	o := &Opt{
		f:   flag.NewFlagSet(path.Base(os.Args[0]), flag.ExitOnError),
		cli: make(map[string]struct{}),
	}
	err := o.init(s, "")
	if err != nil {
		return nil, err
	}
	o.initialized = true
	return o, nil
}

// Parse deal with command line arguments. Most common use is
// Parse(os.Args[1:]).
func (o *Opt) Parse(arg []string) error {
	if !o.initialized {
		return errors.New("not initialized")
	}
	return o.f.Parse(arg)
}

// Args returns the non-flag arguments from underlying flagset.
func (o *Opt) Args() []string {
	return o.f.Args()
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
		if _, ok := o.cli[f.Name]; ok {
			return
		}
		if f.Usage != "" {
			b.WriteString(fmt.Sprintf("# %s\n", f.Usage))
		}
		b.WriteString(fmt.Sprintf("%s = %s\n", f.Name, f.DefValue))
	}
	o.f.VisitAll(f)
	return b.String()
}

// Load reads option from config file. The format of this file is:
//   # comment started by hash tag
//   key = value
func (o *Opt) Load(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return o.LoadConfig(f, true)
}

// LoadConfig works like Load if overwrite is true, otherwise, it ignore
// the options which already has value other than default.
func (o *Opt) LoadConfig(f io.Reader, overwrite bool) error {
	if !o.initialized {
		return errors.New("not initialized")
	}

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
		if _, ok := o.cli[key]; ok {
			continue
		}

		v := o.f.Lookup(key)
		if v == nil {
			return fmt.Errorf("flag provided but not defined: %s", key)
		}

		// ignore already set option(has value other than default)
		if !overwrite && v.Value.String() != v.DefValue {
			continue
		}

		if err := v.Value.Set(value); err != nil {
			return err
		}
	}
	return nil
}

func (o *Opt) init(des interface{}, prefix string) error {
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
		return errors.New("invalid argument, option description needs to be a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		item := v.Field(i)
		field := v.Type().Field(i)
		if !item.CanSet() {
			continue
		}

		// parse tags
		name := field.Tag.Get("name")
		usage := field.Tag.Get("usage")
		def := field.Tag.Get("default")
		cli := field.Tag.Get("cli")
		if cli == "" {
			// Compatible with legacy name
			cli = field.Tag.Get("nocfg")
		}

		if name == "" {
			name = normalize(field.Name)
		}

		if prefix != "" {
			name = fmt.Sprintf("%s.%s", prefix, name)
		}

		ptr := unsafe.Pointer(item.UnsafeAddr())
		switch item.Kind() {
		case reflect.Bool:
			o.f.BoolVar((*bool)(ptr), name, item.Bool(), usage)
		case reflect.Int:
			o.f.IntVar((*int)(ptr), name, int(item.Int()), usage)
		case reflect.Int64:
			switch item.Type().String() {
			case "time.Duration":
				o.f.DurationVar((*time.Duration)(ptr), name, time.Duration(item.Int()), usage)
			default:
				o.f.Int64Var((*int64)(ptr), name, item.Int(), usage)
			}
		case reflect.Uint:
			o.f.UintVar((*uint)(ptr), name, uint(item.Uint()), usage)
		case reflect.Uint64:
			o.f.Uint64Var((*uint64)(ptr), name, item.Uint(), usage)
		case reflect.Float64:
			o.f.Float64Var((*float64)(ptr), name, item.Float(), usage)
		case reflect.String:
			o.f.StringVar((*string)(ptr), name, item.String(), usage)
		case reflect.Struct:
			o.init(item, name)
		case reflect.Int8, reflect.Int16, reflect.Int32:
			fallthrough
		case reflect.Uint8, reflect.Uint16, reflect.Uint32:
			fallthrough
		case reflect.Float32:
			fallthrough
		default:
			return fmt.Errorf("parsing of type %s(%s) not implemented", item.Type(), item.Kind())
		}

		if def != "" {
			if f := o.f.Lookup(name); f != nil {
				f.DefValue = def
				o.f.Set(name, def)
			}
		}
		if cli == "true" || cli == "1" {
			o.cli[name] = struct{}{}
		}
	}
	return nil
}

func normalize(s string) string {
	if s == "" {
		return s
	}

	var b []rune
	var cont bool
	for i, c := range s {
		switch {
		case i == 0:
			b = append(b, c)
		case strings.ContainsRune(Breaker, c):
			if !cont {
				b = append(b, Sep)
				cont = true
			}
		case unicode.IsUpper(c):
			if !cont {
				b = append(b, Sep)
			}
			b = append(b, c)
			cont = false
		default:
			b = append(b, c)
			cont = false
		}
	}
	return strings.ToLower(string(b))
}

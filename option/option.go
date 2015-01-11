// Extend package flag with internal storage and generate config file
// support.
//
// Motivation:
//   1. option should be declared once, then used everywhere with
//      clear reference.
//   2. to prevent create _hidden_ options, there should be a way to
//      list all supported options with description.
//
package option

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

// struct to hold internal context
type Option struct {
	name  string
	set   *flag.FlagSet
	store map[string]interface{}
	cli   map[string]struct{}
}

// initialization helper
func NewOption(name string) *Option {
	return &Option{
		name:  name,
		set:   flag.NewFlagSet(name, flag.ExitOnError),
		store: make(map[string]interface{}),
		cli:   make(map[string]struct{}),
	}
}

func (o *Option) Bool(name string, value bool, usage string) {
	o.store[name] = o.set.Bool(name, value, usage)
}

func (o *Option) GetBool(name string) bool {
	return *o.store[name].(*bool)
}

func (o *Option) Duration(name string, value time.Duration, usage string) {
	o.store[name] = o.set.Duration(name, value, usage)
}

func (o *Option) GetDuration(name string) time.Duration {
	return *o.store[name].(*time.Duration)
}

func (o *Option) Float64(name string, value float64, usage string) {
	o.store[name] = o.set.Float64(name, value, usage)
}

func (o *Option) GetFloat64(name string) float64 {
	return *o.store[name].(*float64)
}

func (o *Option) Int(name string, value int, usage string) {
	o.store[name] = o.set.Int(name, value, usage)
}

func (o *Option) GetInt(name string) int {
	return *o.store[name].(*int)
}

func (o *Option) Uint(name string, value uint, usage string) {
	o.store[name] = o.set.Uint(name, value, usage)
}

func (o *Option) GetUint(name string) uint {
	return *o.store[name].(*uint)
}

func (o *Option) Int64(name string, value int64, usage string) {
	o.store[name] = o.set.Int64(name, value, usage)
}

func (o *Option) GetInt64(name string) int64 {
	return *o.store[name].(*int64)
}

func (o *Option) Uint64(name string, value uint64, usage string) {
	o.store[name] = o.set.Uint64(name, value, usage)
}

func (o *Option) GetUint64(name string) uint64 {
	return *o.store[name].(*uint64)
}

func (o *Option) String(name string, value string, usage string) {
	o.store[name] = o.set.String(name, value, usage)
}

func (o *Option) GetString(name string) string {
	return *o.store[name].(*string)
}

// parse option from args
func (o *Option) Parse(args []string) error {
	return o.set.Parse(args)
}

// load option from config file. The format of this file is:
//   # comment started by hash tag
//   key = value
func (o *Option) LoadConfig(name string) error {
	f, err := os.Open(name)
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
		if _, ok := o.cli[key]; ok {
			continue
		}

		if _, ok := o.store[key]; !ok {
			msg := fmt.Sprintf("flag provided but not defined: %s", key)
			return errors.New(msg)
		}

		if err := o.set.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

// mark args only available in command line interface
func (o *Option) CliOnly(keys []string) {
	for _, k := range keys {
		o.cli[k] = struct{}{}
	}
}

// dump out default config file
func (o *Option) Defaults() string {
	b := &bytes.Buffer{}

	b.WriteString(fmt.Sprintf(
		"# auto generated configuration file for profile %s\n\n",
		o.name))

	f := func(f *flag.Flag) {
		if _, ok := o.cli[f.Name]; ok {
			return
		}
		b.WriteString(fmt.Sprintf("# %s\n", f.Usage))
		b.WriteString(fmt.Sprintf("%s = %s\n", f.Name, f.DefValue))
	}
	o.set.VisitAll(f)
	return b.String()
}

// dump current option set with their values converted to string
func (o *Option) All() map[string]string {
	m := make(map[string]string)

	f := func(f *flag.Flag) {
		if _, ok := o.cli[f.Name]; ok {
			return
		}
		m[f.Name] = f.Value.String()
	}
	o.set.VisitAll(f)
	return m
}

// global default option context
var Default = NewOption(path.Base(os.Args[0]))

func Bool(name string, value bool, usage string) {
	Default.store[name] = Default.set.Bool(name, value, usage)
}

func GetBool(name string) bool {
	return *Default.store[name].(*bool)
}

func Duration(name string, value time.Duration, usage string) {
	Default.store[name] = Default.set.Duration(name, value, usage)
}

func GetDuration(name string) time.Duration {
	return *Default.store[name].(*time.Duration)
}

func Float64(name string, value float64, usage string) {
	Default.store[name] = Default.set.Float64(name, value, usage)
}

func GetFloat64(name string) float64 {
	return *Default.store[name].(*float64)
}

func Int(name string, value int, usage string) {
	Default.store[name] = Default.set.Int(name, value, usage)
}

func GetInt(name string) int {
	return *Default.store[name].(*int)
}

func Uint(name string, value uint, usage string) {
	Default.store[name] = Default.set.Uint(name, value, usage)
}

func GetUint(name string) uint {
	return *Default.store[name].(*uint)
}

func Int64(name string, value int64, usage string) {
	Default.store[name] = Default.set.Int64(name, value, usage)
}

func GetInt64(name string) int64 {
	return *Default.store[name].(*int64)
}

func Uint64(name string, value uint64, usage string) {
	Default.store[name] = Default.set.Uint64(name, value, usage)
}

func GetUint64(name string) uint64 {
	return *Default.store[name].(*uint64)
}

func String(name string, value string, usage string) {
	Default.store[name] = Default.set.String(name, value, usage)
}

func GetString(name string) string {
	return *Default.store[name].(*string)
}

func Parse() error {
	return Default.set.Parse(os.Args[1:])
}

func LoadConfig(name string) error {
	return Default.LoadConfig(name)
}

// mark args only available in command line interface
func CliOnly(keys []string) {
	Default.CliOnly(keys)
}

// dump out default config file
func Defaults() string {
	return Default.Defaults()
}

func All() map[string]string {
	return Default.All()
}

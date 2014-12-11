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

type Options struct {
	name  string
	set   *flag.FlagSet
	store map[string]interface{}
	cli   map[string]struct{}
}

func NewOption(name string) *Options {
	return &Options{
		name:  name,
		set:   flag.NewFlagSet(name, flag.ExitOnError),
		store: make(map[string]interface{}),
		cli:   make(map[string]struct{}),
	}
}

func (o *Options) Bool(name string, value bool, usage string) {
	o.store[name] = o.set.Bool(name, value, usage)
}

func (o *Options) GetBool(name string) bool {
	return *o.store[name].(*bool)
}

func (o *Options) Duration(name string, value time.Duration, usage string) {
	o.store[name] = o.set.Duration(name, value, usage)
}

func (o *Options) GetDuration(name string) time.Duration {
	return *o.store[name].(*time.Duration)
}

func (o *Options) Float64(name string, value float64, usage string) {
	o.store[name] = o.set.Float64(name, value, usage)
}

func (o *Options) GetFloat64(name string) float64 {
	return *o.store[name].(*float64)
}

func (o *Options) Int(name string, value int, usage string) {
	o.store[name] = o.set.Int(name, value, usage)
}

func (o *Options) GetInt(name string) int {
	return *o.store[name].(*int)
}

func (o *Options) Uint(name string, value uint, usage string) {
	o.store[name] = o.set.Uint(name, value, usage)
}

func (o *Options) GetUint(name string) uint {
	return *o.store[name].(*uint)
}

func (o *Options) Int64(name string, value int64, usage string) {
	o.store[name] = o.set.Int64(name, value, usage)
}

func (o *Options) GetInt64(name string) int64 {
	return *o.store[name].(*int64)
}

func (o *Options) Uint64(name string, value uint64, usage string) {
	o.store[name] = o.set.Uint64(name, value, usage)
}

func (o *Options) GetUint64(name string) uint64 {
	return *o.store[name].(*uint64)
}

func (o *Options) String(name string, value string, usage string) {
	o.store[name] = o.set.String(name, value, usage)
}

func (o *Options) GetString(name string) string {
	return *o.store[name].(*string)
}

func (o *Options) Parse(args []string) error {
	return o.set.Parse(args)
}

func (o *Options) LoadConfig(name string) error {
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
func (o *Options) CliOnly(keys []string) {
	for _, k := range keys {
		o.cli[k] = struct{}{}
	}
}

// dump out default config file
func (o *Options) Defaults() string {
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

func Parse(args []string) error {
	return Default.set.Parse(args)
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

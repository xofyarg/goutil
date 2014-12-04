package option

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Options struct {
	set   *flag.FlagSet
	store map[string]interface{}
	cli   map[string]struct{}
}

func NewOpt(name string) *Options {
	return &Options{
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

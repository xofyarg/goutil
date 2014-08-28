// Config:
//
// Damn simple config file parser.
package util

import (
	"bufio"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// LoadConfig parse file and save result into config
func LoadConfig(f io.Reader, c interface{}) error {
	v := reflect.ValueOf(c)
	for {
		if v.Kind() == reflect.Struct {
			break
		}
		v = v.Elem()
	}

	keys := parseKeys(v)

	buf := bufio.NewScanner(f)

	for buf.Scan() {
		line := strings.TrimLeft(buf.Text(), " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.Trim(parts[0], " '\"")
		value := strings.Trim(parts[1], " '\"")

		if i, ok := keys[key]; !ok {
			Log(Debug, "unknown option: %s", key)
			continue
		} else {
			setKey(v, i, value)
		}
	}
	return nil
}

func parseKeys(v reflect.Value) map[string]int {
	keys := make(map[string]int)

	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		t := string(f.Tag)
		if t != "" {
			keys[t] = i
		} else {
			keys[f.Name] = i
		}
	}
	return keys
}

func setKey(v reflect.Value, i int, s string) {
	f := v.Field(i)
	if f.CanSet() {
		switch f.Kind() {
		case reflect.String:
			f.SetString(s)
		case reflect.Bool:
			f.SetBool(s == "true")
		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			n, _ := strconv.ParseInt(s, 10, 64)
			f.SetInt(n)
		case reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64:
			n, _ := strconv.ParseUint(s, 10, 64)
			f.SetUint(n)
		default:
		}
	} else {
		Log(Debug, "cannot set member: %s", v.Type().Field(i).Name)
	}
}

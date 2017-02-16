package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"

	"strings"

	"github.com/alistanis/size"
)

var (
	ErrNotAStruct    = errors.New("The type given is not a struct")
	ErrInvalidFormat = errors.New("Invalid format given: must be json or yaml")
)

const (
	Json = "json"
	Yaml = "yaml"
)

func Create(i interface{}, indent int) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}
	if indent > 0 {
		fmt.Println("Struct: ", strings.Repeat(" ", indent), v.Type().Name())

	} else {
		fmt.Println("Struct:", v.Type().Name())
	}
	indent++
	scanner := bufio.NewScanner(os.Stdin)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		n := v.Type().Field(i).Name

		switch f.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array:
			return nil, errors.New("Only primitive types are currently supported (embedded structs are supported)")
		}
		var t string
		if f.Kind() != reflect.Struct {
			fmt.Printf("%sPlease enter a value for %s (type: %s)\n%s", strings.Repeat(" ", indent), n, f.Kind().String(), strings.Repeat(" ", indent))
			scanner.Scan()
			t = scanner.Text()
		} else {
			fmt.Printf("%sEmbedded struct: %s (name: %s)\n", strings.Repeat(" ", indent), n, v.Type().Name())
		}

		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			i, err := strconv.Atoi(t)
			if err != nil {
				return nil, err
			}
			m[n] = i
		case reflect.Bool:
			b, err := strconv.ParseBool(t)
			if err != nil {
				return nil, err
			}
			m[n] = b
		case reflect.String:
			m[n] = t
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(t, size.WordBits)
			if err != nil {
				return nil, err
			}
			m[n] = f
		case reflect.Struct:
			rm, err := Create(reflect.New(f.Type()).Interface(), indent+1)
			if err != nil {
				return nil, err
			}
			m[n] = rm
		}

	}
	return m, nil
}

func GenerateData(i interface{}, format string) ([]byte, error) {
	m, err := Create(i, 0)
	if err != nil {
		return nil, err
	}
	switch format {
	case Json:
		return json.MarshalIndent(m, "", "  ")
	case Yaml:
		return yaml.Marshal(m)
	}
	return nil, ErrInvalidFormat
}

func main() {
	data, err := GenerateData(&Example{}, Json)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println(string(data))
}

type Example struct {
	Value1 string
	Value2 bool
	Value3 Embed
}

type Embed struct {
	String string
}

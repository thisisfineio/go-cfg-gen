package cfgen

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

	"io/ioutil"

	"bytes"
	"encoding/csv"
	"io"
)

var (
	ErrNotAStruct    = errors.New("The type given is not a struct")
	ErrInvalidFormat = errors.New("Invalid format given: must be json or yaml")
)

type Format int

const (
	Json Format = iota
	Yaml

	WordBits = 32 << (^uint(0) >> 63) // 64 or 32
)

var (
	// reader is the internal reader that we'll take input on, the default is os.Stdin
	reader io.Reader
	// writer is the internal writer we'll write output to, the default is os.Stdout
	writer io.Writer
)

func init() {
	reader = os.Stdin
	writer = os.Stdout
}

// SetReader sets the reader that CreateMap will use in order to take in input
// defaults to os.Stdin
func SetReader(r io.Reader) {
	reader = r
}

// SetWriter sets the writer that CreateMap will use in order to write output to
func SetWriter(w io.Writer) {
	writer = w
}

// CreateMap creates a map of the interface given (which much be a struct)
func CreateMap(i interface{}, indent int) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}
	if indent > 0 {
		fmt.Fprintln(writer,"Struct: ", strings.Repeat(" ", indent), v.Type().Name())
	} else {
		fmt.Fprintln(writer,"Struct:", v.Type().Name())
	}
	indent++
	scanner := bufio.NewScanner(reader)
	for i := 0; i < v.NumField(); i++ {
		var t string
		f := v.Field(i)
		n := v.Type().Field(i).Name

		switch f.Kind() {
		case reflect.Struct:
			fmt.Fprintf(writer, "%sEmbedded struct: %s (name: %s)\n", strings.Repeat(" ", indent+1), n, v.Type().Name())
		case reflect.Ptr:
			fmt.Fprintf(writer, "%sEmbedded pointer: %s (name: %s)\n", strings.Repeat(" ", indent+1), n, v.Type().Name())
		case reflect.Func, reflect.Uintptr, reflect.UnsafePointer, reflect.Chan:
			continue
		case reflect.Interface:
			// if this isn't the empty interface we don't want to store data in it
			if f.NumMethod() != 0 {
				continue
			}
		}

		if f.Kind() == reflect.Map || f.Kind() == reflect.Slice {
			// if we don't have a valid value type, skip it
			vtyp := reflect.New(f.Type().Elem())
			// TODO - fix - this shouldn't be here, too tired when working on this
			if !ValueTypeIsValid(vtyp, 0) {
				continue
			}
		}

		if f.Kind() != reflect.Struct && f.Kind() != reflect.Ptr {
			fmt.Fprintf(writer, "%sPlease enter a value for %s (type: %s)", strings.Repeat(" ", indent), n, f.Kind().String())
			if f.Kind() == reflect.Slice {
				fmt.Fprintf(writer, "(%s) (enter your values as a comma separated list) ex: '1,2,3', 'I love configs!' - using double quotes will ignore commas inside them, like a csv. For slices of slices, use double quotes around each slice value: ex: \"1,2,3\",\"a,b,c\"", f.Type().Elem())
			}

			if f.Kind() == reflect.Map {
				fmt.Fprintf(writer, "KeyType: %s, ValueType:%s, (enter your values as a comma separated list of key value pairs separated by a colon) ex: 'first_key:first_value,second_key:secondvalue'", f.Type().Key(), f.Type().Elem())
			}

			fmt.Fprintf(writer, "\n%s", strings.Repeat(" ", indent))
			scanner.Scan()
			t = scanner.Text()
		}


		i, err := ParseType(t, f.Type(), indent)
		if err != nil {
			return nil, err
		}
		if i != nil {
			m[n] = i
		}

	}
	return m, nil
}

// ParseType returns the actual type necessary for encoding to work successfully as an interface - recursively calls container types/structs
func ParseType(t string, typ reflect.Type, indent int) (interface{}, error) {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.Atoi(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return Uatoi(t)
	case reflect.Bool:
		return strconv.ParseBool(t)
	case reflect.String:
		return t, nil
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(t, WordBits)
	case reflect.Slice:
		return ParseSlice(t, typ.Elem(), indent)
	case reflect.Map:
		return ParseMap(t, typ, indent)
	case reflect.Struct:
		return CreateMap(reflect.New(typ).Interface(), indent+1)
	case reflect.Ptr:
		return CreateMap(reflect.New(typ.Elem()).Interface(), indent+1)
	// just return the string since we don't know what its real type is
	case reflect.Interface:
		return t, nil
	// ignore these types for now
	case reflect.Func, reflect.Uintptr, reflect.Chan, reflect.UnsafePointer:
		return nil, nil
	default:
		return nil, fmt.Errorf("cfgen: Unsupported type given to ParseType (%s)", typ.Kind())
	}
}

// ParseSlice parses each value in the string for its type and adds it to a slice of interfaces, returned itself as an interface
func ParseSlice(t string, typ reflect.Type, indent int) (interface{}, error) {
	r := bytes.NewReader([]byte(t))
	csvR := csv.NewReader(r)
	records, err := csvR.ReadAll()
	if err != nil {
		return nil, err
	}

	i := make([]interface{}, 0)
	for _, slc := range records {
		for _, s := range slc {
			t, err := ParseType(s, typ, indent)
			if err != nil {
				return nil, err
			}
			i = append(i, t)
		}
	}
	return i, nil
}

// ParseMap parses a string for key/value pairs, creates a map of the appropriate type, and returns the map as an interface
func ParseMap(t string, typ reflect.Type, indent int) (interface{}, error) {
	r := bytes.NewReader([]byte(t))
	csvR := csv.NewReader(r)
	records, err := csvR.ReadAll()
	if err != nil {
		return nil, err
	}
	m := reflect.MakeMap(typ)

	ktyp := typ.Key()
	vtyp := typ.Elem()

	for _, slc := range records {
		for _, s := range slc {
			// TODO - fix this, this is bad and will break if there are any colons inside of a string
			kvslc := strings.Split(s, ":")
			if len(kvslc) != 2 {
				return nil, fmt.Errorf("cfgen: Missing full k/v pair for map, got %d of 2 entries", len(kvslc))
			}
			k, err := ParseType(kvslc[0], ktyp, indent)
			if err != nil {
				return nil, err
			}

			t, err := ParseType(kvslc[1], vtyp, indent)
			if err != nil {
				return nil, err
			}
			m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(t))
		}
	}
	return m.Interface(), nil
}

// ValueTypeIsValid returns whether or not the map value type is valid for our purposes
func ValueTypeIsValid(v reflect.Value, depth int) (valid bool) {

	if depth > 1 {
		if v.Kind() == reflect.Slice {
			return false
		}
	}

	switch v.Kind() {
	case reflect.Uintptr, reflect.Chan, reflect.Struct, reflect.Invalid:
		return false
	case reflect.Ptr:
		v = reflect.Indirect(v)
		return ValueTypeIsValid(v, depth + 1)
	}
	return true
}

// Generate data creates a map of the interface i, and marshals it into the given format (yaml or json)
func GenerateData(i interface{}, format Format) ([]byte, error) {
	m, err := CreateMap(i, 0)
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

// GenerateAndSave generates config data in the given format and saves it to the path given
func GenerateAndSave(i interface{}, format Format, path string) error {
	data, err := GenerateData(i, format)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// Uatoi converts a string to a uint
func Uatoi(s string) (uint, error) {
	ui64, err := strconv.ParseUint(s, 10, 0)
	return uint(ui64), err
}

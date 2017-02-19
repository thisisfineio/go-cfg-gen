package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thisisfineio/go-cfg-gen"
)

var (
	genConfig bool
)

func init() {
	flag.BoolVar(&genConfig, "g", false, "Use this flag in order to generate a config for your ")
	flag.Parse()
}

type TestNonEmptyInterface interface {
	TestFunction() error
}

type TestConfig struct {
	Data interface{}
	TestNonEmptyInterface
	TestString string
	TestInt int
	TestFloat float64
	TestBool bool
	TestMapStringString map[string]string
	TestMapIntString map[int]string
	TestMapStringInterface map[string]interface{}
	TestStringSlice []string
	TestIntSlice []int
	TestSliceOfStringSlices [][]string
	TestSliceOfSliceOfSlices [][][]string
	TestEmbedded TestEmbeddedStruct
	TestEmbeddedStructPtr *TestEmbeddedStruct
	TestComplexMapType map[string]*TestEmbeddedStruct
}

type TestEmbeddedStruct struct {
	EmbeddedString string
}

func main() {
	// config flag given
	if genConfig {
		// do generation (user GenerateAndSave(interface{}, format, path) to write directly to file)
		data, err := cfgen.GenerateData(TestConfig{}, cfgen.Json)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		// do something with result (save to file, etc)
		fmt.Println(string(data))
		// we're done generating the config, exit
		os.Exit(0)
	}

	// do other stuff in your application
}

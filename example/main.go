package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thisisfineio/go-cfg-gen"
)

type Config struct {
	FileType  string
	Deployers []string
}

var (
	genConfig bool
)

func init() {
	flag.BoolVar(&genConfig, "g", false, "Use this flag in order to generate a config for your ")
	flag.Parse()
}

func main() {
	// config flag given
	if genConfig {
		// do generation (user GenerateAndSave(interface{}, format, path) to write directly to file)
		data, err := cfgen.GenerateData(&Config{}, cfgen.Json)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		// do something with result (save to file, etc)
		fmt.Println(string(data))
	}
}

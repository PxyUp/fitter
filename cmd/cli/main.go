package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/registry"
	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
)

func getConfig(filePath string) *config.CliItem {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("unable to read config file %s with error %s", filePath, err.Error())
		return nil
	}
	cfg := &config.CliItem{}
	if path.Ext(filePath) == ".json" {
		err = json.Unmarshal(file, &cfg)
		if err != nil {
			log.Fatalf("unable to json unmarshal config file %s with error %s", filePath, err.Error())
			return nil
		}

		return cfg
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("unable to yaml unmarshal config file %s with error %s", filePath, err.Error())
		return nil
	}

	return cfg
}

func main() {
	filePath := flag.String("path", "config.yaml", "Path for config file yaml|json")
	copyFlag := flag.Bool("copy", false, "Copy to clip board")
	prettyFlag := flag.Bool("pretty", true, "Make result pretty")
	flag.Parse()

	cfg := getConfig(*filePath)
	name := "fitter_cli"
	cfg.Item.Name = name
	res, err := registry.FromItem(cfg).Get(name).Process()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	result := res.ToJson()
	if *prettyFlag {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(result), "", "\t")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		result = prettyJSON.String()
	}
	fmt.Fprintln(os.Stdout, result)
	if *copyFlag {
		clipboard.WriteAll(result)
	}
}

package main

import (
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

func getConfig(filePath string) *config.Item {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("unable to read config file %s with error %s", filePath, err.Error())
		return nil
	}
	cfg := &config.Item{}
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
	flag.Parse()

	cfg := getConfig(*filePath)
	name := "fitter_cli"
	cfg.Name = name
	cfg.NotifierConfig = nil
	res, err := registry.FromItem(cfg).Get(name).Process()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, res.ToJson())
	if *copyFlag {
		clipboard.WriteAll(res.ToJson())
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/atotto/clipboard"
	"github.com/tidwall/gjson"
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
	verboseFlag := flag.Bool("verbose", false, "Provide logger")
	omitPrettyErrorFlag := flag.Bool("omit-error-pretty", false, "Provide pure value if pretty is invalid")
	pluginsFlag := flag.String("plugins", "", "Provide plugins folder")
	logLevel := flag.String("log-level", "info", "Level for logger")
	inputFlag := flag.String("input", "", "Input for model")
	flag.Parse()

	log := logger.Null
	if *verboseFlag {
		log = logger.NewLogger(*logLevel)
	}

	if *pluginsFlag != "" {
		err := store.PluginInitialize(*pluginsFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}

	cfg := getConfig(*filePath)
	res, err := lib.Parse(cfg.Item, cfg.Limits, cfg.References, builder.PureString(gjson.Parse(*inputFlag).String()), log)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	result := res.ToJson()
	if *prettyFlag {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(result), "", "\t")
		if err != nil {
			if *omitPrettyErrorFlag {
				fmt.Fprintln(os.Stdout, result)
				return
			}
			fmt.Fprintln(os.Stderr, "unable prettify: ", err.Error())
			fmt.Fprintln(os.Stdout, "You can execute with --omit-error-pretty=true to get raw data")
			return
		}
		result = prettyJSON.String()
	}
	fmt.Fprintln(os.Stdout, result)
	if *copyFlag {
		_ = clipboard.WriteAll(result)
	}
}

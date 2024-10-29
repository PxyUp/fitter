package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/atotto/clipboard"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path"
)

func getConfig(filePath string, urlPath string) *config.CliItem {
	var content []byte

	if filePath != "" {
		file, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("unable to read config file %s with error %s", filePath, err.Error())
			return nil
		}

		content = file
	}

	if urlPath != "" {
		resp, err := http_client.GetDefaultClient().Get(urlPath)
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("unable to read config file %s with error %s", urlPath, err.Error())
			return nil
		}
		content = responseBody
	}

	cfg := &config.CliItem{}
	if path.Ext(filePath) == ".json" || path.Ext(urlPath) == ".json" {
		err := json.Unmarshal(content, &cfg)
		if err != nil {
			log.Fatalf("unable to json unmarshal config file %s with error %s", filePath, err.Error())
			return nil
		}

		return cfg
	}

	err := yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatalf("unable to yaml unmarshal config file %s with error %s", filePath, err.Error())
		return nil
	}

	return cfg
}

func main() {
	filePath := flag.String("path", "", "Path for config file yaml|json")
	urlPath := flag.String("url", "", "URL for path for config")
	copyFlag := flag.Bool("copy", false, "Copy to clip board")
	prettyFlag := flag.Bool("pretty", true, "Make result pretty")
	verboseFlag := flag.Bool("verbose", false, "Provide logger")
	omitPrettyErrorFlag := flag.Bool("omit-error-pretty", false, "Provide pure value if pretty is invalid")
	pluginsFlag := flag.String("plugins", "", "Provide plugins folder")
	logLevel := flag.String("log-level", "info", "Level for logger")
	inputFlag := flag.String("input", "", "Input for model")
	flag.Parse()

	if *filePath == "" && *urlPath == "" {
		log.Fatal("path or url flag is required")
		return
	}

	log := logger.Null
	if *verboseFlag {
		log = logger.NewLogger(*logLevel)
		utils.SetLogger(*logLevel)
	}

	if *pluginsFlag != "" {
		err := store.PluginInitialize(*pluginsFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}

	cfg := getConfig(*filePath, *urlPath)
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

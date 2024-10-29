package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/runtime"
	"github.com/PxyUp/fitter/pkg/utils"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"time"
)

func getConfig(filePath string, urlPath string) *config.Config {
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

	cfg := &config.Config{}
	if path.Ext(filePath) == ".json" || path.Ext(urlPath) == ".json" {
		err := json.Unmarshal(content, &cfg)
		if err != nil {
			log.Fatalf("unable to json unmarshal config file %s with error %s", filePath, err.Error())
			return nil
		}

		if len(cfg.Items) == 0 {
			log.Fatal("empty config")
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
	verboseFlag := flag.Bool("verbose", false, "Provide logger")
	pluginsFlag := flag.String("plugins", "", "Provide plugins folder")
	logLevel := flag.String("log-level", "info", "Level for logger")
	flag.Parse()

	if *filePath == "" && *urlPath == "" {
		log.Fatal("path or url flag is required")
		return
	}

	if *pluginsFlag != "" {
		err := store.PluginInitialize(*pluginsFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}

	cfg := getConfig(*filePath, *urlPath)
	if cfg == nil {
		log.Fatalf("empty config file %s or url %s", *filePath, *urlPath)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	lg := logger.Null
	if *verboseFlag {
		lg = logger.NewLogger(*logLevel)
		utils.SetLogger(*logLevel)
	}
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		lg.Infof("Shutdown....")
		time.Sleep(time.Second * 4)
		close(done)
	}()
	runtime.New(ctx, cfg, lg.With("component", "runtime")).Start()
	<-done
}

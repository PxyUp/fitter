//go:build js

// Command wasm exposes the fitter engine to JavaScript for the browser
// playground. Build with:
//
//	GOOS=js GOARCH=wasm go build -o demo/main.wasm ./cmd/wasm
//
// It registers a global fitterRun(configJSON, input) -> Promise<string>.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"gopkg.in/yaml.v3"
)

func parseConfig(content string) (*config.CliItem, error) {
	cfg := &config.CliItem{}
	if errJson := json.Unmarshal([]byte(content), cfg); errJson == nil {
		return cfg, nil
	}

	if errYaml := yaml.Unmarshal([]byte(content), cfg); errYaml != nil {
		return nil, fmt.Errorf("config is neither valid JSON nor YAML: %s", errYaml.Error())
	}

	return cfg, nil
}

func fitterRun(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.Global().Get("Promise").Call("reject", "fitterRun(configJSON, input?) requires a config argument")
	}

	content := args[0].String()
	input := ""
	if len(args) > 1 && args[1].Type() == js.TypeString {
		input = args[1].String()
	}

	handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		// network calls block, so the run must leave the JS event loop
		go func() {
			defer func() {
				if r := recover(); r != nil {
					reject.Invoke(fmt.Sprintf("fitter panicked: %v", r))
				}
			}()

			cfg, err := parseConfig(content)
			if err != nil {
				reject.Invoke(err.Error())
				return
			}
			if cfg.Item == nil {
				reject.Invoke(`missing "item" object at the top level of the config`)
				return
			}

			res, errParse := lib.ParseCtx(context.Background(), cfg.Item, cfg.Limits, cfg.References, builder.PureString(input), logger.Null)
			if errParse != nil {
				reject.Invoke(errParse.Error())
				return
			}

			resolve.Invoke(res.ToJson())
		}()

		return nil
	})

	return js.Global().Get("Promise").New(handler)
}

func main() {
	js.Global().Set("fitterRun", js.FuncOf(fitterRun))
	js.Global().Set("fitterReady", js.ValueOf(true))
	select {}
}

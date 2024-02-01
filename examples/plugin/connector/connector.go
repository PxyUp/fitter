package main

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	pl "github.com/PxyUp/fitter/pkg/plugins/plugin"
)

var (
	_ pl.ConnectorPlugin = &plugin{}

	Plugin plugin
)

type plugin struct {
	log  logger.Logger
	Name string `json:"name" yaml:"name"`
}

func (pl *plugin) Get(parsedValue builder.Jsonable, index *uint32, input builder.Jsonable) ([]byte, error) {
	return []byte(fmt.Sprintf(`{"name": "%s"}`, pl.Name)), nil
}

func (pl *plugin) SetConfig(cfg *config.PluginConnectorConfig, logger logger.Logger) {
	pl.log = logger

	if cfg.Config != nil {
		err := json.Unmarshal(cfg.Config, pl)
		if err != nil {
			pl.log.Errorw("cant unmarshal plugin configuration", "error", err.Error())
			return
		}
	}
}

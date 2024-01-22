package main

import (
	"encoding/json"
	"fmt"
	builder "github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	pl "github.com/PxyUp/fitter/pkg/plugins/plugin"
)

var (
	_ pl.FieldPlugin = &plugin{}

	Plugin plugin
)

type plugin struct {
	Name string `json:"name" yaml:"name"`
}

func (pl *plugin) Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	if field.Config != nil {
		err := json.Unmarshal(field.Config, pl)
		if err != nil {
			logger.Errorw("cant unmarshal plugin configuration", "error", err.Error())
			return builder.NullValue
		}
		return builder.String(fmt.Sprintf("Hello %s", pl.Name))
	}

	return builder.String(fmt.Sprintf("Hello %s", parsedValue.ToJson()))
}

package main

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	pl "github.com/PxyUp/fitter/pkg/plugins/plugin"
)

var (
	_ pl.Plugin = &plugin{}

	Plugin plugin
)

type plugin struct {
}

func (pl *plugin) Format(parsedValue builder.Jsonable, field *config.GeneratedFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	return builder.String("Hello hardcoder!")
}

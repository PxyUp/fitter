package lib

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/registry"
	"github.com/google/uuid"
)

func Parse(item *config.Item, limits *config.Limits, refMap config.RefMap, input builder.Interfacable, log logger.Logger) (*parser.ParseResult, error) {
	cfg := &config.CliItem{
		Item:       item,
		Limits:     limits,
		References: refMap,
	}
	name := uuid.New().String()
	cfg.Item.Name = name
	if log == nil {
		log = logger.Null
	}
	return registry.FromItem(cfg, log).Get(name).Process(input)
}

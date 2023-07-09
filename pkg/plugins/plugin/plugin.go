package plugin

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
)

type Plugin interface {
	Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable
}

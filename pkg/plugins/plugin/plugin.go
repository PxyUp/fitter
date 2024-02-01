package plugin

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
)

type FieldPlugin interface {
	Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32, input builder.Jsonable) builder.Jsonable
}

type ConnectorPlugin interface {
	connectors.Connector

	SetConfig(cfg *config.PluginConnectorConfig, logger logger.Logger)
}

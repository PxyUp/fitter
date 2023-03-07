package connectors

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

type staticConnector struct {
	cfg    *config.StaticConnectorConfig
	logger logger.Logger
}

func (j *staticConnector) Get() ([]byte, error) {
	return []byte(j.cfg.Value), nil
}

func NewStatic(cfg *config.StaticConnectorConfig) *staticConnector {
	return &staticConnector{
		cfg: cfg,
	}
}

func (j *staticConnector) WithLogger(logger logger.Logger) *staticConnector {
	j.logger = logger
	return j
}

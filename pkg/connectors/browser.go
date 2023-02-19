package connectors

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

type browserConnector struct {
	cfg    *config.BrowserConnectorConfig
	logger logger.Logger
}

func NewBrowser(cfg *config.BrowserConnectorConfig) *browserConnector {
	return &browserConnector{
		cfg:    cfg,
		logger: logger.Null,
	}
}

func (c *browserConnector) WithLogger(logger logger.Logger) *browserConnector {
	c.logger = logger
	return c
}

func (c *browserConnector) Get() ([]byte, error) {
	if c.cfg.Chromium != nil {
		return getFromChromium(c.cfg.Url, c.cfg.Chromium, c.logger.With("emulator", "chromium"))
	}

	return nil, nil
}

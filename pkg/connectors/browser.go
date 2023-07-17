package connectors

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/PxyUp/fitter/pkg/utils"
)

type browserConnector struct {
	cfg    *config.BrowserConnectorConfig
	logger logger.Logger
	url    string
}

func NewBrowser(url string, cfg *config.BrowserConnectorConfig) *browserConnector {
	return &browserConnector{
		cfg:    cfg,
		url:    url,
		logger: logger.Null,
	}
}

func (c *browserConnector) WithLogger(logger logger.Logger) *browserConnector {
	c.logger = logger
	return c
}

func (c *browserConnector) Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error) {
	c.url = utils.Format(c.url, parsedValue, index)

	if c.url == "" {
		return nil, errEmpty
	}
	if c.cfg.Chromium != nil {
		return getFromChromium(c.url, c.cfg.Chromium, c.logger.With("emulator", "chromium"))
	}

	if c.cfg.Docker != nil {
		return getFromDocker(c.url, c.cfg.Docker, c.logger.With("emulator", "docker"))
	}

	if c.cfg.Playwright != nil {
		return getFromPlaywright(c.url, c.cfg.Playwright, c.logger.With("emulator", "playwright"))
	}

	return nil, nil
}

package connectors

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
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

func (c *browserConnector) Get(parsedValue builder.Interfacable, index *uint32, input builder.Interfacable) ([]byte, error) {
	formattedURL := utils.Format(c.url, parsedValue, index, input)

	if formattedURL == "" {
		return nil, errEmpty
	}

	if c.cfg.Chromium != nil {
		return getFromChromium(formattedURL, c.cfg.Chromium, c.logger.With("emulator", "chromium"))
	}

	if c.cfg.Docker != nil {
		return getFromDocker(formattedURL, c.cfg.Docker, c.logger.With("emulator", "docker"))
	}

	if c.cfg.Playwright != nil {
		return getFromPlaywright(formattedURL, c.cfg.Playwright, parsedValue, index, input, c.logger.With("emulator", "playwright"))
	}

	return nil, nil
}

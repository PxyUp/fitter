package connectors

import (
	"context"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"os/exec"
	"time"
)

type browserConnector struct {
	cfg    *config.BrowserConnectorConfig
	logger logger.Logger
}

func NewBrowser(cfg *config.BrowserConnectorConfig) *browserConnector {
	return &browserConnector{
		cfg: cfg,
	}
}

func (c *browserConnector) WithLogger(logger logger.Logger) *browserConnector {
	c.logger = logger
	return c
}

func (c *browserConnector) Get() ([]byte, error) {
	if c.cfg.Chromium != nil {
		return c.getFromChromium()
	}

	return nil, nil
}

func (c *browserConnector) getFromChromium() ([]byte, error) {
	ctxB := context.Background()
	if instanceLimit := limitter.ChromiumLimiter(); instanceLimit != nil {
		errInstance := instanceLimit.Acquire(ctxB, 1)
		if errInstance != nil {
			c.logger.Errorw("unable to acquire chromium limit semaphore", "url", c.cfg.Url, "error", errInstance.Error())
			return nil, errInstance
		}
		defer instanceLimit.Release(1)
	}
	t := timeout
	if c.cfg.Chromium.Timeout > 0 {
		t = time.Second * time.Duration(c.cfg.Chromium.Timeout)
	}
	ctxT, cancel := context.WithTimeout(ctxB, t)
	defer cancel()

	args := []string{
		"--headless",
		"--proxy-auto-detect",
		"--temp-profile",
		"--incognito",
		"--disable-logging",
		"--disable-gpu",
		fmt.Sprintf("--timeout=%d", t.Milliseconds()),
		"--dump-dom",
		c.cfg.Url,
	}
	if c.cfg.Chromium.Wait > 0 {
		args = append(args, fmt.Sprintf("--virtual-time-budget=%d", (time.Duration(c.cfg.Chromium.Wait)*time.Millisecond).Milliseconds()))
	}

	out, err := exec.CommandContext(ctxT, c.cfg.Chromium.Path, args...).Output()
	if err != nil {
		c.logger.Errorw("unable to get response from chromium", "url", c.cfg.Url, "error", err.Error())
		return nil, err
	}

	return out, nil
}

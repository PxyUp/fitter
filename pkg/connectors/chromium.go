package connectors

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"os/exec"
	"time"
)

var (
	defaultChromiumFlags = []string{
		"--headless",
		"--proxy-auto-detect",
		"--temp-profile",
		"--incognito",
		"--disable-logging",
		"--disable-gpu",
	}
)

func getFromChromium(url string, cfg *config.ChromiumConfig, logger logger.Logger) ([]byte, error) {
	ctxB := context.Background()

	if instanceLimit := limitter.ChromiumLimiter(); instanceLimit != nil {
		errInstance := instanceLimit.Acquire(ctxB, 1)
		if errInstance != nil {
			logger.Errorw("unable to acquire chromium limit semaphore", "url", url, "error", errInstance.Error())
			return nil, errInstance
		}
		defer instanceLimit.Release(1)
	}

	t := timeout
	if cfg.Timeout > 0 {
		t = time.Second * time.Duration(cfg.Timeout)
	}
	ctxT, cancel := context.WithTimeout(ctxB, t)
	defer cancel()

	var args []string
	if len(cfg.Flags) != 0 {
		args = append(args, cfg.Flags...)
	} else {
		args = append(args, defaultChromiumFlags...)
	}

	if cfg.Wait > 0 {
		args = append(args, fmt.Sprintf("--virtual-time-budget=%d", (time.Duration(cfg.Wait)*time.Millisecond).Milliseconds()))
	}

	args = append(args, fmt.Sprintf("--timeout=%d", t.Milliseconds()), "--dump-dom", url)

	cmd := exec.CommandContext(ctxT, cfg.Path, args...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		logger.Errorw("fatal error during chromium run", "url", url, "error", err.Error())
		return nil, err
	}

	if errb.Len() > 0 {
		logger.Errorw("error during chromium running", "url", url, "error", errb.String())
	}

	return outb.Bytes(), nil
}

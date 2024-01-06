package connectors

import (
	"context"
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/playwright-community/playwright-go"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
	"time"
)

var (
	installPlaywrightSem = semaphore.NewWeighted(1)
	installedPlaywright  = atomic.NewBool(false)
	errNoDriver          = errors.New("empty playwright driver")
)

func getFromPlaywright(url string, cfg *config.PlaywrightConfig, logger logger.Logger) ([]byte, error) {
	if instanceLimit := limitter.PlaywrightLimiter(); instanceLimit != nil {
		errInstance := instanceLimit.Acquire(ctx, 1)
		if errInstance != nil {
			logger.Errorw("unable to acquire playwright limit semaphore", "url", url, "error", errInstance.Error())
			return nil, errInstance
		}
		defer instanceLimit.Release(1)
	}

	t := timeout
	if cfg.Timeout > 0 {
		t = time.Second * time.Duration(cfg.Timeout)
	}

	ctxT, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	res := make(chan struct{})

	var content string
	var err error

	go func() {
		defer close(res)
		if cfg.Install {
			err = installPlaywrightSem.Acquire(ctxT, 1)
			if err != nil {
				logger.Errorw("unable to install playwright", "error", err.Error())
				return
			}

			func() {
				defer installPlaywrightSem.Release(1)
				if installedPlaywright.Load() {
					logger.Info("playwright already installed")
					return
				}

				logger.Info("run only one playwright installation in parallel")

				err = playwright.Install(&playwright.RunOptions{
					Verbose: false,
				})
				if err != nil {
					logger.Errorw("unable to install playwright", "error", err.Error())
					return
				}
				installedPlaywright.Store(true)
			}()
		}

		pw, err := playwright.Run(&playwright.RunOptions{
			Verbose: false,
		})
		if err != nil {
			logger.Errorw("could not start playwright", "error", err.Error())
			return
		}

		defer func() {
			if errStop := pw.Stop(); errStop != nil {
				logger.Errorw("could not stop playwright", "error", errStop.Error())
			}
		}()

		var browserInstance playwright.Browser
		if cfg.Browser == config.Chromium {
			browserInstance, err = pw.Chromium.Launch()
			if err != nil {
				logger.Errorw("could not launch Chromium", "error", err.Error())
				return
			}
		}
		if cfg.Browser == config.FireFox {
			browserInstance, err = pw.Firefox.Launch()
			if err != nil {
				logger.Errorw("could not launch Firefox", "error", err.Error())
				return
			}
		}
		if cfg.Browser == config.WebKit {
			browserInstance, err = pw.WebKit.Launch()
			if err != nil {
				logger.Errorw("could not launch WebKit", "error", err.Error())
				return
			}
		}

		if browserInstance == nil {
			err = errNoDriver
			logger.Errorw("could not launch", "error", err.Error())
			return
		}

		defer func() {
			errClose := browserInstance.Close()
			if errClose != nil {
				logger.Errorw("could not close browser", "error", errClose.Error())
			}
		}()

		page, err := browserInstance.NewPage()
		if err != nil {
			logger.Errorw("could not create page: %v", "error", err.Error())
			return
		}

		defer func() {
			errPageClose := page.Close()
			if errPageClose != nil {
				logger.Errorw("could not close page", "error", errPageClose.Error())
			}
		}()

		tt := timeout
		if cfg.Wait > 0 {
			tt = time.Second * time.Duration(cfg.Wait)
		}

		logger.Infof("going to url: %s", url)
		_, err = page.Goto(url, playwright.PageGotoOptions{
			Timeout:   playwright.Float(float64(tt.Milliseconds())),
			WaitUntil: cfg.TypeOfWait,
		})
		if err != nil {
			logger.Errorw("could not goto", "error", err.Error())
			return
		}

		if cfg.PreRunScript != "" {
			_, err = page.Evaluate(cfg.PreRunScript)
			if err != nil {
				logger.Errorw("could execute script on page", "error", err.Error(), "script", cfg.PreRunScript)
				return
			}
		}

		content, err = page.Content()
		if err != nil {
			logger.Errorw("unable to get page content", "error", err.Error())
			return
		}
	}()

	select {
	case <-res:
		return []byte(content), err
	case <-ctxT.Done():
		return nil, ctxT.Err()
	}
}

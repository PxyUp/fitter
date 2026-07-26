//go:build js

package connectors

import (
	"context"
	"errors"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

var errNotSupportedInWasm = errors.New("browser emulation is not supported in the WebAssembly build")

func getFromChromium(_ context.Context, _ string, _ *config.ChromiumConfig, logger logger.Logger) ([]byte, error) {
	logger.Errorw("chromium connector unavailable", "error", errNotSupportedInWasm.Error())
	return nil, errNotSupportedInWasm
}

func getFromDocker(_ context.Context, _ string, _ *config.DockerConfig, logger logger.Logger) ([]byte, error) {
	logger.Errorw("docker connector unavailable", "error", errNotSupportedInWasm.Error())
	return nil, errNotSupportedInWasm
}

func getFromPlaywright(_ context.Context, _ string, _ *config.PlaywrightConfig, _ builder.Interfacable, _ *uint32, _ builder.Interfacable, logger logger.Logger) ([]byte, error) {
	logger.Errorw("playwright connector unavailable", "error", errNotSupportedInWasm.Error())
	return nil, errNotSupportedInWasm
}

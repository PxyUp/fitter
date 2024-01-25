package connectors

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"io"
	"os"
)

type fileConnector struct {
	cfg    *config.FileConnectorConfig
	logger logger.Logger
}

func (j *fileConnector) Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error) {
	file, err := os.Open(utils.Format(j.cfg.Path, parsedValue, index))
	if err != nil {
		j.logger.Errorw("cant open file", "error", err.Error())
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	body, err := io.ReadAll(file)
	if err != nil {
		j.logger.Errorw("cant read file content", "error", err.Error())
		return nil, err
	}

	if !j.cfg.UseFormatting {
		return body, nil
	}

	return []byte(utils.Format(string(body), parsedValue, index)), nil
}

func NewFile(cfg *config.FileConnectorConfig) *fileConnector {
	return &fileConnector{
		cfg:    cfg,
		logger: logger.Null,
	}
}

func (j *fileConnector) WithLogger(logger logger.Logger) *fileConnector {
	j.logger = logger
	return j
}

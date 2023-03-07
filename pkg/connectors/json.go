package connectors

import (
	"encoding/json"
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

type jsonConnector struct {
	cfg    *config.JsonConnectorConfig
	logger logger.Logger
}

var (
	errNotJson = errors.New("not a json")
)

func (j *jsonConnector) Get() ([]byte, error) {
	if !isJSON(j.cfg.Json) {
		return nil, errNotJson
	}

	return []byte(j.cfg.Json), nil
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func NewJSON(cfg *config.JsonConnectorConfig) *jsonConnector {
	return &jsonConnector{
		cfg: cfg,
	}
}

func (j *jsonConnector) WithLogger(logger logger.Logger) *jsonConnector {
	j.logger = logger
	return j
}

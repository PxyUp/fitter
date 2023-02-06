package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
)

type jsonParser struct {
	logger logger.Logger
}

func NewJson() *jsonParser {
	return &jsonParser{
		logger: logger.Null,
	}
}

func (j *jsonParser) WithLogger(logger logger.Logger) {
	j.logger = logger
}

func (j *jsonParser) Parse(model config.Model) (*ParseResult, error) {

}

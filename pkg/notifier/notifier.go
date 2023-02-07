package notifier

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
)

type notifier struct {
	logger logger.Logger
	name   string
	cfg    *config.NotifierConfig
}

type Notifier interface {
	Inform(result *parser.ParseResult, err error)
}

func New(name string, cfg *config.NotifierConfig) *notifier {
	return &notifier{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}

func (o *notifier) WithLogger(logger logger.Logger) *notifier {
	o.logger = logger
	return o
}

func (o *notifier) Inform(result *parser.ParseResult, err error) {
	if o.cfg != nil && o.cfg.Console {
		if err != nil {
			o.logger.Errorf("result for %s is error: %s", o.name, err.Error())
		} else {
			o.logger.Infow("Processing done", "response", result.Raw)
		}
	}
}

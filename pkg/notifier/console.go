package notifier

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
)

type console struct {
	logger logger.Logger
	name   string
	cfg    *config.ConsoleConfig
}

var (
	_ Notifier = &console{}
)

func NewConsole(name string, cfg *config.ConsoleConfig) *console {
	return &console{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}

func (o *console) WithLogger(logger logger.Logger) *console {
	o.logger = logger
	return o
}

func (o *console) Inform(result *parser.ParseResult, errResult error, isArray bool) error {
	if errResult != nil {
		o.logger.Errorf("result for %s is error: %s", o.name, errResult.Error())
		return nil
	}

	o.logger.Infow("Processing done", "response", result.ToJson())
	return nil
}

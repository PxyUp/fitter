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

	notifierCfg *config.NotifierConfig
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

func (o *console) Inform(result *parser.ParseResult, err error, isArray bool) error {
	should, err := shouldInform(o.notifierCfg.Expression, result, o.notifierCfg.Force)
	if err != nil {
		o.logger.Errorw("unable to calculate expression for informing", "error", err.Error())
		return err
	}
	if !(should) {
		return nil
	}
	if err != nil {
		o.logger.Errorf("result for %s is error: %s", o.name, err.Error())
	} else {
		o.logger.Infow("Processing done", "response", result.ToJson())
	}
	return nil
}

func (o *console) SetConfig(cfg *config.NotifierConfig) {
	if o == nil {
		return
	}
	o.notifierCfg = cfg
}

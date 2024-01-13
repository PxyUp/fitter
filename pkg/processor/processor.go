package processor

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/notifier"
	"github.com/PxyUp/fitter/pkg/parser"
)

var (
	errEmpty       = errors.New("empty response")
	errMissingName = errors.New("missing name in configuration of the fitter")
)

type Processor interface {
	Process() (*parser.ParseResult, error)
}

type processor struct {
	logger   logger.Logger
	model    *config.Model
	notifier notifier.Notifier
	engine   parser.Engine
}

type nullProcessor struct {
	err error
}

func Null(errs ...error) *nullProcessor {
	err := errEmpty
	if len(errs) >= 1 {
		err = errs[1]
	}
	return &nullProcessor{
		err: err,
	}
}

func (n *nullProcessor) Process() (*parser.ParseResult, error) {
	return nil, n.err
}

func New(engine parser.Engine, model *config.Model, notifier notifier.Notifier) *processor {
	return &processor{
		engine:   engine,
		logger:   logger.Null,
		model:    model,
		notifier: notifier,
	}
}

func (p *processor) WithLogger(logger logger.Logger) *processor {
	p.logger = logger
	return p
}

func (p *processor) Process() (*parser.ParseResult, error) {
	result, err := p.engine.Get(p.model, nil, nil)
	if p.notifier != nil {
		isArray := false
		if p.model.ArrayConfig != nil {
			isArray = true
		}

		errNot := p.notifier.Inform(result, err, isArray)
		if errNot != nil {
			p.logger.Errorw("cannot notify about result", "error", errNot.Error())
		}
	}

	if err != nil {
		p.logger.Errorw("parser return error processing data", "error", err.Error())
		return nil, err
	}
	return result, nil
}

func CreateProcessor(item *config.Item, references map[string]*config.ModelField, logger logger.Logger) Processor {
	if item.Name == "" {
		return Null(errMissingName)
	}

	parser.SetReference(references, logger)

	var notifierInstance notifier.Notifier

	if item.NotifierConfig != nil {
		if item.NotifierConfig.TelegramBot != nil {
			tgBot, errBot := notifier.NewTelegramBot(item.Name, item.NotifierConfig.TelegramBot)
			if errBot != nil {
				logger.Infow("cant setup telegram bot notifier", "error", errBot.Error())
			} else {
				notifierInstance = tgBot.WithLogger(logger.With("notifier", "telegram_bot"))
			}
		}
		if item.NotifierConfig.Console != nil {
			notifierInstance = notifier.NewConsole(item.Name, item.NotifierConfig.Console).WithLogger(logger.With("notifier", "console"))
		}

		if notifierInstance != nil {
			notifierInstance.SetConfig(item.NotifierConfig)
		}

	}

	logger = logger.With("name", item.Name)

	return New(parser.NewEngine(item.ConnectorConfig, logger.With("component", "processor_engine")), item.Model, notifierInstance).WithLogger(logger)
}

package processor

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/notifier"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/references"
	"github.com/PxyUp/fitter/pkg/utils"
)

var (
	errEmpty       = errors.New("empty response")
	errMissingName = errors.New("missing name in configuration of the fitter")
)

type Processor interface {
	Process(input builder.Interfacable) (*parser.ParseResult, error)
}

type processor struct {
	logger      logger.Logger
	model       *config.Model
	notifier    notifier.Notifier
	notifierCfg *config.NotifierConfig
	engine      parser.Engine
	name        string
}

type nullProcessor struct {
	err error
}

func Null(errs ...error) *nullProcessor {
	err := errEmpty
	if len(errs) >= 1 {
		err = errs[0]
	}
	return &nullProcessor{
		err: err,
	}
}

func (n *nullProcessor) Process(input builder.Interfacable) (*parser.ParseResult, error) {
	return nil, n.err
}

func New(name string, engine parser.Engine, model *config.Model, notifier notifier.Notifier, notifierCfg *config.NotifierConfig) *processor {
	return &processor{
		name:        name,
		engine:      engine,
		logger:      logger.Null,
		notifierCfg: notifierCfg,
		model:       model,
		notifier:    notifier,
	}
}

func (p *processor) WithLogger(logger logger.Logger) *processor {
	p.logger = logger
	return p
}

func (p *processor) Process(input builder.Interfacable) (*parser.ParseResult, error) {
	result, err := p.engine.Get(p.model, nil, nil, input)
	if p.notifier != nil {
		isArray := false
		if p.model.ArrayConfig != nil || p.model.IsArray {
			isArray = true
		}

		if p.notifierCfg != nil {
			if p.notifierCfg.Template != "" {
				strValue := builder.ToJsonableFromString(utils.Format(p.notifierCfg.Template, result, nil, nil))
				result = &parser.ParseResult{
					RawResult: strValue.Raw(),
					Json:      strValue.ToJson(),
				}
			}
			need, errShInform := notifier.ShouldInform(p.notifierCfg, result)
			if errShInform != nil {
				p.logger.Errorw("cannot calculate notification setting", "error", errShInform.Error())
				return nil, errShInform
			}
			if !need {
				p.logger.Debug("skip notification because not match expression")
				return result, nil
			}
		}
		errNot := notifier.Inform(p.notifier, p.name, result, err, isArray && p.notifierCfg.SendArrayByItem && !result.IsEmpty(), p.notifier.GetLogger(), input)
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

func CreateProcessor(item *config.Item, refMap config.RefMap, logger logger.Logger) Processor {
	if item.Name == "" {
		return Null(errMissingName, nil)
	}

	references.SetReference(refMap, func(refName string, model *config.ModelField) (builder.Jsonable, error) {
		return parser.NewEngine(model.ConnectorConfig, logger.With("reference_name", refName)).Get(model.Model, nil, nil, nil)
	})

	var notifierInstance notifier.Notifier

	if item.NotifierConfig != nil {
		if item.NotifierConfig.TelegramBot != nil {
			notifierInstance = notifier.NewTelegramBot(item.Name, item.NotifierConfig.TelegramBot).WithLogger(logger.With("notifier", "telegram_bot"))
		}

		if item.NotifierConfig.Console != nil {
			notifierInstance = notifier.NewConsole(item.Name, item.NotifierConfig.Console).WithLogger(logger.With("notifier", "console"))
		}

		if item.NotifierConfig.Http != nil {
			notifierInstance = notifier.NewHttpNotifier(item.Name, item.NotifierConfig.Http).WithLogger(logger.With("notifier", "http"))
		}

		if item.NotifierConfig.Redis != nil {
			notifierInstance = notifier.NewRedis(item.Name, item.NotifierConfig.Redis).WithLogger(logger.With("notifier", "redis"))
		}

		if item.NotifierConfig.File != nil {
			notifierInstance = notifier.NewFile(item.Name, item.NotifierConfig.File).WithLogger(logger.With("notifier", "file"))
		}
	}

	logger = logger.With("name", item.Name)

	return New(item.Name, parser.NewEngine(item.ConnectorConfig, logger.With("component", "processor_engine")), item.Model, notifierInstance, item.NotifierConfig).WithLogger(logger)
}

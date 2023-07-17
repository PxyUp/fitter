package processor

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/notifier"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/plugins/store"
)

var (
	errEmpty       = errors.New("empty response")
	errMissingName = errors.New("missing name in configuration of the fitter")
)

type Processor interface {
	Process() (*parser.ParseResult, error)
}

type processor struct {
	connector     connectors.Connector
	parserFactory parser.Factory
	logger        logger.Logger
	model         *config.Model
	notifier      notifier.Notifier
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

func New(connector connectors.Connector, parserFactory parser.Factory, model *config.Model, notifier notifier.Notifier) *processor {
	return &processor{
		connector:     connector,
		parserFactory: parserFactory,
		logger:        logger.Null,
		model:         model,
		notifier:      notifier,
	}
}

func (p *processor) WithLogger(logger logger.Logger) *processor {
	p.logger = logger
	return p
}

func (p *processor) Process() (*parser.ParseResult, error) {
	body, err := p.connector.Get(nil, nil)
	if err != nil {
		p.logger.Errorw("connector return error during fetch data", "error", err.Error())
		return nil, err
	}

	result, err := p.parserFactory(body, p.logger).Parse(p.model)
	if p.notifier != nil {
		isArray := false
		if p.model.ArrayConfig != nil {
			isArray = true
		}

		errNot := p.notifier.Inform(result, err, isArray)
		if errNot != nil {
			p.logger.Errorw("cannot notify about result", "error", errNot.Error(), "body", string(body))
		}
	}

	if err != nil {
		p.logger.Errorw("parser return error processing data", "error", err.Error(), "body", string(body))
		return nil, err
	}
	return result, nil
}

func CreateProcessor(item *config.Item, logger logger.Logger) Processor {
	if item.Name == "" {
		return Null(errMissingName)
	}

	if item.ConnectorConfig == nil {
		return Null()
	}

	var connector connectors.Connector
	if item.ConnectorConfig.StaticConfig != nil {
		connector = connectors.NewStatic(item.ConnectorConfig.StaticConfig).WithLogger(logger.With("connector", "static"))
	}
	if item.ConnectorConfig.ServerConfig != nil {
		connector = connectors.NewAPI(item.ConnectorConfig.Url, item.ConnectorConfig.ServerConfig, nil).WithLogger(logger.With("connector", "server"))
	}
	if item.ConnectorConfig.BrowserConfig != nil {
		connector = connectors.NewBrowser(item.ConnectorConfig.Url, item.ConnectorConfig.BrowserConfig).WithLogger(logger.With("connector", "browser"))
	}
	if item.ConnectorConfig.PluginConnectorConfig != nil {
		connector = store.Store.GetConnectorPlugin(item.ConnectorConfig.PluginConnectorConfig.Name, item.ConnectorConfig.PluginConnectorConfig, logger.With("connector", item.ConnectorConfig.PluginConnectorConfig.Name))
	}

	var parserFactory parser.Factory
	if item.ConnectorConfig.ResponseType == config.Json {
		parserFactory = parser.JsonFactory
	}
	if item.ConnectorConfig.ResponseType == config.HTML {
		parserFactory = parser.HTMLFactory
	}
	if item.ConnectorConfig.ResponseType == config.XPath {
		parserFactory = parser.XPathFactory
	}

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

	if connector == nil || parserFactory == nil {
		return Null()
	}

	connector = connectors.WithAttempts(connector, item.ConnectorConfig.Attempts)

	logger = logger.With("name", item.Name)

	return New(connector, parserFactory, item.Model, notifierInstance).WithLogger(logger)
}

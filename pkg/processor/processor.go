package processor

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/notifier"
	"github.com/PxyUp/fitter/pkg/parser"
)

var (
	null = &nullProcessor{}

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
	body, err := p.connector.Get()
	if err != nil {
		p.logger.Errorw("connector return error during fetch data", "error", err.Error())
		return nil, err
	}

	result, err := p.parserFactory(body).Parse(p.model)

	p.notifier.Inform(result, err)
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
	if item.ConnectorConfig.ConnectorType == config.Server && item.ConnectorConfig.ServerConfig != nil {
		connector = connectors.NewAPI(item.ConnectorConfig.ServerConfig, nil).WithLogger(logger)
	}
	if item.ConnectorConfig.ConnectorType == config.Browser && item.ConnectorConfig.BrowserConfig != nil {
		connector = connectors.NewBrowser(item.ConnectorConfig.BrowserConfig).WithLogger(logger)
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

	if connector == nil || parserFactory == nil {
		return Null()
	}

	logger = logger.With("name", item.Name)

	return New(connector, parserFactory, item.Model, notifier.New(item.Name, item.NotifierConfig).WithLogger(logger)).WithLogger(logger)
}

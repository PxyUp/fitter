package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
)

var (
	JsonFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return newJson(bytes).WithLogger(logger.With("parser", "json"))
	}

	HTMLFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return newHTML(bytes).WithLogger(logger.With("parser", "html"))
	}

	XPathFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return newXPath(bytes).WithLogger(logger.With("parser", "xpath"))
	}
)

type Factory func([]byte, logger.Logger) Parser

type Parser interface {
	Parse(model *config.Model) (*ParseResult, error)
}

type ParseResult struct {
	Raw string `json:"raw"`
}

func (p *ParseResult) ToJson() string {
	return p.Raw
}

func buildGeneratedField(parsedValue builder.Jsonable, field *config.GeneratedFieldConfig, logger logger.Logger) builder.Jsonable {
	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	if field.Formatted != nil {
		return builder.String(format(field.Formatted.Template, parsedValue))
	}

	if field.Model != nil {
		if field.Model.ConnectorConfig == nil || field.Model.Model == nil {
			return builder.Null()
		}

		var connector connectors.Connector

		if field.Model.ConnectorConfig.ConnectorType == config.Server && field.Model.ConnectorConfig.ServerConfig != nil {
			connector = connectors.NewAPI(&config.ServerConnectorConfig{
				Method:  field.Model.ConnectorConfig.ServerConfig.Method,
				Headers: field.Model.ConnectorConfig.ServerConfig.Headers,
				Url:     format(field.Model.ConnectorConfig.ServerConfig.Url, parsedValue),
			}, nil).WithLogger(logger.With("connector", "server"))
		}

		if field.Model.ConnectorConfig.ConnectorType == config.Browser && field.Model.ConnectorConfig.BrowserConfig != nil {
			connector = connectors.NewBrowser(&config.BrowserConnectorConfig{
				Url:      format(field.Model.ConnectorConfig.BrowserConfig.Url, parsedValue),
				Chromium: field.Model.ConnectorConfig.BrowserConfig.Chromium,
			}).WithLogger(logger.With("connector", "browser"))
		}

		var parserFactory Factory
		if field.Model.ConnectorConfig.ResponseType == config.Json {
			parserFactory = JsonFactory
		}
		if field.Model.ConnectorConfig.ResponseType == config.HTML {
			parserFactory = HTMLFactory
		}
		if field.Model.ConnectorConfig.ResponseType == config.XPath {
			parserFactory = XPathFactory
		}

		if connector == nil || parserFactory == nil {
			return builder.Null()
		}

		connector = connectors.WithAttempts(connector, field.Model.ConnectorConfig.Attempts)

		body, err := connector.Get()
		if err != nil {
			return builder.Null()
		}

		result, err := parserFactory(body, logger).Parse(field.Model.Model)
		if err != nil {
			return builder.Null()
		}

		return result
	}

	return builder.Null()
}

func fillUpBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
	if !source.Exists() {
		return builder.Null()
	}
	switch field.Type {
	case config.Null:
		return builder.Null()
	case config.String:
		return builder.String(source.String())
	case config.Bool:
		if !source.IsBool() {
			return builder.Null()
		}
		return builder.Bool(source.Bool())
	case config.Float:
		return builder.Float(float32(source.Float()))
	case config.Int:
		return builder.Int(int(source.Int()))
	}

	return builder.EMPTY
}

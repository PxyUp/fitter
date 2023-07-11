package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/PxyUp/fitter/pkg/plugins/store"
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

var (
	_ builder.Jsonable = &ParseResult{}
)

type ParseResult struct {
	Json      string `json:"raw"`
	RawResult interface{}
}

func (p *ParseResult) Raw() interface{} {
	return p.RawResult
}

func (p *ParseResult) IsEmpty() bool {
	return len(p.Json) == 0
}

func (p *ParseResult) ToJson() string {
	return p.Json
}

func buildGeneratedField(parsedValue builder.Jsonable, field *config.GeneratedFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	if field.Formatted != nil {
		return builder.String(format(field.Formatted.Template, parsedValue, index))
	}

	if field.Plugin != nil {
		return store.Store.Get(field.Plugin.Name, logger).Format(parsedValue, field.Plugin, logger.With("plugin", field.Plugin.Name), index)
	}

	if field.Model != nil {
		if field.Model.ConnectorConfig == nil || field.Model.Model == nil {
			return builder.Null()
		}

		var connector connectors.Connector

		if field.Model.ConnectorConfig.StaticConfig != nil {
			staticValue := format(field.Model.ConnectorConfig.StaticConfig.Value, parsedValue, index)
			connector = connectors.NewStatic(&config.StaticConnectorConfig{Value: staticValue}).WithLogger(logger.With("connector", "static"))
		}

		if field.Model.ConnectorConfig.ServerConfig != nil {
			connector = connectors.NewAPI(format(field.Model.ConnectorConfig.Url, parsedValue, index), &config.ServerConnectorConfig{
				Method:  field.Model.ConnectorConfig.ServerConfig.Method,
				Headers: field.Model.ConnectorConfig.ServerConfig.Headers,
				Timeout: field.Model.ConnectorConfig.ServerConfig.Timeout,
				Body:    format(field.Model.ConnectorConfig.ServerConfig.Body, parsedValue, index),
			}, nil).WithLogger(logger.With("connector", "server"))
		}

		if field.Model.ConnectorConfig.BrowserConfig != nil {
			connector = connectors.NewBrowser(format(field.Model.ConnectorConfig.Url, parsedValue, index), &config.BrowserConnectorConfig{
				Chromium:   field.Model.ConnectorConfig.BrowserConfig.Chromium,
				Docker:     field.Model.ConnectorConfig.BrowserConfig.Docker,
				Playwright: field.Model.ConnectorConfig.BrowserConfig.Playwright,
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
	case config.Float64:
		return builder.Float64(source.Float())
	case config.Int:
		return builder.Int(int(source.Int()))
	case config.Int64:
		return builder.Int64(source.Int())
	case config.Array:
		return builder.PureString(source.String())
	case config.Object:
		return builder.PureString(source.String())
	}

	return builder.EMPTY
}

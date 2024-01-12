package parser

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/utils"
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

func buildGeneratedField(parsedValue builder.Jsonable, fieldType config.FieldType, field *config.GeneratedFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	if fieldType == config.String {
		parsedValue = builder.PureString(parsedValue.ToJson())
	}

	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.File != nil {
		filePath, err := ProcessFileField(parsedValue, index, field.File, logger)
		if err != nil {
			logger.Errorw("error during process file field", "error", err.Error())
			return builder.Null()
		}
		return builder.String(filePath)
	}

	if field.Calculated != nil && field.Calculated.Expression != "" {
		res, err := ProcessExpression(field.Calculated.Expression, parsedValue, index)
		if err != nil {
			logger.Errorw("error during process calculated field", "error", err.Error())
			return builder.Null()
		}

		return builder.Static(&config.StaticGeneratedFieldConfig{
			Type:  field.Calculated.Type,
			Value: fmt.Sprintf("%v", res),
		})
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	if field.Formatted != nil {
		return builder.String(utils.Format(field.Formatted.Template, parsedValue, index))
	}

	if field.Plugin != nil {
		return store.Store.GetFieldPlugin(field.Plugin.Name, logger).Format(parsedValue, field.Plugin, logger.With("plugin", field.Plugin.Name), index)
	}

	if field.Model != nil {
		if field.Model.ConnectorConfig == nil || field.Model.Model == nil {
			return builder.Null()
		}

		var connector connectors.Connector

		if field.Model.ConnectorConfig.StaticConfig != nil {
			connector = connectors.NewStatic(field.Model.ConnectorConfig.StaticConfig).WithLogger(logger.With("connector", "static"))
		}

		if field.Model.ConnectorConfig.ServerConfig != nil {
			connector = connectors.NewAPI(field.Model.ConnectorConfig.Url, field.Model.ConnectorConfig.ServerConfig, nil).WithLogger(logger.With("connector", "server"))
		}

		if field.Model.ConnectorConfig.BrowserConfig != nil {
			connector = connectors.NewBrowser(field.Model.ConnectorConfig.Url, field.Model.ConnectorConfig.BrowserConfig).WithLogger(logger.With("connector", "browser"))
		}

		if field.Model.ConnectorConfig.PluginConnectorConfig != nil {
			connector = store.Store.GetConnectorPlugin(field.Model.ConnectorConfig.PluginConnectorConfig.Name, field.Model.ConnectorConfig.PluginConnectorConfig, logger.With("connector", field.Model.ConnectorConfig.PluginConnectorConfig.Name))
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

		body, err := connector.Get(parsedValue, index)
		if err != nil {
			return builder.Null()
		}

		logger.Debugw("connector answer", "content", string(body))

		result, err := parserFactory(body, logger).Parse(field.Model.Model)
		if err != nil {
			return builder.Null()
		}

		if field.Model.Type == config.Array || field.Model.Type == config.Object {
			if field.Model.Path != "" {
				return builder.PureString(gjson.Parse(result.ToJson()).Get(field.Model.Path).Raw)
			}
			return result
		}
		if field.Model.Path != "" {
			return fillUpBaseField(gjson.Parse(result.ToJson()).Get(field.Model.Path), &config.BaseField{
				Type: field.Model.Type,
			})
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

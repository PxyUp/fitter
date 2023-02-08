package parser

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
)

var (
	JsonFactory Factory = func(bytes []byte) Parser {
		return NewJson(bytes)
	}

	HTMLFactory Factory = func(bytes []byte) Parser {
		return NewHTML(bytes)
	}
)

type Factory func([]byte) Parser

type Parser interface {
	Parse(model *config.Model) (*ParseResult, error)
}

type ParseResult struct {
	Raw string `json:"raw"`
}

func (p *ParseResult) ToJson() string {
	return p.Raw
}

func buildGeneratedField(parsedValue builder.Jsonable, field *config.GeneratedFieldConfig) builder.Jsonable {
	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	if field.Formatted != nil {
		return builder.String(fmt.Sprintf(field.Formatted.Template, parsedValue.ToJson()))
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
				Url:     fmt.Sprintf(field.Model.ConnectorConfig.ServerConfig.Url, parsedValue.ToJson()),
			})
		}

		var parserFactory Factory
		if field.Model.ConnectorConfig.ResponseType == config.Json {
			parserFactory = JsonFactory
		}
		if field.Model.ConnectorConfig.ResponseType == config.HTML {
			parserFactory = HTMLFactory
		}

		if connector == nil || parserFactory == nil {
			return builder.Null()
		}

		body, err := connector.Get()
		if err != nil {
			return builder.Null()
		}

		result, err := parserFactory(body).Parse(field.Model.Model)
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

	return builder.Null()
}

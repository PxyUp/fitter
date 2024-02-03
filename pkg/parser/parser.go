package parser

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/tidwall/gjson"
)

var (
	JsonFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return NewJson(bytes, logger.With("parser", "json"))
	}

	HTMLFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return NewHTML(bytes, logger.With("parser", "html"))
	}

	XPathFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return NewXPath(bytes, logger.With("parser", "xpath"))
	}

	XMLFactory Factory = func(bytes []byte, logger logger.Logger) Parser {
		return NewXML(bytes, logger.With("parser", "xml"))
	}
)

type Factory func([]byte, logger.Logger) Parser

type Parser interface {
	Parse(model *config.Model, input builder.Interfacable) (*ParseResult, error)
}

var (
	_ builder.Interfacable = &ParseResult{}
)

type ParseResult struct {
	Json      string `json:"raw"`
	RawResult json.RawMessage
}

func (p *ParseResult) ToInterface() interface{} {
	var t interface{}
	err := json.Unmarshal(p.RawResult, &t)
	if err != nil {
		return nil
	}
	return t
}

func (p *ParseResult) Raw() json.RawMessage {
	return p.RawResult
}

func (p *ParseResult) IsEmpty() bool {
	return len(p.Json) == 0
}

func (p *ParseResult) ToJson() string {
	return p.Json
}

func getExpressionResult(expr string, fieldType config.FieldType, value builder.Interfacable, index *uint32, input builder.Interfacable, logger logger.Logger) builder.Interfacable {
	res, err := utils.ProcessExpression(expr, value, index, input)
	if err != nil {
		logger.Errorw("error during process calculated field", "error", err.Error())
		return builder.NullValue
	}

	return builder.Static(&config.StaticGeneratedFieldConfig{
		Type:  fieldType,
		Value: fmt.Sprintf("%v", res),
	})
}

func buildGeneratedField(parsedValue builder.Interfacable, fieldType config.FieldType, field *config.GeneratedFieldConfig, logger logger.Logger, index *uint32, input builder.Interfacable) builder.Interfacable {
	if fieldType == config.String {
		parsedValue = builder.PureString(parsedValue.ToJson())
	}

	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.File != nil {
		filePath, err := ProcessFileField(parsedValue, index, input, field.File, logger)
		if err != nil {
			logger.Errorw("error during process file field", "error", err.Error())
			return builder.NullValue
		}
		return builder.String(filePath)
	}

	if field.Calculated != nil && field.Calculated.Expression != "" {
		return getExpressionResult(field.Calculated.Expression, field.Calculated.Type, parsedValue, index, input, logger)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	if field.Formatted != nil {
		return builder.String(utils.Format(field.Formatted.Template, parsedValue, index, input), false)
	}

	if field.Plugin != nil {
		return store.Store.GetFieldPlugin(field.Plugin.Name, logger).Format(parsedValue, field.Plugin, logger.With("plugin", field.Plugin.Name), index, input)
	}

	if field.Model != nil {
		if field.Model.Model == nil {
			return builder.NullValue
		}
		result, err := NewEngine(field.Model.ConnectorConfig, logger.With("component", "engine")).Get(field.Model.Model, parsedValue, index, input)
		if err != nil {
			return builder.NullValue
		}

		if field.Model.Expression != "" {
			return getExpressionResult(field.Model.Expression, field.Model.Type, result, index, input, logger)
		}

		if field.Model.Type == config.Array || field.Model.Type == config.Object {
			if field.Model.Path != "" {
				return builder.ToJsonable([]byte(gjson.Parse(result.ToJson()).Get(field.Model.Path).Raw))
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

	return builder.NullValue
}

func fillUpBaseField(source gjson.Result, field *config.BaseField) builder.Interfacable {
	if !source.Exists() {
		return builder.NullValue
	}
	switch field.Type {
	case config.Null:
		return builder.NullValue
	case config.RawString:
		return builder.String(source.String(), false)
	case config.String:
		return builder.String(source.String())
	case config.Bool:
		if !source.IsBool() {
			return builder.NullValue
		}
		return builder.Bool(source.Bool())
	case config.Float, config.Float64, config.Int, config.Int64:
		return builder.Number(source.Float())
	case config.Array:
		return builder.ToJsonable([]byte(source.String()))
	case config.Object:
		return builder.ToJsonable([]byte(source.String()))
	}

	return builder.EMPTY
}

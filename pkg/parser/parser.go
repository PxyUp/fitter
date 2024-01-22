package parser

import (
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
			return builder.NullValue
		}
		return builder.String(filePath)
	}

	if field.Calculated != nil && field.Calculated.Expression != "" {
		res, err := ProcessExpression(field.Calculated.Expression, parsedValue, index)
		if err != nil {
			logger.Errorw("error during process calculated field", "error", err.Error())
			return builder.NullValue
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
		return builder.String(utils.Format(field.Formatted.Template, parsedValue, index), false)
	}

	if field.Plugin != nil {
		return store.Store.GetFieldPlugin(field.Plugin.Name, logger).Format(parsedValue, field.Plugin, logger.With("plugin", field.Plugin.Name), index)
	}

	if field.Model != nil {
		if field.Model.Model == nil {
			return builder.NullValue
		}
		result, err := NewEngine(field.Model.ConnectorConfig, logger.With("component", "engine")).Get(field.Model.Model, parsedValue, index)
		if err != nil {
			return builder.NullValue
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

	return builder.NullValue
}

func fillUpBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
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

package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
)

type jsonParser struct {
	logger     logger.Logger
	body       []byte
	parserBody gjson.Result
}

func NewJson(body []byte) *jsonParser {
	return &jsonParser{
		body:       body,
		logger:     logger.Null,
		parserBody: gjson.ParseBytes(body),
	}
}

func (j *jsonParser) WithLogger(logger logger.Logger) {
	j.logger = logger
}

func (j *jsonParser) Parse(model *config.Model) (*ParseResult, error) {
	if model.Type == config.ArrayModel {
		return &ParseResult{
			Raw: j.buildArray(model.ArrayConfig).ToJson(),
		}, nil
	}
	return &ParseResult{
		Raw: j.buildObject(model.ObjectConfig).ToJson(),
	}, nil
}

func (j *jsonParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	if array.RootPath == "" {
		return buildArrayField(j.parserBody, array)
	}
	return buildArrayField(j.parserBody.Get(array.RootPath), array)
}

func (j *jsonParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return buildObjectField(j.parserBody, object.Fields)
}

func buildObjectField(parent gjson.Result, fields map[string]*config.Field) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	for k, v := range fields {
		if v.BaseField != nil {
			kv[k] = buildBaseField(parent.Get(v.BaseField.Path), v.BaseField)
			continue
		}

		if v.ObjectConfig != nil {
			kv[k] = buildObjectField(parent, v.ObjectConfig.Fields)
			continue
		}

		if v.ArrayConfig != nil {
			kv[k] = buildArrayField(parent.Get(v.ArrayConfig.RootPath), v.ArrayConfig)
		}
	}

	return builder.Object(kv)
}

func buildArrayField(parent gjson.Result, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, len(parent.Array()))
	if array.ItemConfig.Field != nil {
		for index, res := range parent.Array() {
			if array.ItemConfig.Field.Path == "" {
				values[index] = buildBaseField(res, array.ItemConfig.Field)
				continue
			}
			values[index] = buildBaseField(res.Get(array.ItemConfig.Field.Path), array.ItemConfig.Field)
		}
		return builder.Array(values)
	}

	for index, res := range parent.Array() {
		values[index] = buildObjectField(res, array.ItemConfig.Fields)
	}

	return builder.Array(values)
}

func buildGeneratedField(field *config.GeneratedFieldConfig) builder.Jsonable {
	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	return builder.Null()
}

func buildBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
	if field.Generated != nil {
		return buildGeneratedField(field.Generated)
	}

	switch field.Type {
	case config.Null:
		return builder.Null()
	case config.String:
		if !source.Exists() {
			return builder.Null()
		}
		return builder.String(source.String())
	case config.Bool:
		if !source.Exists() || !source.IsBool() {
			return builder.Null()
		}
		return builder.Bool(source.Bool())
	case config.Float:
		if !source.Exists() {
			return builder.Null()
		}
		return builder.Float(float32(source.Float()))
	case config.Int:
		if !source.Exists() {
			return builder.Null()
		}
		return builder.Int(int(source.Int()))
	}

	return builder.Null()
}

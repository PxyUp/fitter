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
		return j.buildArrayField(j.parserBody, array)
	}
	return j.buildArrayField(j.parserBody.Get(array.RootPath), array)
}

func (j *jsonParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return j.buildObjectField(j.parserBody, object.Fields)
}

func (j *jsonParser) buildObjectField(parent gjson.Result, fields map[string]*config.Field) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	for k, v := range fields {
		if v.BaseField != nil {
			kv[k] = j.buildBaseField(parent.Get(v.BaseField.Path), v.BaseField)
			continue
		}

		if v.ObjectConfig != nil {
			kv[k] = j.buildObjectField(parent, v.ObjectConfig.Fields)
			continue
		}

		if v.ArrayConfig != nil {
			kv[k] = j.buildArrayField(parent.Get(v.ArrayConfig.RootPath), v.ArrayConfig)
		}
	}

	return builder.Object(kv)
}

func (j *jsonParser) buildArrayField(parent gjson.Result, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, len(parent.Array()))
	if array.ItemConfig.Field != nil {
		for index, res := range parent.Array() {
			if array.ItemConfig.Field.Path == "" {
				values[index] = j.buildBaseField(res, array.ItemConfig.Field)
				continue
			}
			values[index] = j.buildBaseField(res.Get(array.ItemConfig.Field.Path), array.ItemConfig.Field)
		}
		return builder.Array(values)
	}

	for index, res := range parent.Array() {
		values[index] = j.buildObjectField(res, array.ItemConfig.Fields)
	}

	return builder.Array(values)
}

func (j *jsonParser) buildBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
	if field.Generated != nil {
		return buildGeneratedField(field.Generated)
	}
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

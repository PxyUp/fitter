package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
	"sync"
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

	var wg sync.WaitGroup
	for lKey, lValue := range fields {
		key := lKey
		value := lValue

		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			if v.BaseField != nil {
				if v.BaseField.Path == "" {
					kv[k] = j.buildBaseField(parent, v.BaseField)
					return
				}
				kv[k] = j.buildBaseField(parent.Get(v.BaseField.Path), v.BaseField)
				return
			}

			if v.ObjectConfig != nil {
				kv[k] = j.buildObjectField(parent, v.ObjectConfig.Fields)
				return
			}

			if v.ArrayConfig != nil {
				if v.ArrayConfig.RootPath == "" {
					kv[k] = j.buildArrayField(parent, v.ArrayConfig)
					return
				}
				kv[k] = j.buildArrayField(parent.Get(v.ArrayConfig.RootPath), v.ArrayConfig)
			}
		}(key, value)
	}

	wg.Wait()

	return builder.Object(kv)
}

func (j *jsonParser) buildArrayField(parent gjson.Result, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, len(parent.Array()))

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		for lIndex, lRes := range parent.Array() {
			i := lIndex
			r := lRes

			wg.Add(1)
			go func(index int, res gjson.Result) {
				defer wg.Done()

				if array.ItemConfig.Field.Path == "" {
					values[index] = j.buildBaseField(res, array.ItemConfig.Field)
					return
				}
				values[index] = j.buildBaseField(res.Get(array.ItemConfig.Field.Path), array.ItemConfig.Field)
			}(i, r)

		}
		wg.Wait()
		return builder.Array(values)
	}

	if array.ItemConfig.ArrayConfig != nil {
		var wg sync.WaitGroup
		for lIndex, lRes := range parent.Array() {
			i := lIndex
			r := lRes

			wg.Add(1)
			go func(index int, res gjson.Result) {
				defer wg.Done()

				if array.ItemConfig.ArrayConfig.RootPath == "" {
					values[index] = j.buildArrayField(res, array.ItemConfig.ArrayConfig)
					return
				}
				values[index] = j.buildArrayField(res.Get(array.ItemConfig.ArrayConfig.RootPath), array.ItemConfig.ArrayConfig)
			}(i, r)
		}
		wg.Wait()
		return builder.Array(values)
	}

	var wg sync.WaitGroup
	for lIndex, lRes := range parent.Array() {
		i := lIndex
		r := lRes

		wg.Add(1)
		go func(index int, res gjson.Result) {
			defer wg.Done()
			values[index] = j.buildObjectField(res, array.ItemConfig.Fields)
		}(i, r)
	}

	wg.Wait()

	return builder.Array(values)
}

func (j *jsonParser) buildBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
	tempValue := fillUpBaseField(source, field)

	if field.Generated != nil {
		if field.Type == config.String {
			tempValue = builder.PureString(tempValue.ToJson())
		}
		generatedValue := buildGeneratedField(tempValue, field.Generated)
		if field.Generated.Model != nil {
			if field.Generated.Model.Type == config.Array || field.Generated.Model.Type == config.Object {
				if field.Generated.Model.Path != "" {
					return builder.PureString(gjson.Parse(generatedValue.ToJson()).Get(field.Generated.Model.Path).Raw)
				}
				return builder.PureString(generatedValue.ToJson())
			}
			if field.Generated.Model.Path != "" {
				return fillUpBaseField(gjson.Parse(generatedValue.ToJson()).Get(field.Generated.Model.Path), &config.BaseField{
					Type: config.FieldType(field.Generated.Model.Type),
				})
			}
		}

		return generatedValue
	}

	return tempValue
}

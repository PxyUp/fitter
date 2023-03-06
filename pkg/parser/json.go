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

func newJson(body []byte) *jsonParser {
	return &jsonParser{
		body:       body,
		logger:     logger.Null,
		parserBody: gjson.ParseBytes(body),
	}
}

func (j *jsonParser) WithLogger(logger logger.Logger) *jsonParser {
	j.logger = logger
	return j
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
	return j.buildArrayField(j.parserBody, array)
}

func (j *jsonParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return j.buildObjectField(j.parserBody, object.Fields)
}

func (j *jsonParser) buildStaticArray(cfg *config.StaticArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, len(cfg.Items))

	var wg sync.WaitGroup

	for lKey, lValue := range cfg.Items {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k uint32, v *config.Field) {
			defer wg.Done()

			values[k] = j.resolveField(j.parserBody, v)
		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (j *jsonParser) buildObjectField(parent gjson.Result, fields map[string]*config.Field) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	var mutex sync.Mutex

	var wg sync.WaitGroup
	for lKey, lValue := range fields {
		key := lKey
		value := lValue

		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			mutex.Lock()
			kv[k] = j.resolveField(parent, v)
			mutex.Unlock()
		}(key, value)
	}

	wg.Wait()

	return builder.Object(kv)
}

func (j *jsonParser) buildFirstOfField(source gjson.Result, fields []*config.Field) builder.Jsonable {
	for _, value := range fields {
		tempValue := j.resolveField(source, value)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (j *jsonParser) resolveField(parent gjson.Result, field *config.Field) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return j.buildFirstOfField(parent, field.FirstOf)
	}

	if field.BaseField != nil {
		return j.buildBaseField(parent, field.BaseField)
	}

	if field.ObjectConfig != nil {
		return j.buildObjectField(parent, field.ObjectConfig.Fields)
	}

	if field.ArrayConfig != nil {
		return j.buildArrayField(parent, field.ArrayConfig)
	}

	return builder.Null()
}

func (j *jsonParser) buildArrayField(parent gjson.Result, array *config.ArrayConfig) builder.Jsonable {
	if array.StaticConfig != nil {
		return j.buildStaticArray(array.StaticConfig)
	}

	if array.RootPath != "" {
		parent = parent.Get(array.RootPath)
	}

	values := make([]builder.Jsonable, len(parent.Array()))

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		for lIndex, lRes := range parent.Array() {
			i := lIndex
			r := lRes

			wg.Add(1)
			go func(index int, res gjson.Result) {
				defer wg.Done()

				values[index] = j.buildBaseField(res, array.ItemConfig.Field)
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

				values[index] = j.buildArrayField(res, array.ItemConfig.ArrayConfig)
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

func (j *jsonParser) buildFirstOfBaseField(source gjson.Result, fields []*config.BaseField) builder.Jsonable {
	for _, value := range fields {
		tempValue := j.buildBaseField(source, value)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (j *jsonParser) buildBaseField(source gjson.Result, field *config.BaseField) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return j.buildFirstOfBaseField(source, field.FirstOf)
	}

	if field.Path != "" {
		source = source.Get(field.Path)
	}

	tempValue := fillUpBaseField(source, field)

	if field.Generated != nil {
		if field.Type == config.String {
			tempValue = builder.PureString(tempValue.ToJson())
		}
		generatedValue := buildGeneratedField(tempValue, field.Generated, j.logger)
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

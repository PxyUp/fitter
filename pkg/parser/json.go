package parser

import (
	builder "github.com/PxyUp/fitter/pkg/builder"
	"sync"

	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/tidwall/gjson"
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
	if model.BaseField != nil {
		res := j.buildBaseField(j.parserBody, model.BaseField, nil)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		res := j.buildArray(model.ArrayConfig)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	res := j.buildObject(model.ObjectConfig)
	return &ParseResult{
		RawResult: res.Raw(),
		Json:      res.ToJson(),
	}, nil
}

func (j *jsonParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	return j.buildArrayField(j.parserBody, array)
}

func (j *jsonParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return j.buildObjectField(j.parserBody, object)
}

func (j *jsonParser) buildStaticArray(cfg *config.StaticArrayConfig) builder.Jsonable {
	length := len(cfg.Items)
	if cfg.Length > 0 {
		length = int(cfg.Length)
	}
	values := make([]builder.Jsonable, length)

	var wg sync.WaitGroup

	for lKey, lValue := range cfg.Items {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k uint32, v *config.Field) {
			defer wg.Done()

			arrIndex := k
			values[k] = j.resolveField(j.parserBody, v, &arrIndex)
		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (j *jsonParser) buildObjectField(parent gjson.Result, object *config.ObjectConfig) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	var mutex sync.Mutex

	var wg sync.WaitGroup
	for lKey, lValue := range object.Fields {
		key := lKey
		value := lValue

		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			mutex.Lock()
			kv[k] = j.resolveField(parent, v, nil)
			mutex.Unlock()
		}(key, value)
	}

	wg.Wait()

	return builder.Object(kv)
}

func (j *jsonParser) buildFirstOfField(source gjson.Result, fields []*config.Field, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := j.resolveField(source, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (j *jsonParser) resolveField(parent gjson.Result, field *config.Field, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return j.buildFirstOfField(parent, field.FirstOf, index)
	}

	if field.BaseField != nil {
		return j.buildBaseField(parent, field.BaseField, index)
	}

	if field.ObjectConfig != nil {
		return j.buildObjectField(parent, field.ObjectConfig)
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

	size := len(parent.Array())
	if array.LengthLimit > 0 {
		size = int(array.LengthLimit)
	}

	values := make([]builder.Jsonable, size)

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		for lIndex, lRes := range parent.Array() {
			if lIndex >= size {
				break
			}
			i := lIndex
			r := lRes

			wg.Add(1)
			go func(index int, res gjson.Result) {
				defer wg.Done()

				arrIndex := uint32(index)
				values[index] = j.buildBaseField(res, array.ItemConfig.Field, &arrIndex)
			}(i, r)

		}
		wg.Wait()
		return builder.Array(values)
	}

	if array.ItemConfig.ArrayConfig != nil {
		var wg sync.WaitGroup
		for lIndex, lRes := range parent.Array() {
			if lIndex >= size {
				break
			}
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
		if lIndex >= size {
			break
		}

		i := lIndex
		r := lRes

		wg.Add(1)
		go func(index int, res gjson.Result) {
			defer wg.Done()
			values[index] = j.buildObjectField(res, array.ItemConfig)
		}(i, r)
	}

	wg.Wait()

	return builder.Array(values)
}

func (j *jsonParser) buildFirstOfBaseField(source gjson.Result, fields []*config.BaseField, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := j.buildBaseField(source, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (j *jsonParser) buildBaseField(source gjson.Result, field *config.BaseField, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return j.buildFirstOfBaseField(source, field.FirstOf, index)
	}

	if field.Path != "" {
		source = source.Get(field.Path)
	}

	tempValue := fillUpBaseField(source, field)

	if field.Generated != nil {
		return buildGeneratedField(tempValue, field.Type, field.Generated, j.logger, index)
	}

	return tempValue
}

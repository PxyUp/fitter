package parser

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"strconv"
	"sync"
)

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

type engineParser[T comparable] struct {
	parserBody T
	getAll     func(T, string) []T
	getOne     func(T, string) T
	getText    func(T) string

	customFillUpBaseField func(T, *config.BaseField) builder.Jsonable
	logger                logger.Logger
}

func (e *engineParser[T]) fillUpBaseField(source T, field *config.BaseField) builder.Jsonable {
	if IsZero(source) {
		return builder.NullValue
	}

	text := e.getText(source)

	switch field.Type {
	case config.Null:
		return builder.NullValue
	case config.RawString:
		return builder.String(text, false)
	case config.String:
		return builder.String(text)
	case config.Bool:
		boolValue, err := strconv.ParseBool(text)
		if err != nil {
			return builder.NullValue
		}
		return builder.Bool(boolValue)
	case config.Float:
		float32Value, err := strconv.ParseFloat(text, 32)
		if err != nil {
			return builder.NullValue
		}
		return builder.Float(float32(float32Value))
	case config.Float64:
		float64Value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return builder.NullValue
		}
		return builder.Float64(float64Value)
	case config.Int:
		intValue, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return builder.NullValue
		}
		return builder.Int(int(intValue))
	case config.Int64:
		int64Value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return builder.NullValue
		}
		return builder.Int64(int64Value)
	case config.Array:
		return builder.PureString(text)
	case config.Object:
		return builder.PureString(text)
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildObjectField(source T, objectConfig *config.ObjectConfig) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for lKey, lValue := range objectConfig.Fields {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			mutex.Lock()
			kv[k] = e.resolveField(source, v, nil)
			mutex.Unlock()

		}(key, value)

	}

	wg.Wait()

	return builder.Object(kv)
}

func (e *engineParser[T]) buildFirstOfBaseField(source T, fields []*config.BaseField, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := e.buildBaseField(source, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildFirstOfField(parent T, fields []*config.Field, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := e.resolveField(parent, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildBaseField(source T, field *config.BaseField, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return e.buildFirstOfBaseField(source, field.FirstOf, index)
	}

	if field.Path != "" {
		source = e.getOne(source, field.Path)
	}

	var tempValue builder.Jsonable
	if e.customFillUpBaseField != nil {
		tempValue = e.customFillUpBaseField(source, field)
	} else {
		tempValue = e.fillUpBaseField(source, field)
	}

	if field.Generated != nil {
		return buildGeneratedField(tempValue, field.Type, field.Generated, e.logger, index)
	}

	return tempValue
}

func (e *engineParser[T]) resolveField(parent T, field *config.Field, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return e.buildFirstOfField(parent, field.FirstOf, index)
	}

	if field.BaseField != nil {
		return e.buildBaseField(parent, field.BaseField, index)
	}

	if field.ObjectConfig != nil {
		return e.buildObjectField(parent, field.ObjectConfig)
	}

	if field.ArrayConfig != nil {
		return e.buildArrayField(e.getAll(parent, field.ArrayConfig.RootPath), field.ArrayConfig)
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildStaticArray(cfg *config.StaticArrayConfig) builder.Jsonable {
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
			values[k] = e.resolveField(e.parserBody, v, &arrIndex)

		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (e *engineParser[T]) buildArray(array *config.ArrayConfig) builder.Jsonable {
	return e.buildArrayField(e.getAll(e.parserBody, array.RootPath), array)
}

func (e *engineParser[T]) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return e.buildObjectField(e.parserBody, object)
}

func (e *engineParser[T]) Parse(model *config.Model) (*ParseResult, error) {
	if IsZero(e.parserBody) {
		return &ParseResult{
			RawResult: builder.NullValue.Raw(),
			Json:      builder.NullValue.ToJson(),
		}, nil
	}

	if model.BaseField != nil {
		res := e.buildBaseField(e.parserBody, model.BaseField, nil)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		res := e.buildArray(model.ArrayConfig)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	res := e.buildObject(model.ObjectConfig)
	return &ParseResult{
		RawResult: res.Raw(),
		Json:      res.ToJson(),
	}, nil
}

func (e *engineParser[T]) buildArrayField(parent []T, cfg *config.ArrayConfig) builder.Jsonable {
	if cfg.StaticConfig != nil {
		return e.buildStaticArray(cfg.StaticConfig)
	}

	size := len(parent)
	if cfg.LengthLimit > 0 {
		size = int(cfg.LengthLimit)
	}

	if cfg.ItemConfig.Field != nil {
		return FillArrayBaseField(e, parent, size, cfg.ItemConfig.Field)
	}

	if cfg.ItemConfig.ArrayConfig != nil {
		return FillArrayArrayField(e, parent, size, e.getAll, cfg.ItemConfig.ArrayConfig)
	}

	return FillArrayObjectField(e, parent, size, cfg.ItemConfig)
}

func FillArrayBaseField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.BaseField) builder.Jsonable {
	values := make([]builder.Jsonable, size)

	var wg sync.WaitGroup
	for iL, sL := range parent {
		if iL >= size {
			break
		}

		i := iL
		s := sL
		wg.Add(1)
		go func(index int, selection T) {
			defer wg.Done()

			arrIndex := uint32(index)

			values[index] = engine.buildBaseField(selection, cfg, &arrIndex)
		}(i, s)

	}
	wg.Wait()

	return builder.Array(values)
}

func FillArrayArrayField[T comparable](engine *engineParser[T], parent []T, size int, fn func(T, string) []T, cfg *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, size)

	var wg sync.WaitGroup
	for iL, iS := range parent {
		if iL >= size {
			break
		}

		i := iL
		s := iS
		wg.Add(1)
		go func(index int, selection T) {
			defer wg.Done()

			values[index] = engine.buildArrayField(fn(selection, cfg.RootPath), cfg)
		}(i, s)
	}
	wg.Wait()

	return builder.Array(values)
}

func FillArrayObjectField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.ObjectConfig) builder.Jsonable {
	values := make([]builder.Jsonable, size)

	var wg sync.WaitGroup
	for iL, iS := range parent {
		if iL >= size {
			break
		}

		i := iL
		s := iS
		wg.Add(1)
		go func(index int, selection T) {
			defer wg.Done()

			values[index] = engine.buildObjectField(selection, cfg)
		}(i, s)
	}
	wg.Wait()

	return builder.Array(values)
}

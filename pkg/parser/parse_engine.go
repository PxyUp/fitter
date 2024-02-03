package parser

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"slices"
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

	customFillUpBaseField func(T, *config.BaseField) builder.Interfacable
	logger                logger.Logger
}

func (e *engineParser[T]) fillUpBaseField(source T, field *config.BaseField) builder.Interfacable {
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
	case config.Float, config.Float64, config.Int, config.Int64:
		float32Value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return builder.NullValue
		}
		return builder.Number(float32Value)
	case config.Array:
		return builder.PureString(text)
	case config.Object:
		return builder.PureString(text)
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildObjectField(source T, objectConfig *config.ObjectConfig, input builder.Interfacable) builder.Interfacable {
	kv := make(map[string]builder.Interfacable)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for lKey, lValue := range objectConfig.Fields {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			mutex.Lock()
			kv[k] = e.resolveField(source, v, nil, input)
			mutex.Unlock()

		}(key, value)

	}

	wg.Wait()

	return builder.Object(kv)
}

func (e *engineParser[T]) buildFirstOfBaseField(source T, fields []*config.BaseField, index *uint32, input builder.Interfacable) builder.Interfacable {
	for _, value := range fields {
		tempValue := e.buildBaseField(source, value, index, input)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildFirstOfField(parent T, fields []*config.Field, index *uint32, input builder.Interfacable) builder.Interfacable {
	for _, value := range fields {
		tempValue := e.resolveField(parent, value, index, input)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildBaseField(source T, field *config.BaseField, index *uint32, input builder.Interfacable) builder.Interfacable {
	if len(field.FirstOf) != 0 {
		return e.buildFirstOfBaseField(source, field.FirstOf, index, input)
	}

	if field.Path != "" {
		source = e.getOne(source, field.Path)
	}

	var tempValue builder.Interfacable
	if e.customFillUpBaseField != nil {
		tempValue = e.customFillUpBaseField(source, field)
	} else {
		tempValue = e.fillUpBaseField(source, field)
	}

	if field.Generated != nil {
		return buildGeneratedField(tempValue, field.Type, field.Generated, e.logger, index, input)
	}

	return tempValue
}

func (e *engineParser[T]) resolveField(parent T, field *config.Field, index *uint32, input builder.Interfacable) builder.Interfacable {
	if len(field.FirstOf) != 0 {
		return e.buildFirstOfField(parent, field.FirstOf, index, input)
	}

	if field.BaseField != nil {
		return e.buildBaseField(parent, field.BaseField, index, input)
	}

	if field.ObjectConfig != nil {
		return e.buildObjectField(parent, field.ObjectConfig, input)
	}

	if field.ArrayConfig != nil {
		return e.buildArrayField(e.getAll(parent, field.ArrayConfig.RootPath), field.ArrayConfig, input)
	}

	return builder.NullValue
}

func (e *engineParser[T]) buildStaticArray(cfg *config.StaticArrayConfig, input builder.Interfacable) builder.Interfacable {
	length := len(cfg.Items)
	if cfg.Length > 0 {
		length = int(cfg.Length)
	}
	values := make([]builder.Interfacable, length)

	var wg sync.WaitGroup

	for lKey, lValue := range cfg.Items {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k uint32, v *config.Field) {
			defer wg.Done()

			arrIndex := k
			values[k] = e.resolveField(e.parserBody, v, &arrIndex, input)

		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (e *engineParser[T]) buildArray(array *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	return e.buildArrayField(e.getAll(e.parserBody, array.RootPath), array, input)
}

func (e *engineParser[T]) buildObject(object *config.ObjectConfig, input builder.Interfacable) builder.Interfacable {
	return e.buildObjectField(e.parserBody, object, input)
}

func (e *engineParser[T]) Parse(model *config.Model, input builder.Interfacable) (*ParseResult, error) {
	if IsZero(e.parserBody) {
		return &ParseResult{
			RawResult: builder.NullValue.Raw(),
			Json:      builder.NullValue.ToJson(),
		}, nil
	}

	if model.BaseField != nil {
		res := e.buildBaseField(e.parserBody, model.BaseField, nil, input)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		res := e.buildArray(model.ArrayConfig, input)
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	res := e.buildObject(model.ObjectConfig, input)
	return &ParseResult{
		RawResult: res.Raw(),
		Json:      res.ToJson(),
	}, nil
}

func (e *engineParser[T]) buildArrayField(parent []T, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	if cfg.StaticConfig != nil {
		return e.buildStaticArray(cfg.StaticConfig, input)
	}

	if cfg.Reverse {
		slices.Reverse(parent)
	}

	size := len(parent)
	if cfg.LengthLimit > 0 {
		size = int(cfg.LengthLimit)
	}

	if cfg.ItemConfig.Field != nil {
		return FillArrayBaseField(e, parent, size, cfg.ItemConfig.Field, input)
	}

	if cfg.ItemConfig.ArrayConfig != nil {
		return FillArrayArrayField(e, parent, size, e.getAll, cfg.ItemConfig.ArrayConfig, input)
	}

	return FillArrayObjectField(e, parent, size, cfg.ItemConfig, input)
}

func FillArrayBaseField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.BaseField, input builder.Interfacable) builder.Interfacable {
	values := make([]builder.Interfacable, size)

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

			values[index] = engine.buildBaseField(selection, cfg, &arrIndex, input)
		}(i, s)

	}
	wg.Wait()

	return builder.Array(values)
}

func FillArrayArrayField[T comparable](engine *engineParser[T], parent []T, size int, fn func(T, string) []T, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	values := make([]builder.Interfacable, size)

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

			values[index] = engine.buildArrayField(fn(selection, cfg.RootPath), cfg, input)
		}(i, s)
	}
	wg.Wait()

	return builder.Array(values)
}

func FillArrayObjectField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.ObjectConfig, input builder.Interfacable) builder.Interfacable {
	values := make([]builder.Interfacable, size)

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

			values[index] = engine.buildObjectField(selection, cfg, input)
		}(i, s)
	}
	wg.Wait()

	return builder.Array(values)
}

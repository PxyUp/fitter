package parser

import (
	"context"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/tidwall/gjson"
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
	ctx                   context.Context
}

// WithContext attaches the request context so nested fetches (generated
// model fields, file downloads) inherit cancellation. Parser instances are
// short-lived (one per parsed body), so storing the context is safe.
func (e *engineParser[T]) WithContext(ctx context.Context) *engineParser[T] {
	e.ctx = ctx
	return e
}

func (e *engineParser[T]) context() context.Context {
	if e.ctx == nil {
		return context.Background()
	}
	return e.ctx
}

// checkCondition reports whether the field must be kept; evaluation errors
// omit the field instead of failing the whole parse. source is exposed to the
// expression as fSrc
func (e *engineParser[T]) checkCondition(condition string, value builder.Interfacable, source builder.Interfacable, index *uint32, input builder.Interfacable) bool {
	pass, err := utils.ProcessConditionWithSource(condition, value, source, index, input)
	if err != nil {
		e.logger.Errorw("error during process condition, field will be omitted", "error", err.Error(), "condition", condition)
		return false
	}

	return pass
}

// sourceValue exposes the current source node to condition expressions:
// structured value when the node text is valid JSON (json parser, numeric
// html text), plain string otherwise
func (e *engineParser[T]) sourceValue(source T) builder.Interfacable {
	if IsZero(source) {
		return builder.NullValue
	}

	text := e.getText(source)
	if gjson.Valid(text) {
		return builder.ToJsonableFromString(text)
	}

	return builder.String(text)
}

func unomit(value builder.Interfacable) builder.Interfacable {
	if builder.IsOmitted(value) {
		return builder.NullValue
	}

	return value
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

func (e *engineParser[T]) buildObjectField(source T, objectConfig *config.ObjectConfig, index *uint32, input builder.Interfacable) builder.Interfacable {
	if objectConfig.Condition != "" {
		sourceVal := e.sourceValue(source)
		if !e.checkCondition(objectConfig.Condition, sourceVal, sourceVal, index, input) {
			return builder.OmitValue
		}
	}

	kv := make(map[string]builder.Interfacable)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for lKey, lValue := range objectConfig.Fields {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			resolved := e.resolveField(source, v, nil, input)
			if builder.IsOmitted(resolved) {
				return
			}

			mutex.Lock()
			kv[k] = resolved
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

	parentSource := source
	if field.Path != "" {
		source = e.getOne(source, field.Path)
	}

	var tempValue builder.Interfacable
	if e.customFillUpBaseField != nil {
		tempValue = e.customFillUpBaseField(source, field)
	} else {
		tempValue = e.fillUpBaseField(source, field)
	}

	if field.Condition != "" && !e.checkCondition(field.Condition, tempValue, e.sourceValue(parentSource), index, input) {
		return builder.OmitValue
	}

	if field.Generated != nil {
		return buildGeneratedField(e.context(), tempValue, field.Type, field.Generated, e.logger, index, input)
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
		return e.buildObjectField(parent, field.ObjectConfig, index, input)
	}

	if field.ArrayConfig != nil {
		if field.ArrayConfig.Condition != "" {
			sourceVal := e.sourceValue(parent)
			if !e.checkCondition(field.ArrayConfig.Condition, sourceVal, sourceVal, index, input) {
				return builder.OmitValue
			}
		}

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

	// static arrays are positional: omitted/unset slots stay null instead of shifting indexes
	for i, v := range values {
		if v == nil || builder.IsOmitted(v) {
			values[i] = builder.NullValue
		}
	}

	return builder.Array(values)
}

func (e *engineParser[T]) buildArray(array *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	if array.Condition != "" {
		sourceVal := e.sourceValue(e.parserBody)
		if !e.checkCondition(array.Condition, sourceVal, sourceVal, nil, input) {
			return builder.OmitValue
		}
	}

	return e.buildArrayField(e.getAll(e.parserBody, array.RootPath), array, input)
}

func (e *engineParser[T]) buildObject(object *config.ObjectConfig, input builder.Interfacable) builder.Interfacable {
	return e.buildObjectField(e.parserBody, object, nil, input)
}

func (e *engineParser[T]) Parse(model *config.Model, input builder.Interfacable) (*ParseResult, error) {
	if IsZero(e.parserBody) {
		return &ParseResult{
			RawResult: builder.NullValue.Raw(),
			Json:      builder.NullValue.ToJson(),
		}, nil
	}

	if model.BaseField != nil {
		res := unomit(e.buildBaseField(e.parserBody, model.BaseField, nil, input))
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		res := unomit(e.buildArray(model.ArrayConfig, input))
		return &ParseResult{
			RawResult: res.Raw(),
			Json:      res.ToJson(),
		}, nil
	}

	res := unomit(e.buildObject(model.ObjectConfig, input))
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
		return FillArrayBaseField(e, parent, size, cfg, input)
	}

	if cfg.ItemConfig.ArrayConfig != nil {
		return FillArrayArrayField(e, parent, size, e.getAll, cfg, input)
	}

	return FillArrayObjectField(e, parent, size, cfg, input)
}

// finalizeArrayItems drops items rejected by a condition (omitted or failing
// item_condition); nil slots stay null to preserve the declared size when
// length_limit exceeds the amount of source elements. parent aligns with
// values by index and feeds fSrc for item_condition
func finalizeArrayItems[T comparable](engine *engineParser[T], parent []T, values []builder.Interfacable, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	res := make([]builder.Interfacable, 0, len(values))
	for i, v := range values {
		if v == nil {
			v = builder.NullValue
		}

		if builder.IsOmitted(v) {
			continue
		}

		if cfg.ItemCondition != "" {
			arrIndex := uint32(i)
			var src builder.Interfacable = builder.NullValue
			if i < len(parent) {
				src = engine.sourceValue(parent[i])
			}
			if !engine.checkCondition(cfg.ItemCondition, v, src, &arrIndex, input) {
				continue
			}
		}

		res = append(res, v)
	}

	return builder.Array(res)
}

func FillArrayBaseField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
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

			values[index] = engine.buildBaseField(selection, cfg.ItemConfig.Field, &arrIndex, input)
		}(i, s)

	}
	wg.Wait()

	return finalizeArrayItems(engine, parent, values, cfg, input)
}

func FillArrayArrayField[T comparable](engine *engineParser[T], parent []T, size int, fn func(T, string) []T, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
	inner := cfg.ItemConfig.ArrayConfig
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

			if inner.Condition != "" {
				arrIndex := uint32(index)
				sourceVal := engine.sourceValue(selection)
				if !engine.checkCondition(inner.Condition, sourceVal, sourceVal, &arrIndex, input) {
					values[index] = builder.OmitValue
					return
				}
			}

			values[index] = engine.buildArrayField(fn(selection, inner.RootPath), inner, input)
		}(i, s)
	}
	wg.Wait()

	return finalizeArrayItems(engine, parent, values, cfg, input)
}

func FillArrayObjectField[T comparable](engine *engineParser[T], parent []T, size int, cfg *config.ArrayConfig, input builder.Interfacable) builder.Interfacable {
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

			arrIndex := uint32(index)

			values[index] = engine.buildObjectField(selection, cfg.ItemConfig, &arrIndex, input)
		}(i, s)
	}
	wg.Wait()

	return finalizeArrayItems(engine, parent, values, cfg, input)
}

package parser

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
)

type htmlParser struct {
	logger     logger.Logger
	body       []byte
	parserBody *goquery.Selection
}

func newHTML(body []byte) *htmlParser {
	document, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))
	return &htmlParser{
		body:       body,
		logger:     logger.Null,
		parserBody: document.Selection,
	}
}

func (h *htmlParser) WithLogger(logger logger.Logger) *htmlParser {
	h.logger = logger
	return h
}

func (h *htmlParser) Parse(model *config.Model) (*ParseResult, error) {
	if h.parserBody == nil {
		return &ParseResult{
			Raw: builder.Null().ToJson(),
		}, nil
	}

	if model.BaseField != nil {
		return &ParseResult{
			Raw: h.buildBaseField(h.parserBody, model.BaseField, nil).ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		return &ParseResult{
			Raw: h.buildArray(model.ArrayConfig).ToJson(),
		}, nil
	}
	return &ParseResult{
		Raw: h.buildObject(model.ObjectConfig).ToJson(),
	}, nil
}

func (h *htmlParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	return h.buildArrayField(h.parserBody, array)
}

func (h *htmlParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return h.buildObjectField(h.parserBody, object)
}

func (h *htmlParser) buildObjectField(parent *goquery.Selection, object *config.ObjectConfig) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for lKey, lValue := range object.Fields {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			mutex.Lock()
			kv[k] = h.resolveField(parent, v, nil)
			mutex.Unlock()
		}(key, value)
	}

	wg.Wait()

	return builder.Object(kv)
}

func (h *htmlParser) buildStaticArray(cfg *config.StaticArrayConfig) builder.Jsonable {
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
			values[int(k)] = h.resolveField(h.parserBody, v, &arrIndex)
		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (h *htmlParser) buildFirstOfField(parent *goquery.Selection, fields []*config.Field, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := h.resolveField(parent, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (h *htmlParser) resolveField(parent *goquery.Selection, field *config.Field, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return h.buildFirstOfField(parent, field.FirstOf, index)
	}

	if field.BaseField != nil {
		return h.buildBaseField(parent, field.BaseField, index)
	}

	if field.ObjectConfig != nil {
		return h.buildObjectField(parent, field.ObjectConfig)
	}

	if field.ArrayConfig != nil {
		return h.buildArrayField(parent, field.ArrayConfig)
	}
	return builder.Null()
}

func (h *htmlParser) buildArrayField(parent *goquery.Selection, array *config.ArrayConfig) builder.Jsonable {
	if array.StaticConfig != nil {
		return h.buildStaticArray(array.StaticConfig)
	}

	if array.RootPath != "" {
		parent = parent.Find(array.RootPath)
	}

	size := parent.Length()
	if array.LengthLimit > 0 {
		size = int(array.LengthLimit)
	}

	values := make([]builder.Jsonable, size)

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		parent.Each(func(i int, s *goquery.Selection) {
			if i >= size {
				return
			}
			wg.Add(1)
			go func(index int, selection *goquery.Selection) {
				defer wg.Done()

				arrIndex := uint32(index)
				values[index] = h.buildBaseField(selection, array.ItemConfig.Field, &arrIndex)
			}(i, s)

		})
		wg.Wait()
		return builder.Array(values)
	}

	if array.ItemConfig.ArrayConfig != nil {
		var wg sync.WaitGroup
		parent.Each(func(i int, s *goquery.Selection) {
			if i >= size {
				return
			}

			wg.Add(1)
			go func(index int, selection *goquery.Selection) {
				defer wg.Done()

				values[index] = h.buildArrayField(selection, array.ItemConfig.ArrayConfig)
			}(i, s)
		})
		wg.Wait()
		return builder.Array(values)
	}

	var wg sync.WaitGroup
	parent.Each(func(i int, s *goquery.Selection) {
		if i >= size {
			return
		}

		wg.Add(1)
		go func(index int, selection *goquery.Selection) {
			defer wg.Done()

			values[index] = h.buildObjectField(selection, array.ItemConfig)
		}(i, s)
	})
	wg.Wait()

	return builder.Array(values)
}

func (h *htmlParser) fillUpBaseField(source *goquery.Selection, field *config.BaseField) builder.Jsonable {
	if source.Length() <= 0 {
		return builder.Null()
	}

	text := source.First().Text()

	switch field.Type {
	case config.Null:
		return builder.Null()
	case config.String:
		return builder.String(text)
	case config.Bool:
		boolValue, err := strconv.ParseBool(text)
		if err != nil {
			return builder.Null()
		}
		return builder.Bool(boolValue)
	case config.Float:
		float32Value, err := strconv.ParseFloat(text, 32)
		if err != nil {
			return builder.Null()
		}
		return builder.Float(float32(float32Value))
	case config.Int:
		intValue, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return builder.Null()
		}
		return builder.Int(int(intValue))
	}

	return builder.Null()
}

func (h *htmlParser) buildFirstOfBaseField(source *goquery.Selection, fields []*config.BaseField, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := h.buildBaseField(source, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (h *htmlParser) buildBaseField(source *goquery.Selection, field *config.BaseField, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return h.buildFirstOfBaseField(source, field.FirstOf, index)
	}

	if field.Path != "" {
		source = source.Find(field.Path)
	}

	tempValue := h.fillUpBaseField(source, field)

	if field.Generated != nil {
		if field.Type == config.String {
			tempValue = builder.PureString(tempValue.ToJson())
		}
		generatedValue := buildGeneratedField(tempValue, field.Generated, h.logger, index)
		if field.Generated.Model != nil {
			if field.Generated.Model.Type == config.Array || field.Generated.Model.Type == config.Object {
				if field.Generated.Model.Path != "" {
					return builder.PureString(gjson.Parse(generatedValue.ToJson()).Get(field.Generated.Model.Path).Raw)
				}
				return generatedValue
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

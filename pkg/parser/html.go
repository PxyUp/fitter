package parser

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
	"strconv"
	"sync"
)

type htmlParser struct {
	logger     logger.Logger
	body       []byte
	parserBody *goquery.Selection
}

func NewHTML(body []byte) *htmlParser {
	document, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))
	return &htmlParser{
		body:       body,
		logger:     logger.Null,
		parserBody: document.Selection,
	}
}

func (h *htmlParser) WithLogger(logger logger.Logger) {
	h.logger = logger
}

func (h *htmlParser) Parse(model *config.Model) (*ParseResult, error) {
	if h.parserBody == nil {
		return &ParseResult{
			Raw: builder.Null().ToJson(),
		}, nil
	}

	if model.Type == config.ArrayModel {
		return &ParseResult{
			Raw: h.buildArray(model.ArrayConfig).ToJson(),
		}, nil
	}
	return &ParseResult{
		Raw: h.buildObject(model.ObjectConfig).ToJson(),
	}, nil
}

func (h *htmlParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	if array.RootPath == "" {
		return h.buildArrayField(h.parserBody, array)
	}
	return h.buildArrayField(h.parserBody.Find(array.RootPath), array)
}

func (h *htmlParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return h.buildObjectField(h.parserBody, object.Fields)
}

func (h *htmlParser) buildObjectField(parent *goquery.Selection, fields map[string]*config.Field) builder.Jsonable {
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
					kv[k] = h.buildBaseField(parent, v.BaseField)
					return
				}
				kv[k] = h.buildBaseField(parent.Find(v.BaseField.Path), v.BaseField)
				return
			}

			if v.ObjectConfig != nil {
				kv[k] = h.buildObjectField(parent, v.ObjectConfig.Fields)
				return
			}

			if v.ArrayConfig != nil {
				if v.ArrayConfig.RootPath == "" {
					kv[k] = h.buildArrayField(parent, v.ArrayConfig)
					return
				}
				kv[k] = h.buildArrayField(parent.Find(v.ArrayConfig.RootPath), v.ArrayConfig)
			}
		}(key, value)

	}

	wg.Wait()

	return builder.Object(kv)
}

func (h *htmlParser) buildArrayField(parent *goquery.Selection, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, parent.Length())
	
	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		parent.Each(func(i int, s *goquery.Selection) {
			wg.Add(1)
			go func(index int, selection *goquery.Selection) {
				defer wg.Done()

				if array.ItemConfig.Field.Path == "" {
					values[index] = h.buildBaseField(selection, array.ItemConfig.Field)
					return
				}
				values[index] = h.buildBaseField(selection.Find(array.ItemConfig.Field.Path), array.ItemConfig.Field)
			}(i, s)

		})
		wg.Wait()
		return builder.Array(values)
	}

	if array.ItemConfig.ArrayConfig != nil {
		var wg sync.WaitGroup
		parent.Each(func(i int, s *goquery.Selection) {
			wg.Add(1)
			go func(index int, selection *goquery.Selection) {
				defer wg.Done()
				if array.ItemConfig.ArrayConfig.RootPath == "" {
					values[index] = h.buildArrayField(selection, array.ItemConfig.ArrayConfig)
				} else {
					values[index] = h.buildArrayField(selection.Find(array.ItemConfig.ArrayConfig.RootPath), array.ItemConfig.ArrayConfig)
				}
			}(i, s)
		})
		wg.Wait()
		return builder.Array(values)
	}

	var wg sync.WaitGroup
	parent.Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func(index int, selection *goquery.Selection) {
			defer wg.Done()

			values[index] = h.buildObjectField(selection, array.ItemConfig.Fields)
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

func (h *htmlParser) buildBaseField(source *goquery.Selection, field *config.BaseField) builder.Jsonable {
	tempValue := h.fillUpBaseField(source, field)

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

package parser

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"strconv"
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
	for k, v := range fields {
		if v.BaseField != nil {
			if v.BaseField.Path == "" {
				kv[k] = h.buildBaseField(parent, v.BaseField)
				continue
			}
			kv[k] = h.buildBaseField(parent.Find(v.BaseField.Path), v.BaseField)
			continue
		}

		if v.ObjectConfig != nil {
			kv[k] = h.buildObjectField(parent, v.ObjectConfig.Fields)
			continue
		}

		if v.ArrayConfig != nil {
			if v.ArrayConfig.RootPath == "" {
				kv[k] = h.buildArrayField(parent, v.ArrayConfig)
				continue
			}
			kv[k] = h.buildArrayField(parent.Find(v.ArrayConfig.RootPath), v.ArrayConfig)
		}
	}

	return builder.Object(kv)
}

func (h *htmlParser) buildArrayField(parent *goquery.Selection, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, parent.Length())
	if array.ItemConfig.Field != nil {
		parent.Each(func(index int, selection *goquery.Selection) {
			if array.ItemConfig.Field.Path == "" {
				values[index] = h.buildBaseField(selection, array.ItemConfig.Field)
				return
			}
			values[index] = h.buildBaseField(selection.Find(array.ItemConfig.Field.Path), array.ItemConfig.Field)
		})
		return builder.Array(values)
	}

	parent.Each(func(index int, selection *goquery.Selection) {
		values[index] = h.buildObjectField(selection, array.ItemConfig.Fields)
	})

	return builder.Array(values)
}

func (h *htmlParser) buildBaseField(source *goquery.Selection, field *config.BaseField) builder.Jsonable {
	if field.Generated != nil {
		return buildGeneratedField(field.Generated)
	}

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

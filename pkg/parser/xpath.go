package parser

import (
	"bytes"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/antchfx/htmlquery"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
	"strconv"
	"sync"
)

type xpathParser struct {
	logger     logger.Logger
	body       []byte
	parserBody *html.Node

	safeFind    func(top *html.Node, expr string) []*html.Node
	safeFindOne func(top *html.Node, expr string) *html.Node
}

func NewXPath(body []byte) *xpathParser {
	document, _ := htmlquery.Parse(bytes.NewReader(body))
	return &xpathParser{
		body:       body,
		logger:     logger.Null,
		parserBody: document,
		safeFind: func(top *html.Node, expr string) []*html.Node {
			nodes, err := htmlquery.QueryAll(top, expr)
			if err != nil {
				return nil
			}
			return nodes
		},
		safeFindOne: func(top *html.Node, expr string) *html.Node {
			node, err := htmlquery.Query(top, expr)
			if err != nil {
				return nil
			}
			return node
		},
	}
}

func (x *xpathParser) WithLogger(logger logger.Logger) *xpathParser {
	x.logger = logger
	return x
}

func (x *xpathParser) Parse(model *config.Model) (*ParseResult, error) {
	if x.parserBody == nil {
		return &ParseResult{
			Raw: builder.Null().ToJson(),
		}, nil
	}

	if model.Type == config.ArrayModel {
		return &ParseResult{
			Raw: x.buildArray(model.ArrayConfig).ToJson(),
		}, nil
	}
	return &ParseResult{
		Raw: x.buildObject(model.ObjectConfig).ToJson(),
	}, nil
}

func (x *xpathParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	if array.RootPath == "" {
		return x.buildArrayField(x.safeFind(x.parserBody, "."), array)
	}
	return x.buildArrayField(x.safeFind(x.parserBody, array.RootPath), array)
}

func (x *xpathParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return x.buildObjectField(x.parserBody, object.Fields)
}

func (x *xpathParser) buildObjectField(parent *html.Node, fields map[string]*config.Field) builder.Jsonable {
	kv := make(map[string]builder.Jsonable)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for lKey, lValue := range fields {
		key := lKey
		value := lValue
		wg.Add(1)
		go func(k string, v *config.Field) {
			defer wg.Done()

			var fnValue builder.Jsonable
			defer func() {
				if fnValue == nil {
					return
				}
				mutex.Lock()
				kv[k] = fnValue
				mutex.Unlock()
			}()

			if v.BaseField != nil {
				if v.BaseField.Path == "" {
					fnValue = x.buildBaseField(parent, v.BaseField)
					return
				}
				fnValue = x.buildBaseField(x.safeFindOne(parent, v.BaseField.Path), v.BaseField)
				return
			}

			if v.ObjectConfig != nil {
				fnValue = x.buildObjectField(parent, v.ObjectConfig.Fields)
				return
			}

			if v.ArrayConfig != nil {
				if v.ArrayConfig.RootPath == "" {
					fnValue = x.buildArrayField(x.safeFind(parent, "."), v.ArrayConfig)
					return
				}
				fnValue = x.buildArrayField(x.safeFind(parent, v.ArrayConfig.RootPath), v.ArrayConfig)
			}
		}(key, value)

	}

	wg.Wait()

	return builder.Object(kv)
}

func (x *xpathParser) buildArrayField(parent []*html.Node, array *config.ArrayConfig) builder.Jsonable {
	values := make([]builder.Jsonable, len(parent))

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		for iL, sL := range parent {
			i := iL
			s := sL
			wg.Add(1)
			go func(index int, selection *html.Node) {
				defer wg.Done()

				if array.ItemConfig.Field.Path == "" {
					values[index] = x.buildBaseField(selection, array.ItemConfig.Field)
					return
				}
				values[index] = x.buildBaseField(x.safeFindOne(selection, array.ItemConfig.Field.Path), array.ItemConfig.Field)
			}(i, s)

		}
		wg.Wait()
		return builder.Array(values)
	}

	if array.ItemConfig.ArrayConfig != nil {
		var wg sync.WaitGroup
		for iL, iS := range parent {
			i := iL
			s := iS
			wg.Add(1)
			go func(index int, selection *html.Node) {
				defer wg.Done()
				if array.ItemConfig.ArrayConfig.RootPath == "" {
					values[index] = x.buildArrayField(x.safeFind(selection, "."), array.ItemConfig.ArrayConfig)
				} else {
					values[index] = x.buildArrayField(x.safeFind(selection, array.ItemConfig.ArrayConfig.RootPath), array.ItemConfig.ArrayConfig)
				}
			}(i, s)
		}
		wg.Wait()
		return builder.Array(values)
	}

	var wg sync.WaitGroup
	for iL, iS := range parent {
		i := iL
		s := iS
		wg.Add(1)
		go func(index int, selection *html.Node) {
			defer wg.Done()

			values[index] = x.buildObjectField(selection, array.ItemConfig.Fields)
		}(i, s)
	}
	wg.Wait()

	return builder.Array(values)
}

func (x *xpathParser) fillUpBaseField(source *html.Node, field *config.BaseField) builder.Jsonable {
	if source == nil {
		return builder.Null()
	}

	text := htmlquery.InnerText(source)

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

func (x *xpathParser) buildBaseField(source *html.Node, field *config.BaseField) builder.Jsonable {
	tempValue := x.fillUpBaseField(source, field)

	if field.Generated != nil {
		if field.Type == config.String {
			tempValue = builder.PureString(tempValue.ToJson())
		}
		generatedValue := buildGeneratedField(tempValue, field.Generated, x.logger)
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
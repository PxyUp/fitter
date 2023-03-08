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

func newXPath(body []byte) *xpathParser {
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

	if model.BaseField != nil {
		return &ParseResult{
			Raw: x.buildBaseField(x.parserBody, model.BaseField, nil).ToJson(),
		}, nil
	}

	if model.ArrayConfig != nil {
		return &ParseResult{
			Raw: x.buildArray(model.ArrayConfig).ToJson(),
		}, nil
	}
	return &ParseResult{
		Raw: x.buildObject(model.ObjectConfig).ToJson(),
	}, nil
}

func (x *xpathParser) buildArray(array *config.ArrayConfig) builder.Jsonable {
	return x.buildArrayField(x.safeFind(x.parserBody, array.RootPath), array)
}

func (x *xpathParser) buildObject(object *config.ObjectConfig) builder.Jsonable {
	return x.buildObjectField(x.parserBody, object.Fields)
}

func (x *xpathParser) buildStaticArray(cfg *config.StaticArrayConfig) builder.Jsonable {
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
			values[k] = x.resolveField(x.parserBody, v, &arrIndex)

		}(key, value)

	}

	wg.Wait()

	return builder.Array(values)
}

func (x *xpathParser) buildFirstOfField(parent *html.Node, fields []*config.Field, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := x.resolveField(parent, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (x *xpathParser) resolveField(parent *html.Node, field *config.Field, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return x.buildFirstOfField(parent, field.FirstOf, index)
	}

	if field.BaseField != nil {
		return x.buildBaseField(parent, field.BaseField, index)
	}

	if field.ObjectConfig != nil {
		return x.buildObjectField(parent, field.ObjectConfig.Fields)
	}

	if field.ArrayConfig != nil {
		return x.buildArrayField(x.safeFind(parent, field.ArrayConfig.RootPath), field.ArrayConfig)
	}

	return builder.Null()
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

			mutex.Lock()
			kv[k] = x.resolveField(parent, v, nil)
			mutex.Unlock()

		}(key, value)

	}

	wg.Wait()

	return builder.Object(kv)
}

func (x *xpathParser) buildArrayField(parent []*html.Node, array *config.ArrayConfig) builder.Jsonable {
	if array.StaticConfig != nil {
		return x.buildStaticArray(array.StaticConfig)
	}

	values := make([]builder.Jsonable, len(parent))

	if array.ItemConfig.Field != nil {
		var wg sync.WaitGroup
		for iL, sL := range parent {
			i := iL
			s := sL
			wg.Add(1)
			go func(index int, selection *html.Node) {
				defer wg.Done()

				arrIndex := uint32(index)

				values[index] = x.buildBaseField(selection, array.ItemConfig.Field, &arrIndex)
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

				values[index] = x.buildArrayField(x.safeFind(selection, array.ItemConfig.ArrayConfig.RootPath), array.ItemConfig.ArrayConfig)
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

func (x *xpathParser) buildFirstOfBaseField(source *html.Node, fields []*config.BaseField, index *uint32) builder.Jsonable {
	for _, value := range fields {
		tempValue := x.buildBaseField(source, value, index)
		if !tempValue.IsEmpty() {
			return tempValue
		}
	}

	return builder.Null()
}

func (x *xpathParser) buildBaseField(source *html.Node, field *config.BaseField, index *uint32) builder.Jsonable {
	if len(field.FirstOf) != 0 {
		return x.buildFirstOfBaseField(source, field.FirstOf, index)
	}

	if field.Path != "" {
		source = x.safeFindOne(source, field.Path)
	}

	tempValue := x.fillUpBaseField(source, field)

	if field.Generated != nil {
		if field.Type == config.String {
			tempValue = builder.PureString(tempValue.ToJson())
		}
		generatedValue := buildGeneratedField(tempValue, field.Generated, x.logger, index)
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

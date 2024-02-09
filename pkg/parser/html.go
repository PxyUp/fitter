package parser

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"strconv"
)

func selectionToArray(parent *goquery.Selection) []*goquery.Selection {
	tmp := make([]*goquery.Selection, len(parent.Nodes))

	parent.Each(func(i int, selection *goquery.Selection) {
		tmp[i] = selection
	})

	return tmp
}

func htmlFillUpBaseField(source *goquery.Selection, field *config.BaseField) builder.Interfacable {
	if source.Length() <= 0 {
		return builder.NullValue
	}

	if field.Type == config.HtmlString {
		htmlString, err := source.Html()
		if err != nil {
			return builder.NullValue
		}
		return builder.String(htmlString)
	}

	var text string

	if field.HTMLAttribute != "" {
		attrValue, attrExists := source.First().Attr(field.HTMLAttribute)
		if !attrExists {
			return builder.NullValue
		}
		text = attrValue
	} else {
		text = source.First().Text()
	}

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
	case config.Array, config.Object:
		return builder.ToJsonableFromString(text)
	}

	return builder.NullValue
}

func NewHTML(body []byte, logger logger.Logger) *engineParser[*goquery.Selection] {
	document, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	return &engineParser[*goquery.Selection]{
		getText: func(r *goquery.Selection) string {
			return r.First().Text()
		},
		parserBody: document.Selection,
		logger:     logger,
		getAll: func(parent *goquery.Selection, path string) []*goquery.Selection {
			if path == "" {
				return selectionToArray(parent)
			}

			res := parent.Find(path)
			return selectionToArray(res)
		},
		getOne: func(parent *goquery.Selection, path string) *goquery.Selection {
			if path == "" {
				return parent
			}
			return parent.Find(path)
		},
		customFillUpBaseField: htmlFillUpBaseField,
	}
}

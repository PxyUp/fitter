package parser

import (
	"bytes"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func NewXPath(body []byte, logger logger.Logger) *engineParser[*html.Node] {
	document, _ := htmlquery.Parse(bytes.NewReader(body))

	return &engineParser[*html.Node]{
		getText:    htmlquery.InnerText,
		parserBody: document,
		logger:     logger,
		getAll: func(top *html.Node, expr string) []*html.Node {
			nodes, err := htmlquery.QueryAll(top, expr)
			if err != nil {
				return nil
			}
			return nodes
		},
		getOne: func(top *html.Node, expr string) *html.Node {
			node, err := htmlquery.Query(top, expr)
			if err != nil {
				return nil
			}
			return node
		},
	}
}

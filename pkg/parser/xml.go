package parser

import (
	"bytes"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/antchfx/xmlquery"
)

func NewXML(body []byte, logger logger.Logger) *engineParser[*xmlquery.Node] {
	document, _ := xmlquery.Parse(bytes.NewReader(body))

	return &engineParser[*xmlquery.Node]{
		getText: func(node *xmlquery.Node) string {
			return node.InnerText()
		},
		parserBody: document,
		logger:     logger,
		getAll: func(top *xmlquery.Node, expr string) []*xmlquery.Node {
			nodes, err := xmlquery.QueryAll(top, expr)
			if err != nil {
				return nil
			}
			return nodes
		},
		getOne: func(top *xmlquery.Node, expr string) *xmlquery.Node {
			node, err := xmlquery.Query(top, expr)
			if err != nil {
				return nil
			}
			return node
		},
	}
}

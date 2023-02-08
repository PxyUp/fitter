package parser

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/parser/builder"
)

type Factory func([]byte) Parser

type Parser interface {
	Parse(model *config.Model) (*ParseResult, error)
}

type ParseResult struct {
	Raw string `json:"raw"`
}

func buildGeneratedField(field *config.GeneratedFieldConfig) builder.Jsonable {
	if field.UUID != nil {
		return builder.UUID(field.UUID)
	}

	if field.Static != nil {
		return builder.Static(field.Static)
	}

	return builder.Null()
}

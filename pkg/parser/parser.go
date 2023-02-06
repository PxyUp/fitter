package parser

import "github.com/PxyUp/fitter/pkg/config"

type Parser interface {
	Parse(model config.Model) (*ParseResult, error)
}

type ParseResult struct {
	Raw string
}

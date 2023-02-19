package parser

import (
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"strings"
)

const (
	placeHolder = "{PL}"
)

func format(str string, value builder.Jsonable) string {
	if strings.Contains(str, placeHolder) && value != nil && value.ToJson() != builder.EmptyString {
		return strings.ReplaceAll(str, placeHolder, value.ToJson())
	}

	return str
}

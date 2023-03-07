package parser

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"strings"
)

const (
	placeHolder           = "{PL}"
	indexPlaceHolder      = "{INDEX}"
	humanIndexPlaceHolder = "{HUMAN_INDEX}"
)

func format(str string, value builder.Jsonable, index *uint32) string {
	if strings.Contains(str, placeHolder) && value != nil && value.ToJson() != builder.EmptyString {
		str = strings.ReplaceAll(str, placeHolder, value.ToJson())
	}

	if strings.Contains(str, indexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, indexPlaceHolder, fmt.Sprintf("%d", *index))
	}

	if strings.Contains(str, humanIndexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, humanIndexPlaceHolder, fmt.Sprintf("%d", *index+1))
	}

	return str
}

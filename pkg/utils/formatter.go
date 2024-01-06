package utils

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/tidwall/gjson"
	"strings"
)

const (
	jsonPathStart         = "{{{"
	jsonPathEnd           = "}}}"
	placeHolder           = "{PL}"
	indexPlaceHolder      = "{INDEX}"
	humanIndexPlaceHolder = "{HUMAN_INDEX}"
)

func Format(str string, value builder.Jsonable, index *uint32) string {
	if len(str) == 0 {
		return str
	}

	if strings.Contains(str, placeHolder) && value != nil && value.ToJson() != builder.EmptyString {
		str = strings.ReplaceAll(str, placeHolder, value.ToJson())
	}

	if strings.Contains(str, indexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, indexPlaceHolder, fmt.Sprintf("%d", *index))
	}

	if strings.Contains(str, humanIndexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, humanIndexPlaceHolder, fmt.Sprintf("%d", *index+1))
	}

	return formatJsonPathString(str, value)
}

func formatJsonPathString(str string, value builder.Jsonable) string {
	new := ""
	runes := []rune(str)
	isInJSONPath := false
	path := ""
	for i := 0; i < len(runes); i++ {
		if !isInJSONPath {
			new += string(runes[i])
		} else {
			path += string(runes[i])
		}

		if isInJSONPath && strings.HasSuffix(path, jsonPathEnd) {
			path = strings.TrimSuffix(path, jsonPathEnd)
			isInJSONPath = false
			new += gjson.Parse(value.ToJson()).Get(path).String()
			path = ""
		}

		if !isInJSONPath && strings.HasSuffix(new, jsonPathStart) {
			new = strings.TrimSuffix(new, jsonPathStart)
			isInJSONPath = true
		}

	}

	if isInJSONPath {
		return str
	}

	return new

}

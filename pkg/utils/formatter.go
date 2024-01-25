package utils

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/references"
	"github.com/tidwall/gjson"
	"html"
	"os"
	"strings"
)

const (
	jsonPathStart         = "{{{"
	jsonPathEnd           = "}}}"
	placeHolder           = "{PL}"
	indexPlaceHolder      = "{INDEX}"
	humanIndexPlaceHolder = "{HUMAN_INDEX}"
	refNamePrefix         = "RefName="
	envNamePrefix         = "FromEnv="
)

func Format(str string, value builder.Jsonable, index *uint32) string {
	if len(str) == 0 {
		return str
	}

	if strings.Contains(str, placeHolder) && value != nil && (value.ToJson() != builder.EmptyString && len(value.ToJson()) != 0) {
		str = strings.ReplaceAll(str, placeHolder, html.UnescapeString(value.ToJson()))
	}

	if strings.Contains(str, indexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, indexPlaceHolder, fmt.Sprintf("%d", *index))
	}

	if strings.Contains(str, humanIndexPlaceHolder) && index != nil {
		str = strings.ReplaceAll(str, humanIndexPlaceHolder, fmt.Sprintf("%d", *index+1))
	}

	return formatJsonPathString(str, value)
}

func processPrefix(prefix string, value builder.Jsonable) string {
	if strings.HasPrefix(prefix, refNamePrefix) {
		refValue := strings.Split(strings.TrimPrefix(prefix, refNamePrefix), " ")
		tmp := ""
		if len(refValue) > 1 {
			tmp = gjson.Parse(html.UnescapeString(references.Get(refValue[0]).ToJson())).Get(refValue[1]).String()
		}
		if len(refValue) == 1 {
			tmp = html.UnescapeString(references.Get(refValue[0]).ToJson())
		}

		return builder.PureString(tmp).ToJson()
	}

	if strings.HasPrefix(prefix, envNamePrefix) {
		envValue := strings.TrimPrefix(prefix, envNamePrefix)
		return builder.PureString(os.Getenv(envValue)).ToJson()
	}

	return gjson.Parse(value.ToJson()).Get(prefix).String()
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

			new += processPrefix(path, value)

			isInJSONPath = false
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

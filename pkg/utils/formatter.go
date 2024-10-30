package utils

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/references"
	"github.com/tidwall/gjson"
	"html"
	"io"
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
	exprNamePrefix        = "FromExp="
	inputNamePrefix       = "FromInput="
	inputFilePrefix       = "FromFile="
	inputURLPrefix        = "FromURL="
)

var (
	formatterLogger = logger.Null
)

func SetLogger(lvl string) {
	formatterLogger = logger.NewLogger(lvl).With("component", "formatter")
}

func Format(str string, value builder.Interfacable, index *uint32, input builder.Interfacable) string {
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

	return strings.ReplaceAll(formatJsonPathString(str, value, index, input), fitterNewLinePlaceholderValue, "\n")
}

func processPrefix(prefix string, value builder.Interfacable, index *uint32, input builder.Interfacable) string {
	if strings.HasPrefix(prefix, inputNamePrefix) {
		path := strings.TrimPrefix(prefix, inputNamePrefix)
		tmp := ""
		if input != nil {
			if path == "" || path == "." {
				tmp = html.UnescapeString(gjson.Parse(input.ToJson()).String())
			} else {
				tmp = gjson.Parse(html.UnescapeString(input.ToJson())).Get(path).String()
			}
		}

		return builder.PureString(tmp).ToJson()
	}

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

	if strings.HasPrefix(prefix, exprNamePrefix) {
		expression := strings.TrimPrefix(prefix, exprNamePrefix)
		raw, err := ProcessExpression(expression, value, index, input)
		if err != nil {
			formatterLogger.Errorw("cant process expression", "value", expression, "error", err.Error())
			return builder.EMPTY.ToJson()
		}

		return builder.PureString(raw.ToJson()).ToJson()
	}

	if strings.HasPrefix(prefix, inputFilePrefix) {
		filePath := strings.TrimPrefix(prefix, inputFilePrefix)
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			formatterLogger.Errorw("cant file expression", "file_path", filePath, "error", err.Error())
			return builder.EMPTY.ToJson()
		}

		return builder.PureString(Format(string(fileContent), value, index, input)).ToJson()
	}

	if strings.HasPrefix(prefix, envNamePrefix) {
		envValue := strings.TrimPrefix(prefix, envNamePrefix)
		return builder.PureString(os.Getenv(envValue)).ToJson()
	}

	if strings.HasPrefix(prefix, inputURLPrefix) {
		urlPath := strings.TrimPrefix(prefix, inputURLPrefix)

		resp, err := http_client.GetDefaultClient().Get(urlPath)
		if err != nil {
			formatterLogger.Errorw("cant process url", "url_path", urlPath, "error", err.Error())
			return builder.EMPTY.ToJson()
		}

		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			formatterLogger.Errorw("cant read url response", "url_path", urlPath, "error", err.Error())
			return builder.EMPTY.ToJson()
		}

		return builder.PureString(Format(string(content), value, index, input)).ToJson()
	}

	if value == nil {
		return builder.EMPTY.ToJson()
	}

	return gjson.Parse(value.ToJson()).Get(prefix).String()
}

func formatJsonPathString(str string, value builder.Interfacable, index *uint32, input builder.Interfacable) string {
	runes := []rune(str)
	stack := []string{
		"",
	}

	for i := 0; i < len(runes); i++ {
		stack[len(stack)-1] += string(runes[i])
		last := stack[len(stack)-1]

		if len(stack) > 1 && strings.HasSuffix(last, jsonPathEnd) {
			tmp := processPrefix(strings.TrimSuffix(last, jsonPathEnd), value, index, input)
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] += tmp
		}

		if strings.HasSuffix(last, jsonPathStart) {
			stack[len(stack)-1] = strings.TrimSuffix(last, jsonPathStart)
			stack = append(stack, "")
		}

	}

	return stack[0]
}

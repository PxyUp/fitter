package parser

import (
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/tidwall/gjson"
)

func gsjsonToArray(value *gjson.Result) []*gjson.Result {
	tmp := value.Array()
	answer := make([]*gjson.Result, len(tmp))

	for i, v := range tmp {
		lv := v
		answer[i] = &lv
	}

	return answer
}

func NewJson(body []byte, logger logger.Logger) *engineParser[*gjson.Result] {
	bb := gjson.ParseBytes(body)
	return &engineParser[*gjson.Result]{
		getText: func(r *gjson.Result) string {
			return r.String()
		},
		parserBody: &bb,
		logger:     logger,
		getAll: func(parent *gjson.Result, path string) []*gjson.Result {
			if path == "" {
				return gsjsonToArray(parent)
			}

			res := parent.Get(path)
			return gsjsonToArray(&res)
		},
		getOne: func(parent *gjson.Result, path string) *gjson.Result {
			v := parent.Get(path)
			return &v
		},
	}
}

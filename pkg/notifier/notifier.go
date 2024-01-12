package notifier

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/parser"
	"strconv"
)

type Notifier interface {
	Inform(result *parser.ParseResult, err error, isArray bool) error
	SetConfig(cfg *config.NotifierConfig)
}

func shouldInform(expression string, result builder.Jsonable, force bool) (bool, error) {
	if force || expression == "" {
		return true, nil
	}

	out, err := parser.ProcessExpression(expression, result, nil)
	if err != nil {
		return false, err
	}

	value, err := strconv.ParseBool(fmt.Sprintf("%v", out))
	if err != nil {
		return false, err
	}

	return value, nil
}

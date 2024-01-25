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
}

func ShouldInform(cfg *config.NotifierConfig, result builder.Jsonable) (bool, error) {
	if cfg.Force || cfg.Expression == "" {
		return true, nil
	}

	out, err := parser.ProcessExpression(cfg.Expression, result, nil)
	if err != nil {
		return false, err
	}

	value, err := strconv.ParseBool(fmt.Sprintf("%v", out))
	if err != nil {
		return false, err
	}

	return value, nil
}

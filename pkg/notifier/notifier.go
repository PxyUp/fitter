package notifier

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/antonmedv/expr"
	"strconv"
)

type Notifier interface {
	Inform(result *parser.ParseResult, err error, isArray bool) error
	SetConfig(cfg *config.NotifierConfig)
}

const (
	fitterResultRef = "fRes"
)

var (
	defEnv = map[string]interface{}{}
)

func extendEnv(env map[string]interface{}, result *parser.ParseResult) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	kv[fitterResultRef] = result.Raw()

	return kv
}

func shouldInform(expression string, result *parser.ParseResult, force bool) (bool, error) {
	if force || expression == "" {
		return true, nil
	}

	env := extendEnv(defEnv, result)

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return false, err
	}

	out, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	value, err := strconv.ParseBool(fmt.Sprintf("%v", out))
	if err != nil {
		return false, err
	}

	return value, nil
}

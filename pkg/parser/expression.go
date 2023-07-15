package parser

import (
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/antonmedv/expr"
)

const (
	fitterResultRef = "fRes"
)

var (
	defEnv = map[string]interface{}{}
)

func extendEnv(env map[string]interface{}, result builder.Jsonable) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	kv[fitterResultRef] = result.Raw()

	return kv
}

func ProcessExpression(expression string, result builder.Jsonable) (interface{}, error) {
	env := extendEnv(defEnv, result)

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return false, err
	}

	out, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return out, nil
}

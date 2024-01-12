package parser

import (
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/antonmedv/expr"
)

const (
	fitterResultRef = "fRes"
	fitterIndexRef  = "fIndex"
)

var (
	defEnv = map[string]interface{}{}
)

func extendEnv(env map[string]interface{}, result builder.Jsonable, index *uint32) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	kv[fitterResultRef] = result.Raw()
	if index != nil {
		kv[fitterIndexRef] = *index
	}

	return kv
}

func ProcessExpression(expression string, result builder.Jsonable, index *uint32) (interface{}, error) {
	env := extendEnv(defEnv, result, index)

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

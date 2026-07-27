package utils

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/expr-lang/expr"
	"strings"
)

const (
	fitterResultJsonRef           = "fResJson"
	fitterResultRef               = "fRes"
	fitterSourceRef               = "fSrc"
	fitterIndexRef                = "fIndex"
	fitterResultRaw               = "fResRaw"
	fitterNewLinePlaceholderKey   = "FNewLine"
	fitterNewLinePlaceholderValue = "$__FLINE__$"
)

var (
	defEnv = map[string]interface{}{
		fitterNewLinePlaceholderKey: fitterNewLinePlaceholderValue,
		"FNull":                     builder.NullValue,
		"FNil":                      nil,
		"isNull": func(value interface{}) bool {
			return builder.NullValue == value
		},
	}
)

func extendEnv(env map[string]interface{}, result builder.Interfacable, index *uint32) map[string]interface{} {
	kv := make(map[string]interface{})

	for k, v := range env {
		kv[k] = v
	}

	if result != nil {
		kv[fitterResultRaw] = result.Raw()
		kv[fitterResultRef] = result.ToInterface()
		kv[fitterResultJsonRef] = result.ToJson()
	}
	if index != nil {
		kv[fitterIndexRef] = *index
	}

	return kv
}

// ValidateExpression compiles the expression without running it, catching
// syntax errors at validation time. Variables stay dynamically typed (their
// real types depend on the parsed data), and expressions containing
// placeholder syntax are skipped: they are formatted at runtime before
// compilation, so they cannot be checked statically
func ValidateExpression(expression string) error {
	if expression == "" || strings.Contains(expression, "{") {
		return nil
	}

	_, err := expr.Compile(expression, expr.AllowUndefinedVariables())
	return err
}

// ProcessCondition reports whether the expression resolved to boolean true
func ProcessCondition(expression string, result builder.Interfacable, index *uint32, input builder.Interfacable) (bool, error) {
	return ProcessConditionWithSource(expression, result, nil, index, input)
}

// ProcessConditionWithSource is ProcessCondition with the source node the
// value was resolved from additionally exposed as fSrc
func ProcessConditionWithSource(expression string, result builder.Interfacable, source builder.Interfacable, index *uint32, input builder.Interfacable) (bool, error) {
	env := extendEnv(defEnv, result, index)
	if source != nil {
		env[fitterSourceRef] = source.ToInterface()
	}

	out, err := processExpression(env, expression, result, index, input)
	if err != nil {
		return false, err
	}

	return out.ToInterface() == true, nil
}

func ProcessExpression(expression string, result builder.Interfacable, index *uint32, input builder.Interfacable) (builder.Interfacable, error) {
	return processExpression(extendEnv(defEnv, result, index), expression, result, index, input)
}

func processExpression(env map[string]interface{}, expression string, result builder.Interfacable, index *uint32, input builder.Interfacable) (builder.Interfacable, error) {
	program, err := expr.Compile(Format(expression, result, index, input), expr.Env(env))
	if err != nil {
		return nil, err
	}

	out, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}

	bb, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	return builder.ToJsonable(bb), nil
}

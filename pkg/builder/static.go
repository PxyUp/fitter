package builder

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/config"
	"strconv"
)

type static struct {
	fieldType   config.FieldType
	stringValue string
}

func (s *static) ToInterface() interface{} {
	switch s.fieldType {
	case config.Null:
		return NullValue.ToInterface()
	case config.RawString:
		return String(s.stringValue, false).ToInterface()
	case config.String:
		return String(s.stringValue).ToInterface()
	case config.Bool:
		boolValue, err := strconv.ParseBool(s.stringValue)
		if err != nil {
			return NullValue.ToInterface()
		}
		return Bool(boolValue).ToInterface()
	case config.Float, config.Float64, config.Int, config.Int64:
		float32Value, err := strconv.ParseFloat(s.stringValue, 64)
		if err != nil {
			return NullValue.ToInterface()
		}
		return Number(float32Value).ToInterface()
	case config.Object, config.Array:
		return ToJsonableFromString(s.stringValue).ToInterface()
	}

	return NullValue.ToInterface()
}

var (
	_ Interfacable = &static{}
)

func Static(cfg *config.StaticGeneratedFieldConfig) *static {
	return &static{
		fieldType:   cfg.Type,
		stringValue: cfg.Value,
	}
}

func (s *static) IsEmpty() bool {
	return false
}

func (s *static) ToJson() string {
	switch s.fieldType {
	case config.Null:
		return NullValue.ToJson()
	case config.RawString:
		return String(s.stringValue, false).ToJson()
	case config.String:
		return String(s.stringValue).ToJson()
	case config.Bool:
		boolValue, err := strconv.ParseBool(s.stringValue)
		if err != nil {
			return NullValue.ToJson()
		}
		return Bool(boolValue).ToJson()
	case config.Float, config.Int, config.Float64, config.Int64:
		float32Value, err := strconv.ParseFloat(s.stringValue, 64)
		if err != nil {
			return NullValue.ToJson()
		}
		return Number(float32Value).ToJson()
	case config.Object, config.Array:
		return ToJsonableFromString(s.stringValue).ToJson()
	}

	return NullValue.ToJson()
}

func (s *static) Raw() json.RawMessage {
	switch s.fieldType {
	case config.Null:
		return NullValue.Raw()
	case config.RawString:
		return String(s.stringValue, false).Raw()
	case config.String:
		return String(s.stringValue).Raw()
	case config.Bool:
		boolValue, err := strconv.ParseBool(s.stringValue)
		if err != nil {
			return NullValue.Raw()
		}
		return Bool(boolValue).Raw()
	case config.Float, config.Float64, config.Int, config.Int64:
		float32Value, err := strconv.ParseFloat(s.stringValue, 64)
		if err != nil {
			return NullValue.Raw()
		}
		return Number(float32Value).Raw()
	case config.Object, config.Array:
		return ToJsonableFromString(s.stringValue).Raw()
	}

	return NullValue.Raw()
}

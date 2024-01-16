package builder

import (
	"github.com/PxyUp/fitter/pkg/config"
	"strconv"
)

type static struct {
	fieldType   config.FieldType
	stringValue string
}

var (
	_ Jsonable = &static{}
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
		return Null().ToJson()
	case config.RawString:
		return String(s.stringValue, false).ToJson()
	case config.String:
		return String(s.stringValue).ToJson()
	case config.Bool:
		boolValue, err := strconv.ParseBool(s.stringValue)
		if err != nil {
			return Null().ToJson()
		}
		return Bool(boolValue).ToJson()
	case config.Float:
		float32Value, err := strconv.ParseFloat(s.stringValue, 32)
		if err != nil {
			return Null().ToJson()
		}
		return Float(float32(float32Value)).ToJson()
	case config.Float64:
		float64Value, err := strconv.ParseFloat(s.stringValue, 64)
		if err != nil {
			return Null().ToJson()
		}
		return Float64(float64Value).ToJson()
	case config.Int:
		int32Value, err := strconv.ParseInt(s.stringValue, 10, 32)
		if err != nil {
			return Null().ToJson()
		}
		return Int(int(int32Value)).ToJson()
	case config.Int64:
		int64Value, err := strconv.ParseInt(s.stringValue, 10, 64)
		if err != nil {
			return Null().ToJson()
		}
		return Int64(int64Value).ToJson()
	}

	return Null().ToJson()
}

func (s *static) Raw() interface{} {
	switch s.fieldType {
	case config.Null:
		return Null().Raw()
	case config.RawString:
		return String(s.stringValue, false).Raw()
	case config.String:
		return String(s.stringValue).Raw()
	case config.Bool:
		boolValue, err := strconv.ParseBool(s.stringValue)
		if err != nil {
			return Null().Raw()
		}
		return Bool(boolValue).Raw()
	case config.Float:
		float32Value, err := strconv.ParseFloat(s.stringValue, 32)
		if err != nil {
			return Null().Raw()
		}
		return Float(float32(float32Value)).Raw()
	case config.Float64:
		float64Value, err := strconv.ParseFloat(s.stringValue, 64)
		if err != nil {
			return Null().Raw()
		}
		return Float64(float64Value).Raw()
	case config.Int:
		int32Value, err := strconv.ParseInt(s.stringValue, 10, 32)
		if err != nil {
			return Null().Raw()
		}
		return Int(int(int32Value)).Raw()
	case config.Int64:
		int64Value, err := strconv.ParseInt(s.stringValue, 10, 64)
		if err != nil {
			return Null().ToJson()
		}
		return Int64(int64Value).Raw()
	}

	return Null().Raw()
}

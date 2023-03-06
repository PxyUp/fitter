package builder

import (
	"github.com/PxyUp/fitter/pkg/config"
	"strconv"
)

type static struct {
	fieldType   config.FieldType
	stringValue string
}

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
	case config.Int:
		int32Value, err := strconv.ParseInt(s.stringValue, 10, 32)
		if err != nil {
			return Null().ToJson()
		}
		return Int(int(int32Value)).ToJson()
	}

	return Null().ToJson()
}

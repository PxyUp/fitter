package builder

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/config"
	"strconv"
)

type static struct {
	value Interfacable
}

func (s *static) ToInterface() interface{} {
	return s.value.ToInterface()
}

var (
	_ Interfacable = &static{}
)

type StaticCfg struct {
	Type  config.FieldType `yaml:"type" json:"type"`
	Value string           `json:"value" yaml:"value"`
}

func Static(cfg *StaticCfg) *static {
	switch cfg.Type {
	case config.Null:
		return &static{
			value: NullValue,
		}
	case config.RawString:
		return &static{
			value: String(cfg.Value, false),
		}
	case config.String:
		return &static{
			value: String(cfg.Value, false),
		}
	case config.Bool:
		boolValue, err := strconv.ParseBool(cfg.Value)
		if err != nil {
			return &static{
				value: NullValue,
			}
		}
		return &static{
			value: Bool(boolValue),
		}
	case config.Float, config.Int, config.Float64, config.Int64:
		float32Value, err := strconv.ParseFloat(cfg.Value, 64)
		if err != nil {
			return &static{
				value: NullValue,
			}
		}
		return &static{
			value: Number(float32Value),
		}
	case config.Object, config.Array:
		return &static{
			value: ToJsonableFromString(cfg.Value),
		}
	}

	return &static{
		value: NullValue,
	}
}

func (s *static) IsEmpty() bool {
	return s.value.IsEmpty()
}

func (s *static) ToJson() string {
	return s.value.ToJson()
}

func (s *static) Raw() json.RawMessage {
	return s.value.Raw()
}

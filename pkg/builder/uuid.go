package builder

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/google/uuid"
	"regexp"
)

func UUID(cfg *config.UUIDGeneratedFieldConfig) Interfacable {
	uuidStr := uuid.New().String()
	if cfg.Regexp != "" {
		re, err := regexp.Compile(cfg.Regexp)
		if err != nil {
			return Null()
		}
		return String(re.FindString(uuidStr))
	}
	return String(uuidStr)
}

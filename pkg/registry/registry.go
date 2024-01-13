package registry

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/processor"
)

var (
	errItemNotExist = errors.New("item with this name not exist")
)

type Registry interface {
	Get(string) processor.Processor
}

type localRegistry struct {
	kv     map[string]processor.Processor
	logger logger.Logger
}

func NewFromConfig(config *config.Config, logger logger.Logger) *localRegistry {
	if config != nil {
		limitter.SetLimits(config.Limits)
	}

	kv := make(map[string]processor.Processor)
	if config != nil {
		for _, item := range config.Items {
			kv[item.Name] = processor.CreateProcessor(item, logger)
		}
	}
	return &localRegistry{
		kv:     kv,
		logger: logger,
	}
}

func FromItem(itemCfg *config.CliItem, logger logger.Logger) *localRegistry {
	limitter.SetLimits(itemCfg.Limits)

	return &localRegistry{
		logger: logger,
		kv: map[string]processor.Processor{
			itemCfg.Item.Name: processor.CreateProcessor(itemCfg.Item, logger),
		},
	}
}

func (r *localRegistry) Get(name string) processor.Processor {
	value, ok := r.kv[name]
	if !ok {
		return processor.Null(errItemNotExist)
	}

	r.logger.Infof("got processor for %s", name)

	return value
}

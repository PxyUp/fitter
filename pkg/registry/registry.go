package registry

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
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

func NewFromConfig(config *config.Config) *localRegistry {
	connectors.SetRequestPerHost(config.HostRequestLimiter)
	
	kv := make(map[string]processor.Processor)
	if config != nil {
		for _, item := range config.Items {
			kv[item.Name] = processor.CreateProcessor(item)
		}
	}
	return &localRegistry{
		kv:     kv,
		logger: logger.Null,
	}
}

func (r *localRegistry) WithLogger(logger logger.Logger) *localRegistry {
	r.logger = logger
	return r
}

func FromItem(itemCfg *config.CliItem) *localRegistry {
	connectors.SetRequestPerHost(itemCfg.HostRequestLimiter)

	return &localRegistry{
		logger: logger.Null,
		kv: map[string]processor.Processor{
			itemCfg.Item.Name: processor.CreateProcessor(itemCfg.Item),
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

package registry

import (
	"errors"
	"github.com/PxyUp/fitter/pkg/config"
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

func FromItem(item *config.Item) *localRegistry {
	return &localRegistry{
		kv: map[string]processor.Processor{
			item.Name: processor.CreateProcessor(item),
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

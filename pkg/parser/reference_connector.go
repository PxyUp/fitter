package parser

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"sync"
)

var (
	once         = sync.Once{}
	refStoreImpl = &refStore{}
)

type refStore struct {
	kv    map[string]builder.Jsonable
	mutex sync.Mutex
}

func (s *refStore) Get(name string) connectors.Connector {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	value, ok := s.kv[name]
	if !ok {
		return nil
	}

	return connectors.NewStatic(&config.StaticConnectorConfig{
		Value: value.ToJson(),
	})
}

func SetReference(references map[string]*config.ModelField, logger logger.Logger) {
	once.Do(func() {
		for k, v := range references {
			res, err := NewEngine(v.ConnectorConfig, logger).Get(v.Model, nil, nil)
			if err != nil {
				refStoreImpl.kv[k] = builder.Null()
				continue
			}
			refStoreImpl.mutex.Lock()
			refStoreImpl.kv[k] = res
			refStoreImpl.mutex.Unlock()
		}
	})
}

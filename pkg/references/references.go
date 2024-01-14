package references

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"sync"
)

var (
	once         = &sync.Once{}
	refStoreImpl = &refStore{
		kv: make(map[string]builder.Jsonable),
	}
)

type refStore struct {
	kv    map[string]builder.Jsonable
	mutex sync.Mutex
}

func Get(name string) builder.Jsonable {
	refStoreImpl.mutex.Lock()
	value, ok := refStoreImpl.kv[name]
	refStoreImpl.mutex.Unlock()

	if !ok {
		return builder.Null()
	}

	return value
}

func SetReference(references map[string]*config.ModelField, cb func(name string, model *config.ModelField) (builder.Jsonable, error)) {
	once.Do(func() {
		for k, v := range references {
			res, err := cb(k, v)
			if err != nil {
				refStoreImpl.mutex.Lock()
				refStoreImpl.kv[k] = builder.Null()
				refStoreImpl.mutex.Unlock()
				continue
			}
			refStoreImpl.mutex.Lock()
			refStoreImpl.kv[k] = res
			refStoreImpl.mutex.Unlock()
		}
	})
}

package references

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"sync"
	"time"
)

var (
	once         = &sync.Once{}
	refStoreImpl = &refStore{
		kv: make(map[string]*refRecord),
	}
)

type refRecord struct {
	fetcher    refFetcher
	cfg        *config.Reference
	expireTime *time.Time

	value builder.Jsonable
}

type refStore struct {
	kv    map[string]*refRecord
	mutex sync.Mutex
}

func getValueFromFetcher(fetcher refFetcher, name string, cfg *config.ModelField) builder.Jsonable {
	res, err := fetcher(name, cfg)
	if err != nil {
		return builder.NullValue
	}

	return res
}

func getExpireTime(expire *uint32) *time.Time {
	if expire == nil {
		return nil
	}

	if *expire > 0 {
		expireTime := time.Now().Add(time.Second * time.Duration(*expire))
		return &expireTime
	}

	zero := time.Time{}
	return &zero
}

func Get(name string) builder.Jsonable {
	defer refStoreImpl.mutex.Unlock()
	refStoreImpl.mutex.Lock()

	record, ok := refStoreImpl.kv[name]
	if !ok {
		return builder.NullValue
	}

	if record.expireTime != nil && time.Now().After(*record.expireTime) {
		record.value = getValueFromFetcher(record.fetcher, name, record.cfg.ModelField)
		record.expireTime = getExpireTime(record.cfg.Expire)
	}

	return record.value
}

func createRecord(name string, cfg *config.Reference, fetcher refFetcher) *refRecord {
	record := &refRecord{
		cfg:     cfg,
		fetcher: fetcher,
	}

	record.value = getValueFromFetcher(fetcher, name, record.cfg.ModelField)
	record.expireTime = getExpireTime(record.cfg.Expire)
	return record
}

type refFetcher func(name string, model *config.ModelField) (builder.Jsonable, error)

func SetReference(references config.RefMap, cb refFetcher) {
	once.Do(func() {
		for k, v := range references {
			lk := k
			lv := v
			refStoreImpl.mutex.Lock()
			refStoreImpl.kv[k] = createRecord(lk, lv, cb)
			refStoreImpl.mutex.Unlock()
		}
	})
}

// ReplaceReferences swaps the reference store, bypassing the process-lifetime
// once semantics of SetReference. Long-lived embedders that execute many
// unrelated configs in a single process (e.g. the WASM playground, which
// never restarts the Go runtime between runs) call it before each run so
// every config gets exactly its own references instead of only ever seeing
// whichever config happened to run first. Not used by the native CLI/service/
// MCP binaries, where SetReference's once-per-process semantics are intentional.
func ReplaceReferences(references config.RefMap, cb refFetcher) {
	refStoreImpl.mutex.Lock()
	defer refStoreImpl.mutex.Unlock()

	refStoreImpl.kv = make(map[string]*refRecord, len(references))
	for k, v := range references {
		refStoreImpl.kv[k] = createRecord(k, v, cb)
	}
}

package limitter

import (
	"github.com/PxyUp/fitter/pkg/config"
	"golang.org/x/sync/semaphore"
)

var (
	limitPerHost     = make(map[string]*semaphore.Weighted)
	chromiumInstance *semaphore.Weighted
)

func setChromiumInstance(count uint32) {
	if count <= 0 {
		return
	}
	chromiumInstance = semaphore.NewWeighted(int64(count))
}

func setRequestPerHost(limits config.HostRequestLimiter) {
	for k, v := range limits {
		if _, ok := limitPerHost[k]; !ok {
			limitPerHost[k] = semaphore.NewWeighted(v)
		}
	}
}

func SetLimits(limits *config.Limits) {
	if limits == nil {
		return
	}
	setRequestPerHost(limits.HostRequestLimiter)
	setChromiumInstance(limits.ChromiumInstance)
}

func HostLimiter(host string) *semaphore.Weighted {
	if hostLimit, ok := limitPerHost[host]; ok {
		return hostLimit
	}

	return nil
}

func ChromiumLimiter() *semaphore.Weighted {
	return chromiumInstance
}

package limitter

import (
	"github.com/PxyUp/fitter/pkg/config"
	"golang.org/x/sync/semaphore"
	"sync"
)

var (
	limitPerHost       = make(map[string]*semaphore.Weighted)
	chromiumInstance   *semaphore.Weighted
	dockerContainers   *semaphore.Weighted
	playwrightInstance *semaphore.Weighted

	once = &sync.Once{}
)

func setSemaphoreLimit(sem **semaphore.Weighted, count uint32) {
	if count <= 0 {
		return
	}
	*sem = semaphore.NewWeighted(int64(count))
}

func setRequestPerHost(limits config.HostRequestLimiter) {
	for k, v := range limits {
		if _, ok := limitPerHost[k]; !ok {
			limitPerHost[k] = semaphore.NewWeighted(v)
		}
	}
}

func SetLimits(limits *config.Limits) {
	once.Do(func() {
		if limits == nil {
			return
		}
		setSemaphoreLimit(&chromiumInstance, limits.ChromiumInstance)
		setSemaphoreLimit(&dockerContainers, limits.DockerContainers)
		setSemaphoreLimit(&playwrightInstance, limits.PlaywrightInstance)
		setRequestPerHost(limits.HostRequestLimiter)
	})
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

func PlaywrightLimiter() *semaphore.Weighted {
	return playwrightInstance
}

func DockerLimiter() *semaphore.Weighted {
	return dockerContainers
}

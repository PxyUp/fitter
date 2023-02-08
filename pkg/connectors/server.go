package connectors

import (
	"context"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	timeout                 = 10 * time.Second
	defaultConcurrentWorker = 20
)

type apiConnector struct {
	headers map[string]string
	url     string

	method string
	logger logger.Logger
}

var (
	DefaultClient = http.Client{
		Timeout: timeout,
	}

	sem *semaphore.Weighted
)

func init() {
	defaultConcurrentRequest := defaultConcurrentWorker
	if value, ok := os.LookupEnv("FITTER_HTTP_WORKER"); ok {
		intValue, err := strconv.ParseInt(value, 10, 32)
		if err == nil && intValue > 0 {
			defaultConcurrentRequest = int(intValue)
		}
	}
	sem = semaphore.NewWeighted(int64(defaultConcurrentRequest))
}

func NewAPI(cfg *config.ServerConnectorConfig) *apiConnector {
	return &apiConnector{
		headers: cfg.Headers,
		url:     cfg.Url,
		method:  cfg.Method,
		logger:  logger.Null,
	}
}

func (api *apiConnector) WithLogger(logger logger.Logger) *apiConnector {
	api.logger = logger
	return api
}

func (api *apiConnector) Get() ([]byte, error) {
	req, err := http.NewRequest(api.method, api.url, nil)
	if err != nil {
		api.logger.Errorw("unable to create http request", "error", err.Error())
		return nil, err
	}

	for k, v := range api.headers {
		req.Header.Add(k, v)
	}

	err = sem.Acquire(context.Background(), 1)
	if err != nil {
		api.logger.Errorw("unable to acquire semaphore", "method", api.method, "url", api.url, "error", err.Error())
		return nil, err
	}

	defer sem.Release(1)

	resp, err := DefaultClient.Do(req)
	if err != nil {
		api.logger.Errorw("unable to send http request", "method", api.method, "url", api.url, "error", err.Error())
		return nil, err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		api.logger.Errorw("unable to read http response", "error", err.Error())
		return nil, err
	}

	return bytes, nil
}

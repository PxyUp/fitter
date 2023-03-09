package connectors

import (
	"bytes"
	"context"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	timeout                 = 60 * time.Second
	defaultConcurrentWorker = 1000
)

type apiConnector struct {
	url    string
	logger logger.Logger
	client *http.Client
	cfg    *config.ServerConnectorConfig
}

var (
	sem *semaphore.Weighted

	ctx = context.Background()
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

func NewAPI(url string, cfg *config.ServerConnectorConfig, client *http.Client) *apiConnector {
	return &apiConnector{
		client: client,
		url:    url,
		cfg:    cfg,
		logger: logger.Null,
	}
}

func (api *apiConnector) WithLogger(logger logger.Logger) *apiConnector {
	api.logger = logger
	return api
}

func (api *apiConnector) Get() ([]byte, error) {
	if api.url == "" {
		return nil, errEmpty
	}

	err := sem.Acquire(ctx, 1)
	if err != nil {
		api.logger.Errorw("unable to acquire semaphore", "method", api.cfg.Method, "url", api.url, "error", err.Error())
		return nil, err
	}

	defer sem.Release(1)

	req, err := http.NewRequest(api.cfg.Method, api.url, bytes.NewBufferString(api.cfg.Body))

	if err != nil {
		api.logger.Errorw("unable to create http request", "error", err.Error())
		return nil, err
	}

	for k, v := range api.cfg.Headers {
		req.Header.Add(k, v)
	}

	client := http.DefaultClient
	if api.client != nil {
		client = api.client
	}

	if hostLimit := limitter.HostLimiter(req.Host); hostLimit != nil {
		errHostLimit := hostLimit.Acquire(ctx, 1)
		if errHostLimit != nil {
			api.logger.Errorw("unable to acquire host limit semaphore", "method", api.cfg.Method, "url", api.url, "error", err.Error(), "host", req.Host)
			return nil, errHostLimit
		}
		defer hostLimit.Release(1)
	}

	tt := timeout
	if api.cfg.Timeout > 0 {
		tt = time.Duration(api.cfg.Timeout) * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, tt)
	defer cancel()

	api.logger.Infof("send request to url: %s", api.url)
	resp, err := client.Do(req.WithContext(reqCtx))
	if err != nil {
		api.logger.Errorw("unable to send http request", "method", api.cfg.Method, "url", api.url, "error", err.Error())
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

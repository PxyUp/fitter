package connectors

import (
	"github.com/PxyUp/fitter/pkg/logger"
	"io/ioutil"
	"net/http"
	"time"
)

type apiConnector struct {
	headers map[string]string
	url     string
	client  *http.Client

	method string
	logger logger.Logger
}

func NewAPI(headers map[string]string, url string, method string, client *http.Client) *apiConnector {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &apiConnector{
		headers: headers,
		url:     url,
		client:  client,
		method:  method,
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

	resp, err := api.client.Do(req)
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

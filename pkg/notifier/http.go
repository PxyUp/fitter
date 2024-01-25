package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"net/http"
	"time"
)

type httpNotifier struct {
	logger logger.Logger
	name   string
	cfg    *config.HttpConfig
}

type HttpRequestBody struct {
	Error bool             `json:"error,omitempty"`
	Value *json.RawMessage `json:"result,omitempty"`
}

func buildBody(result *parser.ParseResult, err error) *HttpRequestBody {
	rr := &HttpRequestBody{}
	if err != nil {
		rr.Error = true
		return rr
	}
	val := json.RawMessage{}
	json.Unmarshal([]byte(result.ToJson()), &val)
	rr.Value = &val

	return rr
}

func (h *httpNotifier) Inform(result *parser.ParseResult, err error, isArray bool) error {
	rr := buildBody(result, err)
	bb, err := json.Marshal(rr)
	if err != nil {
		h.logger.Errorw("cant unmarshal request body", "error", err.Error())
		return err
	}
	req, err := http.NewRequest(h.cfg.Method, h.cfg.Url, bytes.NewReader(bb))
	if err != nil {
		h.logger.Errorw("cant create request", "error", err.Error())
		return err
	}

	for k, v := range h.cfg.Headers {
		req.Header.Add(k, v)
	}

	if h.cfg.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.cfg.Timeout)*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := http_client.GetDefaultClient().Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		h.logger.Errorw("cant inform about results request", "error", err.Error())
		return err
	}

	return nil
}

func (h *httpNotifier) WithLogger(logger logger.Logger) *httpNotifier {
	h.logger = logger
	return h
}

var (
	_ Notifier = &httpNotifier{}
)

func NewHttpNotifier(name string, cfg *config.HttpConfig) *httpNotifier {
	return &httpNotifier{
		logger: logger.Null,
		name:   name,
		cfg:    cfg,
	}
}

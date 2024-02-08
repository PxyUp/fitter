package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/utils"
	"net/http"
	"time"
)

type httpNotifier struct {
	logger logger.Logger
	name   string
	cfg    *config.HttpConfig
}

func (h *httpNotifier) notify(record *singleRecord, input builder.Interfacable) error {
	bb, err := json.Marshal(record)
	if err != nil {
		h.logger.Errorw("cant unmarshal request body", "error", err.Error())
		return err
	}

	var url string
	if record.Error != nil {
		url = utils.Format(h.cfg.Url, builder.String((*record.Error).Error()), record.Index, input)
	} else {
		url = utils.Format(h.cfg.Url, builder.ToJsonable(record.Body), record.Index, input)
	}

	req, err := http.NewRequest(h.cfg.Method, url, bytes.NewReader(bb))
	if err != nil {
		h.logger.Errorw("cant create request", "error", err.Error())
		return err
	}

	for k, v := range h.cfg.Headers {
		var headerValue string
		if record.Error != nil {
			headerValue = utils.Format(v, builder.String((*record.Error).Error()), record.Index, input)
		} else {
			headerValue = utils.Format(v, builder.ToJsonable(record.Body), record.Index, input)
		}
		req.Header.Add(k, headerValue)
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

func (o *httpNotifier) GetLogger() logger.Logger {
	return o.logger
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

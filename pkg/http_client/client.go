package http_client

import (
	"net/http"
	"time"
)

var (
	Client = &http.Client{
		Timeout: time.Minute * 2,
	}
)

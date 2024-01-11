package http_client

import (
	"net/http"
	"time"
)

func GetDefaultClient() *http.Client {
	return &http.Client{
		Timeout: time.Minute * 2,
	}
}

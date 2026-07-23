package connectors_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiConnectorGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, err := connectors.NewAPI(srv.URL, &config.ServerConnectorConfig{Method: http.MethodGet}, nil).Get(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, string(body))
}

func TestApiConnectorGetCancelled(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := connectors.NewAPI(srv.URL, &config.ServerConnectorConfig{Method: http.MethodGet}, nil).Get(ctx, nil, nil, nil)
		done <- err
	}()

	cancel()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("connector did not abort after context cancellation")
	}
}

package connectors_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/oauthflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
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

func TestApiConnectorOAuth2ClientCredentials(t *testing.T) {
	var tokenCalls int32
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "id", r.FormValue("client_id"))
		assert.Equal(t, "secret", r.FormValue("client_secret"))
		assert.Equal(t, "read write", r.FormValue("scope"))
		assert.Equal(t, "acme", r.FormValue("audience"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-1","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok-1", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	api := connectors.NewAPI(apiSrv.URL, &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{
			TokenUrl:       tokenSrv.URL,
			ClientId:       "id",
			ClientSecret:   "secret",
			Scopes:         []string{"read", "write"},
			EndpointParams: map[string]string{"audience": "acme"},
			AuthStyle:      "params",
		},
	}, nil)

	for i := 0; i < 2; i++ {
		body, err := api.Get(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, `{"ok":true}`, string(body))
	}
	assert.EqualValues(t, 1, atomic.LoadInt32(&tokenCalls), "token should be cached between requests")
}

func TestApiConnectorOAuth2RefreshToken(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "refresh_token", r.FormValue("grant_type"))
		assert.Equal(t, "refresh-1", r.FormValue("refresh_token"))
		user, pass, ok := r.BasicAuth()
		require.True(t, ok)
		assert.Equal(t, "id", user)
		assert.Equal(t, "secret", pass)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-r","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok-r", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	body, err := connectors.NewAPI(apiSrv.URL, &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{
			TokenUrl:     tokenSrv.URL,
			GrantType:    config.RefreshTokenGrant,
			ClientId:     "id",
			ClientSecret: "secret",
			RefreshToken: "refresh-1",
			AuthStyle:    "header",
		},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, string(body))
}

func TestApiConnectorOAuth2RetryOn401(t *testing.T) {
	var tokenCalls int32
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&tokenCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":"tok-%d","token_type":"Bearer","expires_in":3600}`, n)
	}))
	defer tokenSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok-2" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	body, err := connectors.NewAPI(apiSrv.URL, &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{
			TokenUrl:  tokenSrv.URL,
			ClientId:  "id",
			AuthStyle: "params",
		},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, string(body))
	assert.EqualValues(t, 2, atomic.LoadInt32(&tokenCalls), "401 should force exactly one fresh token")
}

func TestApiConnectorOAuth2InvalidConfig(t *testing.T) {
	_, err := connectors.NewAPI("http://localhost", &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.ErrorContains(t, err, "token_url is required")

	_, err = connectors.NewAPI("http://localhost", &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{TokenUrl: "http://localhost/token", GrantType: config.RefreshTokenGrant},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.ErrorContains(t, err, "refresh_token or token_file is required")

	_, err = connectors.NewAPI("http://localhost", &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{TokenUrl: "http://localhost/token", GrantType: "password"},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.ErrorContains(t, err, "unsupported grant_type")
}

func TestApiConnectorOAuth2TokenFileValidAccessToken(t *testing.T) {
	var tokenCalls int32
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tokenSrv.Close()

	tokenFile := filepath.Join(t.TempDir(), "token.json")
	require.NoError(t, oauthflow.SaveTokenFile(tokenFile, &oauth2.Token{
		AccessToken: "stored-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}))

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer stored-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	body, err := connectors.NewAPI(apiSrv.URL, &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{
			TokenUrl:  tokenSrv.URL,
			GrantType: config.RefreshTokenGrant,
			ClientId:  "id",
			AuthStyle: "params",
			TokenFile: tokenFile,
		},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, string(body))
	assert.EqualValues(t, 0, atomic.LoadInt32(&tokenCalls), "valid stored access token should avoid the token endpoint")
}

func TestApiConnectorOAuth2TokenFileRotationPersisted(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		// single-use rotation: only the token persisted from the seed exchange works
		assert.Equal(t, "file-refresh-1", r.FormValue("refresh_token"), "must use the refresh token from the file, not the config seed")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"acc-2","refresh_token":"file-refresh-2","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	tokenFile := filepath.Join(t.TempDir(), "token.json")
	require.NoError(t, oauthflow.SaveTokenFile(tokenFile, &oauth2.Token{
		RefreshToken: "file-refresh-1",
	}))

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer acc-2", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	body, err := connectors.NewAPI(apiSrv.URL, &config.ServerConnectorConfig{
		Method: http.MethodGet,
		OAuth2: &config.OAuth2Config{
			TokenUrl:     tokenSrv.URL,
			GrantType:    config.RefreshTokenGrant,
			ClientId:     "id",
			RefreshToken: "dead-config-seed",
			AuthStyle:    "params",
			TokenFile:    tokenFile,
		},
	}, nil).Get(context.Background(), nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, string(body))

	persisted, err := oauthflow.LoadTokenFile(tokenFile)
	require.NoError(t, err)
	assert.Equal(t, "file-refresh-2", persisted.RefreshToken, "rotated refresh token should be written back")
	assert.Equal(t, "acc-2", persisted.AccessToken)
}

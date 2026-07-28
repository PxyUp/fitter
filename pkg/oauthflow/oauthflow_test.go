package oauthflow_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/PxyUp/fitter/pkg/oauthflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestTokenFileRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "token.json")

	_, err := oauthflow.LoadTokenFile(path)
	assert.ErrorIs(t, err, os.ErrNotExist)

	token := &oauth2.Token{
		AccessToken:  "acc",
		RefreshToken: "ref",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour).Round(time.Second),
	}
	require.NoError(t, oauthflow.SaveTokenFile(path, token))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	loaded, err := oauthflow.LoadTokenFile(path)
	require.NoError(t, err)
	assert.Equal(t, "acc", loaded.AccessToken)
	assert.Equal(t, "ref", loaded.RefreshToken)
	assert.True(t, loaded.Valid())
}

func TestBrowserLogin(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
		assert.Equal(t, "test-code", r.FormValue("code"))
		assert.NotEmpty(t, r.FormValue("code_verifier"), "PKCE verifier should be sent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"acc","refresh_token":"ref","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	cfg := &oauthflow.LoginConfig{
		ClientID:  "id",
		AuthURL:   "https://auth.example.com/authorize",
		TokenURL:  tokenSrv.URL,
		AuthStyle: "params",
		Port:      18988,
		OpenBrowser: func(authURL string) error {
			go func() {
				parsed, errParse := url.Parse(authURL)
				if errParse != nil {
					return
				}
				query := parsed.Query()
				callback := query.Get("redirect_uri") + "?code=test-code&state=" + url.QueryEscape(query.Get("state"))
				resp, errGet := http.Get(callback)
				if errGet == nil {
					_ = resp.Body.Close()
				}
			}()
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := oauthflow.BrowserLogin(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, "acc", token.AccessToken)
	assert.Equal(t, "ref", token.RefreshToken)
}

func TestDeviceLogin(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dev-1","user_code":"ABCD-1234","verification_uri":"https://example.com/activate","interval":1,"expires_in":300}`))
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "urn:ietf:params:oauth:grant-type:device_code", r.FormValue("grant_type"))
		assert.Equal(t, "dev-1", r.FormValue("device_code"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"acc","refresh_token":"ref","token_type":"Bearer","expires_in":3600}`))
	})

	cfg := &oauthflow.LoginConfig{
		ClientID:      "id",
		TokenURL:      srv.URL + "/token",
		DeviceAuthURL: srv.URL + "/device",
		AuthStyle:     "params",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := oauthflow.DeviceLogin(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, "acc", token.AccessToken)
	assert.Equal(t, "ref", token.RefreshToken)
}

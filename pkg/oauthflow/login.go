package oauthflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	errNoDeviceAuthURL = errors.New("oauthflow: device auth url is required for the device flow")
	errNoAuthURL       = errors.New("oauthflow: auth url is required for the browser flow")
	errStateMismatch   = errors.New("oauthflow: state parameter mismatch")
)

const defaultCallbackPort = 8988

type LoginConfig struct {
	ClientID      string
	ClientSecret  string
	Scopes        []string
	AuthURL       string
	TokenURL      string
	DeviceAuthURL string
	// AuthStyle: "" auto detect, "header" or "params"
	AuthStyle string
	// Port for the browser flow local callback, default 8988
	Port int
	// ListenAddr is the browser flow bind address, default 127.0.0.1;
	// inside a container use 0.0.0.0 so the published port is reachable
	ListenAddr string
	// RedirectURL overrides the registered callback url when it differs from
	// the listen address (e.g. docker port mapping); default http://127.0.0.1:<port>/callback
	RedirectURL string
	// ExtraAuthParams are appended to the authorization request (e.g. access_type=offline for Google)
	ExtraAuthParams map[string]string
	// OpenBrowser opens the authorization url; nil = the url is only printed to Out
	OpenBrowser func(url string) error
	// Out receives user-facing instructions, default os.Stderr behaviour is up to the caller
	Out io.Writer
}

func (c *LoginConfig) printf(format string, args ...interface{}) {
	if c.Out != nil {
		_, _ = fmt.Fprintf(c.Out, format, args...)
	}
}

func (c *LoginConfig) oauthConfig(redirectURL string) *oauth2.Config {
	authStyle := oauth2.AuthStyleAutoDetect
	switch c.AuthStyle {
	case "header":
		authStyle = oauth2.AuthStyleInHeader
	case "params":
		authStyle = oauth2.AuthStyleInParams
	}

	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Scopes:       c.Scopes,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:       c.AuthURL,
			TokenURL:      c.TokenURL,
			DeviceAuthURL: c.DeviceAuthURL,
			AuthStyle:     authStyle,
		},
	}
}

// DeviceLogin runs the oauth2 device_code grant: prints the verification url +
// user code and polls the token endpoint until the user approves
func DeviceLogin(ctx context.Context, cfg *LoginConfig) (*oauth2.Token, error) {
	if cfg.DeviceAuthURL == "" {
		return nil, errNoDeviceAuthURL
	}

	oc := cfg.oauthConfig("")
	deviceAuth, err := oc.DeviceAuth(ctx)
	if err != nil {
		return nil, err
	}

	verificationURL := deviceAuth.VerificationURIComplete
	if verificationURL != "" {
		cfg.printf("Open %s in a browser (code %s is prefilled)\n", verificationURL, deviceAuth.UserCode)
	} else {
		verificationURL = deviceAuth.VerificationURI
		cfg.printf("Open %s in a browser and enter the code: %s\n", verificationURL, deviceAuth.UserCode)
	}
	if cfg.OpenBrowser != nil {
		_ = cfg.OpenBrowser(verificationURL)
	}
	cfg.printf("Waiting for approval...\n")

	return oc.DeviceAccessToken(ctx, deviceAuth)
}

// BrowserLogin runs the authorization_code grant with PKCE: starts a one-shot
// callback listener on 127.0.0.1, sends the user to the authorization url and
// exchanges the received code
func BrowserLogin(ctx context.Context, cfg *LoginConfig) (*oauth2.Token, error) {
	if cfg.AuthURL == "" {
		return nil, errNoAuthURL
	}

	port := cfg.Port
	if port == 0 {
		port = defaultCallbackPort
	}
	listenAddr := cfg.ListenAddr
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddr, port))
	if err != nil {
		return nil, err
	}

	redirectURL := cfg.RedirectURL
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	}
	oc := cfg.oauthConfig(redirectURL)

	state := uuid.NewString()
	verifier := oauth2.GenerateVerifier()

	opts := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(verifier)}
	for k, v := range cfg.ExtraAuthParams {
		opts = append(opts, oauth2.SetAuthURLParam(k, v))
	}
	authURL := oc.AuthCodeURL(state, opts...)

	type callback struct {
		code string
		err  error
	}
	resultChan := make(chan callback, 1)

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if errCode := query.Get("error"); errCode != "" {
			http.Error(w, "authorization failed: "+errCode, http.StatusBadRequest)
			resultChan <- callback{err: fmt.Errorf("oauthflow: authorization failed: %s (%s)", errCode, query.Get("error_description"))}
			return
		}
		if query.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			resultChan <- callback{err: errStateMismatch}
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<h1>Authorized</h1>You can close this tab and return to the terminal."))
		resultChan <- callback{code: query.Get("code")}
	})}

	go func() {
		errServe := server.Serve(listener)
		if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
			resultChan <- callback{err: errServe}
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	cfg.printf("Open the following url in a browser to authorize:\n%s\n", authURL)
	if cfg.OpenBrowser != nil {
		_ = cfg.OpenBrowser(authURL)
	}
	cfg.printf("Waiting for the callback on %s...\n", redirectURL)

	select {
	case res := <-resultChan:
		if res.err != nil {
			return nil, res.err
		}
		return oc.Exchange(ctx, res.code, oauth2.VerifierOption(verifier))
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// LoginURLHost extracts the host of the authorization url, for user-facing messages
func LoginURLHost(authURL string) string {
	parsed, err := url.Parse(authURL)
	if err != nil {
		return authURL
	}
	return parsed.Host
}

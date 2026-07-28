package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/PxyUp/fitter/pkg/oauthflow"
	"golang.org/x/oauth2"
)

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func knownProviders() string {
	names := make([]string, 0, len(oauthflow.Presets))
	for name := range oauthflow.Presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, "|")
}

func runAuth(args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	provider := fs.String("provider", "", "Provider preset: "+knownProviders())
	clientID := fs.String("client-id", "", "OAuth2 client id")
	clientSecret := fs.String("client-secret", "", "OAuth2 client secret (some device flows work without it)")
	scopes := fs.String("scopes", "", "Comma separated list of scopes")
	tokenFile := fs.String("token-file", "", "Path to store the received token, e.g. ~/.fitter/tokens/github.json")
	flow := fs.String("flow", "auto", "auto|device|browser")
	port := fs.Int("port", envIntOr("FITTER_AUTH_PORT", 8988), "Local callback port for the browser flow")
	listenAddr := fs.String("listen", envOr("FITTER_AUTH_LISTEN", ""), "Browser flow bind address, default 127.0.0.1; use 0.0.0.0 inside a container")
	redirectURL := fs.String("redirect-url", envOr("FITTER_AUTH_REDIRECT_URL", ""), "Callback url registered at the provider when it differs from the listen address (e.g. docker port mapping)")
	authURL := fs.String("auth-url", "", "Authorization endpoint (overrides provider preset)")
	tokenURL := fs.String("token-url", "", "Token endpoint (overrides provider preset)")
	deviceAuthURL := fs.String("device-auth-url", "", "Device authorization endpoint (overrides provider preset)")
	authStyle := fs.String("auth-style", "", "Token endpoint auth style: header|params (overrides provider preset)")
	noBrowser := fs.Bool("no-browser", false, "Only print the url instead of opening the browser")
	_ = fs.Parse(args)

	cfg := &oauthflow.LoginConfig{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Port:         *port,
		ListenAddr:   *listenAddr,
		RedirectURL:  *redirectURL,
		Out:          os.Stderr,
	}
	if !*noBrowser {
		cfg.OpenBrowser = openBrowser
	}

	presetAuthStyle := ""
	if *provider != "" {
		preset, ok := oauthflow.Presets[*provider]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown provider %q, known: %s\n", *provider, knownProviders())
			os.Exit(1)
		}
		cfg.AuthURL = preset.AuthURL
		cfg.TokenURL = preset.TokenURL
		cfg.DeviceAuthURL = preset.DeviceAuthURL
		cfg.AuthStyle = preset.AuthStyle
		cfg.ExtraAuthParams = preset.ExtraAuthParams
		presetAuthStyle = preset.AuthStyle
	}
	if *authURL != "" {
		cfg.AuthURL = *authURL
	}
	if *tokenURL != "" {
		cfg.TokenURL = *tokenURL
	}
	if *deviceAuthURL != "" {
		cfg.DeviceAuthURL = *deviceAuthURL
	}
	if *authStyle != "" {
		cfg.AuthStyle = *authStyle
		presetAuthStyle = *authStyle
	}
	if *scopes != "" {
		for _, scope := range strings.Split(*scopes, ",") {
			cfg.Scopes = append(cfg.Scopes, strings.TrimSpace(scope))
		}
	}

	if cfg.ClientID == "" {
		fmt.Fprintln(os.Stderr, "--client-id is required")
		os.Exit(1)
	}
	if *tokenFile == "" {
		fmt.Fprintln(os.Stderr, "--token-file is required")
		os.Exit(1)
	}
	if cfg.TokenURL == "" {
		fmt.Fprintln(os.Stderr, "--token-url is required (or use --provider)")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var token *oauth2.Token
	var err error
	switch *flow {
	case "device":
		token, err = oauthflow.DeviceLogin(ctx, cfg)
	case "browser":
		token, err = oauthflow.BrowserLogin(ctx, cfg)
	case "auto":
		if cfg.DeviceAuthURL != "" {
			token, err = oauthflow.DeviceLogin(ctx, cfg)
			if err != nil && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "device flow failed (%s), falling back to browser flow\n", err.Error())
				token, err = oauthflow.BrowserLogin(ctx, cfg)
			}
		} else {
			token, err = oauthflow.BrowserLogin(ctx, cfg)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown flow %q, use auto|device|browser\n", *flow)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = oauthflow.SaveTokenFile(*tokenFile, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to save token file: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Token saved to %s\n", *tokenFile)
	if token.RefreshToken == "" {
		fmt.Fprintln(os.Stderr, "Note: the provider returned no refresh token; the stored access token will be used until it expires")
	}

	fmt.Fprintln(os.Stderr, "\nUse it in a server_config like this:")
	fmt.Fprintf(os.Stdout, `{
  "oauth2": {
    "token_url": %q,
    "grant_type": "refresh_token",
    "client_id": %q,
    "client_secret": "{{{FromEnv=OAUTH_CLIENT_SECRET}}}",%s
    "token_file": %q
  }
}
`, cfg.TokenURL, cfg.ClientID, authStyleSnippet(presetAuthStyle), *tokenFile)
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func authStyleSnippet(style string) string {
	if style == "" {
		return ""
	}
	return fmt.Sprintf("\n    \"auth_style\": %q,", style)
}

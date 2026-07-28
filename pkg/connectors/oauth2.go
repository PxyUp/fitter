package connectors

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/oauthflow"
	"github.com/PxyUp/fitter/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	errOAuth2TokenUrl     = errors.New("oauth2: token_url is required")
	errOAuth2GrantType    = errors.New("oauth2: unsupported grant_type")
	errOAuth2RefreshToken = errors.New("oauth2: refresh_token or token_file is required for refresh_token grant")

	oauth2Mutex   sync.Mutex
	oauth2Sources = make(map[string]*cachedTokenSource)
)

type cachedTokenSource struct {
	source    oauth2.TokenSource
	tokenFile string
}

// persistingTokenSource writes every new access/refresh token pair to the token
// file, so rotated refresh tokens survive process restarts
type persistingTokenSource struct {
	src    oauth2.TokenSource
	path   string
	logger logger.Logger

	mu   sync.Mutex
	last *oauth2.Token
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	token, err := p.src.Token()
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.last == nil || p.last.AccessToken != token.AccessToken || p.last.RefreshToken != token.RefreshToken {
		errSave := oauthflow.SaveTokenFile(p.path, token)
		if errSave != nil {
			p.logger.Errorw("unable to persist oauth2 token", "path", p.path, "error", errSave.Error())
		} else {
			p.last = token
		}
	}

	return token, nil
}

// oauth2AccessToken returns a valid access token for the resolved config,
// creating (or reusing) a cached auto-refreshing token source. The returned
// key identifies the cache entry and can be passed to oauth2Evict to force
// a fresh token on the next call.
func oauth2AccessToken(client *http.Client, cfg *config.OAuth2Config, parsedValue builder.Interfacable, index *uint32, input builder.Interfacable, logger logger.Logger) (*oauth2.Token, string, error) {
	format := func(s string) string {
		return utils.Format(s, parsedValue, index, input)
	}

	tokenUrl := format(cfg.TokenUrl)
	if tokenUrl == "" {
		return nil, "", errOAuth2TokenUrl
	}

	grantType := cfg.GrantType
	if grantType == "" {
		grantType = config.ClientCredentialsGrant
	}
	if grantType != config.ClientCredentialsGrant && grantType != config.RefreshTokenGrant {
		return nil, "", errOAuth2GrantType
	}

	clientId := format(cfg.ClientId)
	clientSecret := format(cfg.ClientSecret)
	refreshToken := format(cfg.RefreshToken)
	tokenFile := format(cfg.TokenFile)
	if grantType == config.RefreshTokenGrant && refreshToken == "" && tokenFile == "" {
		return nil, "", errOAuth2RefreshToken
	}

	scopes := make([]string, 0, len(cfg.Scopes))
	for _, scope := range cfg.Scopes {
		scopes = append(scopes, format(scope))
	}

	endpointParams := url.Values{}
	for k, v := range cfg.EndpointParams {
		endpointParams.Set(format(k), format(v))
	}

	authStyle := oauth2.AuthStyleAutoDetect
	switch cfg.AuthStyle {
	case "header":
		authStyle = oauth2.AuthStyleInHeader
	case "params":
		authStyle = oauth2.AuthStyleInParams
	}

	paramsKeys := make([]string, 0, len(endpointParams))
	for k := range endpointParams {
		paramsKeys = append(paramsKeys, k+"="+endpointParams.Get(k))
	}
	sort.Strings(paramsKeys)
	key := strings.Join([]string{
		string(grantType), tokenUrl, clientId, clientSecret, refreshToken,
		strings.Join(scopes, " "), strings.Join(paramsKeys, "&"), cfg.AuthStyle, tokenFile,
	}, "\x00")

	oauth2Mutex.Lock()
	cached, ok := oauth2Sources[key]
	if !ok {
		var fileToken *oauth2.Token
		if tokenFile != "" {
			loaded, errLoad := oauthflow.LoadTokenFile(tokenFile)
			if errLoad == nil {
				fileToken = loaded
			} else if !errors.Is(errLoad, os.ErrNotExist) {
				logger.Errorw("unable to read oauth2 token file, falling back to config credentials", "path", tokenFile, "error", errLoad.Error())
			}
		}

		// background context: the source outlives the request that created it,
		// token fetches use the connector http client
		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)

		var source oauth2.TokenSource
		if grantType == config.ClientCredentialsGrant {
			source = (&clientcredentials.Config{
				ClientID:       clientId,
				ClientSecret:   clientSecret,
				TokenURL:       tokenUrl,
				Scopes:         scopes,
				EndpointParams: endpointParams,
				AuthStyle:      authStyle,
			}).TokenSource(ctx)
			if fileToken != nil {
				source = oauth2.ReuseTokenSource(fileToken, source)
			}
		} else {
			// the persisted token wins over the config seed: it holds the newest
			// refresh token after rotation and may carry a still-valid access token
			seed := fileToken
			if seed == nil {
				seed = &oauth2.Token{RefreshToken: refreshToken}
			} else if seed.RefreshToken == "" {
				seed.RefreshToken = refreshToken
			}
			if seed.RefreshToken == "" && !seed.Valid() {
				oauth2Mutex.Unlock()
				return nil, "", errOAuth2RefreshToken
			}
			source = (&oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Scopes:       scopes,
				Endpoint: oauth2.Endpoint{
					TokenURL:  tokenUrl,
					AuthStyle: authStyle,
				},
			}).TokenSource(ctx, seed)
		}

		if tokenFile != "" {
			source = &persistingTokenSource{
				src:    source,
				path:   tokenFile,
				logger: logger,
				last:   fileToken,
			}
		}

		cached = &cachedTokenSource{source: source, tokenFile: tokenFile}
		oauth2Sources[key] = cached
	}
	oauth2Mutex.Unlock()

	token, err := cached.source.Token()
	if err != nil {
		return nil, key, err
	}

	return token, key, nil
}

// oauth2Evict drops the cached token source; if a token file is used, the
// persisted access token is invalidated too (the refresh token is kept)
func oauth2Evict(key string) {
	oauth2Mutex.Lock()
	defer oauth2Mutex.Unlock()

	cached, ok := oauth2Sources[key]
	if !ok {
		return
	}
	delete(oauth2Sources, key)

	if cached.tokenFile == "" {
		return
	}
	token, err := oauthflow.LoadTokenFile(cached.tokenFile)
	if err != nil {
		return
	}
	token.AccessToken = ""
	_ = oauthflow.SaveTokenFile(cached.tokenFile, token)
}

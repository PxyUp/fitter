package oauthflow

// Preset holds well-known endpoints and quirks of an oauth2 provider
type Preset struct {
	AuthURL       string
	TokenURL      string
	DeviceAuthURL string
	// AuthStyle matches config.OAuth2Config.AuthStyle: "" auto, "header", "params"
	AuthStyle string
	// SupportsDeviceFlow reports whether the provider offers the device_code grant
	SupportsDeviceFlow bool
	// ExtraAuthParams are appended to the browser-flow authorization request
	ExtraAuthParams map[string]string
}

var Presets = map[string]Preset{
	"github": {
		AuthURL:            "https://github.com/login/oauth/authorize",
		TokenURL:           "https://github.com/login/oauth/access_token",
		DeviceAuthURL:      "https://github.com/login/device/code",
		AuthStyle:          "params",
		SupportsDeviceFlow: true,
	},
	"google": {
		AuthURL:            "https://accounts.google.com/o/oauth2/auth",
		TokenURL:           "https://oauth2.googleapis.com/token",
		DeviceAuthURL:      "https://oauth2.googleapis.com/device/code",
		AuthStyle:          "params",
		SupportsDeviceFlow: true,
		ExtraAuthParams:    map[string]string{"access_type": "offline", "prompt": "consent"},
	},
	"microsoft": {
		AuthURL:            "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL:           "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		DeviceAuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/devicecode",
		AuthStyle:          "params",
		SupportsDeviceFlow: true,
	},
	"gitlab": {
		AuthURL:            "https://gitlab.com/oauth/authorize",
		TokenURL:           "https://gitlab.com/oauth/token",
		DeviceAuthURL:      "https://gitlab.com/oauth/authorize_device",
		AuthStyle:          "params",
		SupportsDeviceFlow: true,
	},
	"spotify": {
		AuthURL:   "https://accounts.spotify.com/authorize",
		TokenURL:  "https://accounts.spotify.com/api/token",
		AuthStyle: "header",
	},
}

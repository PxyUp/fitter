//go:build !js

package connectors

import _ "embed"

// stealth.min.js is vendored from github.com/jonfriesen/playwright-go-stealth (MIT),
// which generates it from the puppeteer-extra-plugin-stealth evasions.
//
//go:embed stealth.min.js
var stealthJS string

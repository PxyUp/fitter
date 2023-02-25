# Fitter + Fitter CLI

Fitter - new way for collect information from the API's/Websites

Fitter CLI - small cli command which provide result from Fitter for test/debug/home usage

# Way to collect information

1. **Server** - parsing response from some API's or http request(usage of http.Client)
2. **Browser** - emulate real browser using chromium + docker + playwright/cypress and get DOM information

**Docker default image**: docker.io/zenika/alpine-chrome

# Format which can be parsed

1. **JSON** - parsing JSON to get specific information
2. **XML** - parsing xml tree to get specific information
3. **HTML** - parsing dom tree to get specific information
4. **XPath** - parsing dom tree to get specific information but by xpath

# Environment variables
1. **FITTER_HTTP_WORKER** - int[1000] - default concurrent HTTP workers

If you're using Docker like Browser connector:
1. **DOCKER_HOST** - string - (EnvOverrideHost) to set the URL to the docker server.
2. **DOCKER_API_VERSION** - string - (EnvOverrideAPIVersion) to set the version of the API to use, leave empty for latest.
3. **DOCKER_CERT_PATH** - string - (EnvOverrideCertPath) to specify the directory from which to load the TLS certificates (ca.pem, cert.pem, key.pem).
4. **DOCKER_TLS_VERIFY** - bool - (EnvTLSVerify) to enable or disable TLS verification (off by default)


# How to Fitter

[Download latest version from the release page](https://github.com/PxyUp/fitter/releases)

or locally:
```bash
go run cmd/fitter/main.go --path=./examples/config_api.json
```

### Arguments
1. **--path** - string[config.yaml] - path for the configuration of the Fitter

# How to use Fitter_CLI

[Download latest version from the release page](https://github.com/PxyUp/fitter/releases)

or locally:
```bash
go run cmd/cli/main.go --path=./examples/cli/config_cli.json
```

### Arguments
1. **--path** - string[config.yaml] - path for the configuration of the Fitter_CLI
2. **--copy** - bool[false] - copy information into clipboard
3. **--pretty** - bool[true] - make readable result(also affect on copy)
4. **--verbose** - bool[false] - enable logging

```bash
./fitter_cli_${VERSION} --path=./examples/cli/config_cli.json --copy=true
```

Examples:
1. [HackerNews + Quotes + Guardian News](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json) - using API + HTML + XPath parsing
2. **Chromium version** [Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_browser.json) - using HTML parsing + browser emulation
3. **Docker version** [Docker version: Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_docker.json) - using HTML parsing + browser from Docker image
4. **Playwright version** [Playwright version: Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_playwright.json) - using HTML parsing + browser from Playwright framework


### Limits

```json
{
  "limits": {
    "host_request_limiter": {
      "hacker-news.firebaseio.com": 5 // 5 concurrent request to how
    },
    "chromium_instance": 3, // Max allow 3 parallel chromium instance
    "docker_containers": 3, // Max allow 3 parallel docker containers
    "playwright_instance": 3 // Max allow 3 parallel playwright instance
  },
  "item": {
    ...
  }
}
```

Example [here](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json#L3)

# Roadmap

1. Add browser emulation via: Docker/Cypress(for run scenario)
2. Add trigger method for Fitter: Webhook/Queue
3. Add notification methods for Fitter: Webhook/Queue
# Fitter + Fitter CLI

Fitter - new way for collect information from the API's/Websites

Fitter CLI - small cli command which provide result from Fitter for test/debug/home usage

Fitter Lib - library which provide functional of fitter CLI as a library

![](https://github.com/PxyUp/fitter/blob/master/demo.gif)


# Way to collect information

1. **Server** - parsing response from some API's or http request(usage of http.Client)
2. **Browser** - emulate real browser using chromium + docker + playwright/cypress and get DOM information

**Docker default image**: docker.io/zenika/alpine-chrome

# Format which can be parsed

1. **JSON** - parsing JSON to get specific information
2. **XML** - parsing xml tree to get specific information
3. **HTML** - parsing dom tree to get specific information
4. **XPath** - parsing dom tree to get specific information but by xpath

# Use like a library

```bash
go get github.com/PxyUp/fitter
```

```go
package main

import (
	"fmt"
	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/config"
	"log"
	"net/http"
)

func main() {
	res, err := lib.Parse(&config.Item{
		ConnectorConfig: &config.ConnectorConfig{
			ConnectorType: config.Server,
			ResponseType:  config.Json,
			Url:    "https://random-data-api.com/api/appliance/random_appliance",
			ServerConfig: &config.ServerConnectorConfig{
				Method: http.MethodGet,
			},
		},
		Model: &config.Model{
			Type: config.ObjectModel,
			ObjectConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"my_id": {
						BaseField: &config.BaseField{
							Type: config.Int,
							Path: "id",
						},
					},
					"generated_id": {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								UUID: &config.UUIDGeneratedFieldConfig{},
							},
						},
					},
					"generated_array": {
						ArrayConfig: &config.ArrayConfig{
							RootPath: "@this|@keys",
							ItemConfig: &config.ObjectConfig{
								Field: &config.BaseField{
									Type: config.String,
								},
							},
						},
					},
				},
			},
		},
	}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.ToJson())
}

```

Output:

```json
{
  "generated_array": ["id","uid","brand","equipment"],
  "my_id": 6000,
  "generated_id": "26b08b73-2f2e-444d-bcf2-dac77ac3130e"
}
```

# How to use Fitter

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
1. **Server version** [HackerNews + Quotes + Guardian News](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json) - using API + HTML + XPath parsing
2. **Chromium version** [Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_browser.json) - using HTML parsing + browser emulation
3. **Docker version** [Docker version: Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_docker.json) - using HTML parsing + browser from Docker image
4. **Playwright version** [Playwright version: Guardian News + Quotes](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_playwright.json) - using HTML parsing + browser from Playwright framework
5. **Playwright version** [Playwright version: England Cities + Weather](https://github.com/PxyUp/fitter/blob/master/examples/cli/config_weather.json) - using HTML + XPath parsing + browser from Playwright framework

# Configuration

## Connector
It is the way how you fetch the data

```go
type ConnectorConfig struct {
    ResponseType  ParserType              `json:"response_type" yaml:"response_type"`
    ConnectorType Connector               `json:"connector_type" yaml:"connector_type"`
    ServerConfig  *ServerConnectorConfig  `json:"server_config" yaml:"server_config"`
    JsonConfig    *JsonConnectorConfig    `json:"json_config" yaml:"json_config"`
    BrowserConfig *BrowserConnectorConfig `yaml:"browser_config" json:"browser_config"`
    Attempts      uint32                  `json:"attempts" yaml:"attempts"`
    Url           string                  `json:"url" yaml:"url"`
}
```

- ResponseType - enum["HTML", "json","xpath"] - in which format data comes from the connector
- ConnectorType - enum["server", "browser", "json"] - way how to fetch data
- Attempts - how many attempts to use for fetch data by connector
- Url - define which address to request

Config can be one of:
- [ServerConfig](#serverconnectorconfig) - only if type of connector "server"
- [BrowserConfig](#browserconnectorconfig) - only if type of connector "browser"
- [JsonConfig](#jsonconnectorconfig) - only if type of connector "json"

Example:
```json
{
  "response_type": "xpath",
  "connector_type": "browser",
  "attempts": 3,
  "url": "https://openweathermap.org/find?q={PL}",
  "browser_config": {
    "playwright": {
      "timeout": 30,
      "wait": 30,
      "install": false,
      "browser": "Chromium"
    }
  }
}
```

### JsonConnectorConfig
Connector type which fetch data from provided JSON string
```go
type JsonConnectorConfig struct {
	Json string `json:"json" yaml:"json"`
}
```

- Json - valid json string


Example:

https://github.com/PxyUp/fitter/blob/master/examples/cli/config_connector_json.json#L5
```json
{
  "json": "[1,2,3,4,5]"
}
```

### ServerConnectorConfig
Connector type which fetch data using golang http.Client(server side request like curl)

```go
type ServerConnectorConfig struct {
    Method  string            `json:"method" yaml:"method"`
    Headers map[string]string `yaml:"headers" json:"headers"`
}
```

- Method - supported all http methods: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
- Headers - predefine headers for using during request

Example:
```json
{
  "method": "GET",
  "url": "https://hacker-news.firebaseio.com/v0/beststories.json?print=pretty&limitToFirst=10&orderBy=%22$key%22"
}
```

Right now default timeout it is 10 sec.

##### Environment variables
1. **FITTER_HTTP_WORKER** - int[1000] - default concurrent HTTP workers

### BrowserConnectorConfig
Connector type which emulate fetching of data via browser

```go
type BrowserConnectorConfig struct {
	Chromium   *ChromiumConfig   `json:"chromium" yaml:"chromium"`
	Docker     *DockerConfig     `json:"docker" yaml:"docker"`
	Playwright *PlaywrightConfig `json:"playwright" yaml:"playwright"`
}
```

Config can be one of:
- [Chromium](#chromium) - use local installed Chromium for fetch data
- [Docker](#docker) - use docker as service for spin up container for fetch data
- [Playwright](#playwright) - use playwright framework for fetch data

Example:
```json
{
    "url": "https://www.theguardian.com/world",
    "docker": {
      "wait": 10000,
      "image": "docker.io/zenika/alpine-chrome:with-node",
      "entry_point": "chromium-browser",
      "purge": true
    }
}
```

#### Chromium
Use locally installed Chromium for fetch the data

```go
type ChromiumConfig struct {
	Path    string   `yaml:"path" json:"path"`
	Timeout uint32   `yaml:"timeout" json:"timeout"`
	Wait    uint32   `yaml:"wait" json:"wait"`
	Flags   []string `yaml:"flags" json:"flags"`
}
```

- Path - path to binary of Chromium
- Timeout[sec] - timeout for execution of the chromium
- Wait[msec] - timeout of page loading
- Flags - flags for Chromium default: "--headless", "--proxy-auto-detect", "--temp-profile", "--incognito", "--disable-logging", "--disable-gpu"

Example:
```json
{
  "path": "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  "wait": 10000
}
```

#### Docker
Use Docker for spin up container for fetch data

```go
type DockerConfig struct {
	Image       string   `yaml:"image" json:"image"`
	EntryPoint  string   `json:"entry_point" yaml:"entry_point"`
	Timeout     uint32   `yaml:"timeout" json:"timeout"`
	Wait        uint32   `yaml:"wait" json:"wait"`
	Flags       []string `yaml:"flags" json:"flags"`
	Purge       bool     `json:"purge" yaml:"purge"`
	NoPull      bool     `yaml:"no_pull" json:"no_pull"`
	PullTimeout uint32   `yaml:"pull_timeout" json:"pull_timeout"`
}
```

- Image - image for the docker registry(please provide which register host)
- EntryPoint - cmd which will be run inside container
- Timeout[sec] - timeout for run container(without pulling image)
- Wait[msec] - timeout of page loading (works just for Chromium based containers)
- Flags - cmd arguments for run containers, default for Chromium based: "--no-sandbox","--headless", "--proxy-auto-detect", "--temp-profile", "--incognito", "--disable-logging", "--disable-gpu"
- Purge - should we remove container after work done(like docker rm)
- NoPull - prevent pulling of the image
- PullTimeout - define timeout for pull contains

##### Environment variables
1. **DOCKER_HOST** - string - (EnvOverrideHost) to set the URL to the docker server.
2. **DOCKER_API_VERSION** - string - (EnvOverrideAPIVersion) to set the version of the API to use, leave empty for latest.
3. **DOCKER_CERT_PATH** - string - (EnvOverrideCertPath) to specify the directory from which to load the TLS certificates (ca.pem, cert.pem, key.pem).
4. **DOCKER_TLS_VERIFY** - bool - (EnvTLSVerify) to enable or disable TLS verification (off by default)


Example:
```json
{
  "wait": 10000,
  "image": "docker.io/zenika/alpine-chrome:with-node",
  "entry_point": "chromium-browser",
  "purge": true
}
```

#### Playwright
Run browsers via playwright framework

```go
type PlaywrightConfig struct {
	Browser    PlaywrightBrowser          `json:"browser" yaml:"browser"`
	Install    bool                       `yaml:"install" json:"install"`
	Timeout    uint32                     `yaml:"timeout" json:"timeout"`
	Wait       uint32                     `yaml:"wait" json:"wait"`
	TypeOfWait *playwright.WaitUntilState `json:"type_of_wait" yaml:"type_of_wait"`
}
```

- Browser - enum["Chromium", "FireFox", "WebKit"] - which browser to use
- Install - should we install browser
- Timeout[sec] - timeout to run playwright
- Wait[sec] - timeout of page loading
- TypeOfWait - enum["load", "domcontentloaded", "networkidle", "commit"] which state of page we waiting, default is "load"

Example
```json
{
  "timeout": 30,
  "wait": 30,
  "install": false,
  "browser": "Chromium"
}
```

## Model
With model we define result of the scrapping

```go
type Model struct {
	Type         ModelType     `yaml:"type" json:"type"`
	ObjectConfig *ObjectConfig `yaml:"object_config" json:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`
}
```

- Type - enum["object", "array"] - in which format we expect result

Config can be one of:
- [ObjectConfig](#objectconfig) - configuration of object format, only if type "object"
- [ArrayConfig](#arrayconfig) - configuration of array format, only if type "array"

Example:
```json
{
  "type": "object",
  "object_config": {}
}
```

### ObjectConfig
Configuration of the object and fields

```go
type ObjectConfig struct {
    Fields      map[string]*Field `json:"fields" yaml:"fields"`
    Field       *BaseField        `json:"field" yaml:"field"`
    ArrayConfig *ArrayConfig      `json:"array_config" yaml:"array_config"`
}
```

Config can be one of:
- [Fields](#field) - map of each field definition; key - field name, value - configuration
- [Field](#basefield) - used for element of array; fields which will be deserialized like basic type like "string", "int" and etc (used here for case array of basic types)
- [ArrayConfig](#arrayconfig) - used for element of array; deserialization array of array

Example:
```json
{
  "fields": {
    "title": {
      "base_field": {
        "type": "string",
        "path": "type"
      }
    }
  }
}
```

### ArrayConfig
Configuration of the array and fields

```go
type ArrayConfig struct {
	RootPath     string             `json:"root_path" yaml:"root_path"`
	ItemConfig   *ObjectConfig      `json:"item_config" yaml:"item_config"`
	StaticConfig *StaticArrayConfig `json:"static_array"  yaml:"static_array"`
}
```

- RootPath - selector for find root element of the array or repeated element in case of html parsing, size of array will be amount of children element under the root

Config can be one of:
- [ItemConfig](#objectconfig) - configuration of each element of the array
- [StaticConfig](#static-array-config) - configuration of the static array

Example:
```json
{
  "root_path": "#content dt.quote > a",
  "item_config": {
    "field": {
      "type": "string"
    }
  }
}
```

#### Field
Common of the field


```go
type Field struct {
	BaseField    *BaseField    `json:"base_field" yaml:"base_field"`
	ObjectConfig *ObjectConfig `yaml:"object_config" yaml:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`

	FirstOf []*Field `json:"first_of" yaml:"first_of"`
}
```

Config can be one of: 
- [BaseField](#basefield) - fields which will be deserialized like basic type like "string", "int" and etc
- [ObjectConfig](#objectconfig) - in case our field in nested object
- [ArrayConfig](#arrayconfig) - in case our field in array
- [FirstOf](#field) - first not empty resolved field will be selected

Example:
```json
{
  "base_field": {
    "type": "string",
    "path": "div.current-temp span.heading"
  }
}
```

#### BaseField
In case we want get some static information or generate new one

```go
type BaseField struct {
	Type FieldType `yaml:"type" json:"type"`
	Path string    `yaml:"path" json:"path"`

	Generated *GeneratedFieldConfig `yaml:"generated" json:"generated"`

	FirstOf []*BaseField `json:"first_of" yaml:"first_of"`
}
```

- FieldType - enum["null", "boolean", "string", "int","int64","float","float64"] - static field for parse
- Path - selector(relative in case it is array child) for parsing

Config can be one of or empty:
- [Generated](#generatedfieldconfig) - field can be generated one which custom configuration
- [FirstOf](#basefield) - first not empty resolved field will be selected

Examples
```json
{
  "generated": {
    "uuid": {}
  }
}
```

```json
{
  "type": "string",
  "path": "text()"
}
```

#### GeneratedFieldConfig
Provide functionality of generating field on the flight

```go
type GeneratedFieldConfig struct {
	UUID      *UUIDGeneratedFieldConfig   `yaml:"uuid" json:"uuid"`
	Static    *StaticGeneratedFieldConfig `yaml:"static" json:"static"`
	Formatted *FormattedFieldConfig       `json:"formatted" yaml:"formatted"`
	Model     *ModelField                 `yaml:"model" json:"model"`
}
```

Config can be one of:
- [UUID](#uuid) - generate random UUID V4
- [Static](#static) - generate static field
- [Formatted](#formatted-field-config) - format field
- [Model](#model-field) - model generated from the other connector and model

Examples:
```json
{
    "uuid": {}
}
```

https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json#L58
```json
{
    "model": {
      "type": "array",
      "model": {
        "type": "array",
        "array_config": {
          "root_path": "#content dt.quote > a",
          "item_config": {
            "field": {
              "type": "string"
            }
          }
        }
      },
      "connector_config": {
        "response_type": "HTML",
        "connector_type": "browser",
        "attempts": 3,
        "browser_config": {
          "url": "http://www.quotationspage.com/random.php",
          "chromium": {
            "path": "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
            "wait": 10000
          }
        }
      }
    }
}
```

#### UUID
Generate random UUID V4 on the flight, can be used for generate uniq id

```go
type UUIDGeneratedFieldConfig struct {
	Regexp string `yaml:"regexp" json:"regexp"`
}
```

- Regexp - provide matcher which can be used for get part of generated uuid

#### Static
Generate static field

```go
type StaticGeneratedFieldConfig struct {
	Type  FieldType `yaml:"type" json:"type"`
	Value string    `json:"value" yaml:"value"`
}
```

- Type - enum["null", "boolean", "string", "int","int64","float","float64"] - type of the field
- Value - string value of the field

Example
```json
{
  "type": "int",
  "value": "65"
}
```

#### Formatted Field Config
Generate formatted field which will pass value from parent [base field](#basefield)

```go
type FormattedFieldConfig struct {
	Template string `yaml:"template" json:"template"`
}
```

- Template - template in with placeholder [{PL}](#placeholder-list) where [parent](#basefield) value will be injected like string
 
Example:
https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json#L98
```json
{
  "template": "https://news.ycombinator.com/item?id={PL}"
}
```

#### Model Field
Field type which can be generated on the flight by news [model](#model) and [connector](#connector)

```go
type ModelField struct {
	// Type of parsing
	ConnectorConfig *ConnectorConfig `yaml:"connector_config" json:"connector_config"`
	// Model of the response
	Model *Model `yaml:"model" json:"model"`

	Type GeneratedFieldType `yaml:"type" json:"type"`
	Path string             `yaml:"path" json:"path"`
}
```

- [ConnectorConfig](#connector) - which connector to use. Important: URL in the connector can be with [inject of the parent value as a string](#placeholder-list)
- [Model](#model) - configuration of the underhood model
- GeneratedFieldType - enum["null", "boolean", "string", "int","int64","float","float64", "array", "object"] - type of generated field
- Path - in case we cant extract some information from generated field we can use json selector for extract

Examples:

https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json#L60
```json
{
  "type": "array",
  "model": {
    "type": "array",
    "array_config": {
      "root_path": "#content dt.quote > a",
      "item_config": {
        "field": {
          "type": "string"
        }
      }
    }
  }
}
```

https://github.com/PxyUp/fitter/blob/master/examples/cli/config_weather.json#L37
```json
{
    "type": "string",
    "path": "temp.temp",
    "model": {
      "type": "object",
      "object_config": {
        "fields": {
          "temp": {
            "base_field": {
              "type": "string",
              "path": "//div[@id='forecast_list_ul']//td/b/a/@href",
              "generated": {
                "model": {
                  "type": "string",
                  "model": {
                    "type": "object",
                    "object_config": {
                      "fields": {
                        "temp": {
                          "base_field": {
                            "type": "string",
                            "path": "div.current-temp span.heading"
                          }
                        }
                      }
                    }
                  },
                  "connector_config": {
                    "response_type": "HTML",
                    "connector_type": "browser",
                    "attempts": 4,
                    "url": "https://openweathermap.org{PL}",
                    "browser_config": {
                      "playwright": {
                        "timeout": 30,
                        "wait": 30,
                        "install": false,
                        "browser": "FireFox",
                        "type_of_wait": "networkidle"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "connector_config": {
      "response_type": "xpath",
      "connector_type": "browser",
      "attempts": 3,
      "url": "https://openweathermap.org/find?q={PL}",
      "browser_config": {
        "playwright": {
          "timeout": 30,
          "wait": 30,
          "install": false,
          "browser": "Chromium"
        }
      }
    }
}
```

#### Static Array Config
Provide static(fixed length) array generation

```go
type StaticArrayConfig struct {
    Items map[uint32]*Field `yaml:"items" json:"items"`
    Length uint32            `yaml:"length" json:"length"`
}
```
- [Items](#field) - map[uint32]*[Field](#field) - key is index in array, value is field definition
- Length - if set(1+) can be used for define custom length of array

Examples:
```json
{
  "0": {
    "base_field": {
      "type": "string",
      "path": "div.current-temp span.heading"
    }
  }
}
```

```json
{
  "length": 4,
  "0": {
    "base_field": {
      "type": "string",
      "path": "div.current-temp span.heading"
    }
  }
}
```

##### Placeholder list
1. {PL} - for inject value
2. {INDEX} - for inject index in parent array
3. {HUMAN_INDEX} - for inject index in parent array in human way

## Limits
Provide limitation for prevent DDOS, big usage of memory

```go
type Limits struct {
	HostRequestLimiter HostRequestLimiter `yaml:"host_request_limiter" json:"host_request_limiter"`
	ChromiumInstance   uint32             `yaml:"chromium_instance" json:"chromium_instance"`
	DockerContainers   uint32             `yaml:"docker_containers" json:"docker_containers"`
	PlaywrightInstance uint32             `yaml:"playwright_instance" json:"playwright_instance"`
}
```

- HostRequestLimiter - map[string]int64 - limitation per host name, key is host, value is amount of parallel request(usage for [server connector](#serverconnectorconfig))
- ChromiumInstance - amount of parallel [chromium](#chromium) instance
- DockerContainers - amount of parallel [docker](#docker) instance
- PlaywrightInstance - amount of parallel [playwright](#playwright) instance

https://github.com/PxyUp/fitter/blob/master/examples/cli/config_cli.json#L2
```json
{
  "limits": {
    "host_request_limiter": {
      "hacker-news.firebaseio.com": 5
    },
    "chromium_instance": 3,
    "docker_containers": 3,
    "playwright_instance": 3
  }
}
```

# Roadmap

1. Add browser scenario for preparing, after parsing
2. Add scrolling support for scenario
3. Add pagination support for scenario
4. Add notification methods for Fitter: Webhook/Queue

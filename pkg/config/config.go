package config

import (
	"encoding/json"
	"github.com/playwright-community/playwright-go"
)

type Connector string

type ParserType string

type ModelType string

const (
	HTML  ParserType = "HTML"
	Json  ParserType = "json"
	XML   ParserType = "XML"
	XPath ParserType = "xpath"
)

type HostRequestLimiter map[string]int64

type Limits struct {
	HostRequestLimiter HostRequestLimiter `yaml:"host_request_limiter" json:"host_request_limiter"`
	ChromiumInstance   uint32             `yaml:"chromium_instance" json:"chromium_instance"`
	DockerContainers   uint32             `yaml:"docker_containers" json:"docker_containers"`
	PlaywrightInstance uint32             `yaml:"playwright_instance" json:"playwright_instance"`
}

type Config struct {
	Items  []*Item `yaml:"items" json:"items"`
	Limits *Limits `yaml:"limits" json:"limits"`
}

type CliItem struct {
	Item   *Item   `yaml:"item" json:"item"`
	Limits *Limits `yaml:"limits" json:"limits"`
}

type ObjectConfig struct {
	Field       *BaseField        `json:"field" yaml:"field"`
	Fields      map[string]*Field `json:"fields" yaml:"fields"`
	ArrayConfig *ArrayConfig      `json:"array_config" yaml:"array_config"`
}

type ArrayConfig struct {
	RootPath    string        `json:"root_path" yaml:"root_path"`
	ItemConfig  *ObjectConfig `json:"item_config" yaml:"item_config"`
	LengthLimit uint32        `json:"length_limit" yaml:"length_limit"`

	StaticConfig *StaticArrayConfig `json:"static_array"  yaml:"static_array"`
}

type StaticArrayConfig struct {
	Items  map[uint32]*Field `yaml:"items" json:"items"`
	Length uint32            `yaml:"length" json:"length"`
}

type Model struct {
	ObjectConfig *ObjectConfig `yaml:"object_config" json:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`
	BaseField    *BaseField    `json:"base_field" yaml:"base_field"`
}

type ConnectorConfig struct {
	ResponseType ParserType `json:"response_type" yaml:"response_type"`
	Url          string     `json:"url" yaml:"url"`
	Attempts     uint32     `json:"attempts" yaml:"attempts"`

	StaticConfig          *StaticConnectorConfig  `json:"static_config" yaml:"static_config"`
	ServerConfig          *ServerConnectorConfig  `json:"server_config" yaml:"server_config"`
	BrowserConfig         *BrowserConnectorConfig `yaml:"browser_config" json:"browser_config"`
	PluginConnectorConfig *PluginConnectorConfig  `json:"plugin_connector_config" yaml:"plugin_connector_config"`
}

type PlaywrightBrowser string

const (
	Chromium PlaywrightBrowser = "Chromium"
	FireFox  PlaywrightBrowser = "FireFox"
	WebKit   PlaywrightBrowser = "WebKit"
)

type PlaywrightConfig struct {
	Browser    PlaywrightBrowser          `json:"browser" yaml:"browser"`
	Install    bool                       `yaml:"install" json:"install"`
	Timeout    uint32                     `yaml:"timeout" json:"timeout"`
	Wait       uint32                     `yaml:"wait" json:"wait"`
	TypeOfWait *playwright.WaitUntilState `json:"type_of_wait" yaml:"type_of_wait"`
}

type StaticConnectorConfig struct {
	Value string `json:"value" yaml:"value"`
}

type BrowserConnectorConfig struct {
	Chromium   *ChromiumConfig   `json:"chromium" yaml:"chromium"`
	Docker     *DockerConfig     `json:"docker" yaml:"docker"`
	Playwright *PlaywrightConfig `json:"playwright" yaml:"playwright"`
}

type PluginConnectorConfig struct {
	Name   string          `json:"name" yaml:"name"`
	Config json.RawMessage `json:"config" yaml:"config"`
}

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

type ChromiumConfig struct {
	Path    string   `yaml:"path" json:"path"`
	Timeout uint32   `yaml:"timeout" json:"timeout"`
	Wait    uint32   `yaml:"wait" json:"wait"`
	Flags   []string `yaml:"flags" json:"flags"`
}

type ServerConnectorConfig struct {
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Timeout uint32            `yaml:"timeout" json:"timeout"`
	Body    string            `yaml:"body" json:"body"`
}

type TriggerConfig struct {
	SchedulerTrigger *SchedulerTrigger `yaml:"scheduler_trigger" json:"scheduler_trigger"`
	HTTPTrigger      *HTTPTrigger      `json:"http_trigger" yaml:"http_trigger"`
}

type SchedulerTrigger struct {
	// Interval for update rerun process in second
	Interval int `yaml:"interval" json:"interval"`
}

type HTTPTrigger struct {
}

type NotifierConfig struct {
	Expression string `yaml:"expression" json:"expression"`
	Force      bool   `json:"force" yaml:"force"`

	Console     *ConsoleConfig     `yaml:"console" json:"console"`
	TelegramBot *TelegramBotConfig `yaml:"telegram_bot" json:"telegram_bot"`
}

type ConsoleConfig struct {
}

type TelegramBotConfig struct {
	Token           string  `json:"token" yaml:"token"`
	UsersId         []int64 `json:"users_id" yaml:"users_id"`
	Pretty          bool    `json:"pretty" yaml:"pretty"`
	SendArrayByItem bool    `yaml:"send_array_by_item" json:"send_array_by_item"`
}

type Item struct {
	Name string `yaml:"name" json:"name"`

	// Type of parsing
	ConnectorConfig *ConnectorConfig `yaml:"connector_config" json:"connector_config"`

	// TriggerConfig
	TriggerConfig *TriggerConfig `yaml:"trigger_config" json:"trigger_config"`
	// Model of the response
	Model *Model `yaml:"model" json:"model"`
	// Where to report result
	NotifierConfig *NotifierConfig `json:"notifier_config" yaml:"notifier_config"`
}

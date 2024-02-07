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
type RefMap map[string]*Reference

type Reference struct {
	*ModelField

	Expire *uint32 `yaml:"expire" json:"expire"`
}

type Limits struct {
	HostRequestLimiter HostRequestLimiter `yaml:"host_request_limiter" json:"host_request_limiter"`
	ChromiumInstance   uint32             `yaml:"chromium_instance" json:"chromium_instance"`
	DockerContainers   uint32             `yaml:"docker_containers" json:"docker_containers"`
	PlaywrightInstance uint32             `yaml:"playwright_instance" json:"playwright_instance"`
}

type Config struct {
	Items []*Item `yaml:"items" json:"items"`

	Limits     *Limits `yaml:"limits" json:"limits"`
	References RefMap  `json:"references" yaml:"references"`

	HttpServer *HttpServerCfg `json:"http_server" yaml:"http_server"`
}

type HttpServerCfg struct {
	Port int `yaml:"port" json:"port"`
}

type CliItem struct {
	Item *Item `yaml:"item" json:"item"`

	Limits     *Limits `yaml:"limits" json:"limits"`
	References RefMap  `json:"references" yaml:"references"`
}

type ObjectConfig struct {
	Field       *BaseField        `json:"field" yaml:"field"`
	Fields      map[string]*Field `json:"fields" yaml:"fields"`
	ArrayConfig *ArrayConfig      `json:"array_config" yaml:"array_config"`
}

type ArrayConfig struct {
	RootPath    string        `json:"root_path" yaml:"root_path"`
	Reverse     bool          `yaml:"reverse" json:"reverse"`
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
	IsArray      bool          `json:"is_array" yaml:"is_array"`
}

type ConnectorConfig struct {
	ResponseType ParserType `json:"response_type" yaml:"response_type"`
	Url          string     `json:"url" yaml:"url"`
	Attempts     uint32     `json:"attempts" yaml:"attempts"`

	StaticConfig          *StaticConnectorConfig      `json:"static_config" yaml:"static_config"`
	IntSequenceConfig     *IntSequenceConnectorConfig `json:"int_sequence_config" yaml:"int_sequence_config"`
	ServerConfig          *ServerConnectorConfig      `json:"server_config" yaml:"server_config"`
	BrowserConfig         *BrowserConnectorConfig     `yaml:"browser_config" json:"browser_config"`
	PluginConnectorConfig *PluginConnectorConfig      `json:"plugin_connector_config" yaml:"plugin_connector_config"`
	ReferenceConfig       *ReferenceConnectorConfig   `yaml:"reference_config" json:"reference_config"`
	FileConfig            *FileConnectorConfig        `json:"file_config" yaml:"file_config"`
}

type FileConnectorConfig struct {
	Path          string `yaml:"path" json:"path"`
	UseFormatting bool   `yaml:"use_formatting" json:"use_formatting"`
}

type IntSequenceConnectorConfig struct {
	Start int `json:"start" yaml:"start"`
	End   int `json:"end" yaml:"end"`
	Step  int `json:"step" yaml:"step"`
}

type ReferenceConnectorConfig struct {
	Name string `yaml:"name" json:"name"`
}

type PlaywrightBrowser string

const (
	Chromium PlaywrightBrowser = "Chromium"
	FireFox  PlaywrightBrowser = "FireFox"
	WebKit   PlaywrightBrowser = "WebKit"
)

type PlaywrightConfig struct {
	Browser      PlaywrightBrowser          `json:"browser" yaml:"browser"`
	Install      bool                       `yaml:"install" json:"install"`
	Timeout      uint32                     `yaml:"timeout" json:"timeout"`
	Wait         uint32                     `yaml:"wait" json:"wait"`
	TypeOfWait   *playwright.WaitUntilState `json:"type_of_wait" yaml:"type_of_wait"`
	PreRunScript string                     `json:"pre_run_script" yaml:"pre_run_script"`

	Proxy *ProxyConfig `json:"proxy" yaml:"proxy"`
}

type StaticConnectorConfig struct {
	Value string          `json:"value" yaml:"value"`
	Raw   json.RawMessage `json:"raw" yaml:"raw"`
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

	Proxy *ProxyConfig `yaml:"proxy" json:"proxy"`
}

type ProxyConfig struct {
	// Proxy to be used for all requests. HTTP and SOCKS proxies are supported, for example
	// `http://myproxy.com:3128` or `socks5://myproxy.com:3128`. Short form `myproxy.com:3128`
	// is considered an HTTP proxy.
	Server string `json:"server" yaml:"server"`
	// Optional username to use if HTTP proxy requires authentication.
	Username string `json:"username" yaml:"username"`
	// Optional password to use if HTTP proxy requires authentication.
	Password string `json:"password" yaml:"password"`
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
	Expression      string `yaml:"expression" json:"expression"`
	Force           bool   `json:"force" yaml:"force"`
	SendArrayByItem bool   `yaml:"send_array_by_item" json:"send_array_by_item"`

	Console     *ConsoleConfig       `yaml:"console" json:"console"`
	TelegramBot *TelegramBotConfig   `yaml:"telegram_bot" json:"telegram_bot"`
	Http        *HttpConfig          `yaml:"http" json:"http"`
	Redis       *RedisNotifierConfig `json:"redis" yaml:"redis"`
	File        *FileStorageField    `json:"file" yaml:"file"`
}

type HttpConfig struct {
	Url     string            `yaml:"url" json:"url"`
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Timeout uint32            `yaml:"timeout" json:"timeout"`
}

type ConsoleConfig struct {
	OnlyResult bool `json:"only_result" yaml:"only_result"`
}

type RedisNotifierConfig struct {
	Addr     string `json:"addr" yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
	Channel  string `json:"channel" yaml:"channel"`
}

type TelegramBotConfig struct {
	Token   string  `json:"token" yaml:"token"`
	UsersId []int64 `json:"users_id" yaml:"users_id"`
	Pretty  bool    `json:"pretty" yaml:"pretty"`
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

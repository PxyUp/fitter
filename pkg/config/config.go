package config

type Connector string

type ParserType string

type ModelType string

const (
	Browser Connector = "browser"
	Server  Connector = "server"

	HTML  ParserType = "HTML"
	Json  ParserType = "json"
	XML   ParserType = "XML"
	XPath ParserType = "xpath"

	ObjectModel ModelType = "object"
	ArrayModel  ModelType = "array"
)

type HostRequestLimiter map[string]int64

type Config struct {
	Items              []*Item            `yaml:"items" json:"items"`
	HostRequestLimiter HostRequestLimiter `yaml:"host_request_limiter" json:"host_request_limiter"`
}

type CliItem struct {
	Item               *Item              `yaml:"item" json:"item"`
	HostRequestLimiter HostRequestLimiter `yaml:"host_request_limiter" json:"host_request_limiter"`
}

type ObjectConfig struct {
	Field       *BaseField        `json:"field" yaml:"field"`
	Fields      map[string]*Field `json:"fields" yaml:"fields"`
	ArrayConfig *ArrayConfig      `json:"array_config" yaml:"array_config"`
}

type ArrayConfig struct {
	RootPath   string        `json:"root_path" yaml:"root_path"`
	ItemConfig *ObjectConfig `json:"item_config" yaml:"item_config"`
}

type Model struct {
	Type         ModelType     `yaml:"type" json:"type"`
	ObjectConfig *ObjectConfig `yaml:"object_config" json:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`
}

type ConnectorConfig struct {
	ResponseType  ParserType              `json:"response_type" yaml:"response_type"`
	ConnectorType Connector               `json:"connector_type" yaml:"connector_type"`
	ServerConfig  *ServerConnectorConfig  `json:"server_config" yaml:"server_config"`
	BrowserConfig *BrowserConnectorConfig `yaml:"browser_config" json:"browser_config"`
}

type BrowserConnectorConfig struct {
}

type ServerConnectorConfig struct {
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Url     string            `json:"url" yaml:"url"`
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
	Console bool `yaml:"console" json:"console"`
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

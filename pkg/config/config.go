package config

import "time"

type Connector string

type ParserType string

type ModelType string

const (
	Browser Connector = "browser"
	Server  Connector = "server"

	DOM  ParserType = "DOM"
	Json ParserType = "Json"

	ObjectModel ModelType = "object"
	ArrayModel  ModelType = "array"
)

type Config struct {
	Items []*Item `yaml:"items" json:"items"`
}

type EntryConfig struct {
	Fields map[string]*Field
}

type ObjectConfig struct {
	Config *EntryConfig `yaml:"config" json:"config"`
}

type ArrayConfig struct {
	RootPath   string       `json:"root_field" yaml:"root_field"`
	ItemConfig *EntryConfig `json:"item_config" yaml:"item_config"`
}

type Model struct {
	Model        ModelType     `yaml:"model" json:"model"`
	ObjectConfig *ObjectConfig `yaml:"object_config" json:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`
}

type ParserConfig struct {
	Type  ParserType `json:"type" yaml:"type"`
	Model *Model     `yaml:"model" json:"model"`
}

type ConnectorConfig struct {
}

type Item struct {
	// Interval for update rerun process
	Interval time.Duration `yaml:"interval" json:"interval"`
	// Type of parsing
	ConnectorConfig *ConnectorConfig `yaml:"connector_config" json:"connector_config"`
	// Type of parsing response
	ParserConfig *ParserConfig `json:"parser_config" yaml:"parser_config"`
}

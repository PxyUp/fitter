package config

import "encoding/json"

type FieldType string

const (
	Null       FieldType = "null"
	Bool       FieldType = "boolean"
	String     FieldType = "string"
	Int        FieldType = "int"
	Int64      FieldType = "int64"
	Float      FieldType = "float"
	Float64    FieldType = "float64"
	HtmlString FieldType = "html"
	RawString  FieldType = "raw_string"

	Array  FieldType = "array"
	Object FieldType = "object"
)

type Field struct {
	BaseField    *BaseField    `json:"base_field" yaml:"base_field"`
	ObjectConfig *ObjectConfig `json:"object_config" yaml:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`

	FirstOf []*Field `json:"first_of" yaml:"first_of"`
}

type BaseField struct {
	Type FieldType `yaml:"type" json:"type"`
	Path string    `yaml:"path" json:"path"`

	HTMLAttribute string `json:"html_attribute" yaml:"html_attribute"`

	Generated *GeneratedFieldConfig `yaml:"generated" json:"generated"`

	FirstOf []*BaseField `json:"first_of" yaml:"first_of"`
}

type FormattedFieldConfig struct {
	Template string `yaml:"template" json:"template"`
}

type GeneratedFieldConfig struct {
	UUID             *UUIDGeneratedFieldConfig   `yaml:"uuid" json:"uuid"`
	Static           *StaticGeneratedFieldConfig `yaml:"static" json:"static"`
	Formatted        *FormattedFieldConfig       `json:"formatted" yaml:"formatted"`
	Plugin           *PluginFieldConfig          `yaml:"plugin" json:"plugin"`
	Calculated       *CalculatedConfig           `yaml:"calculated" json:"calculated"`
	File             *FileFieldConfig            `yaml:"file" json:"file"`
	Model            *ModelField                 `yaml:"model" json:"model"`
	FileStorageField *FileStorageField           `json:"file_storage" yaml:"file_storage"`
}

type FileStorageField struct {
	Content string          `json:"content" yaml:"content"`
	Raw     json.RawMessage `yaml:"raw" yaml:"raw"`

	FileName string `json:"file_name" yaml:"file_name"`
	Path     string `json:"path" yaml:"path"`
	Append   bool   `json:"append" yaml:"append"`
}

type FileFieldConfig struct {
	Config *ServerConnectorConfig `yaml:"config" json:"config"`

	Url      string `yaml:"url" json:"url"`
	FileName string `json:"file_name" yaml:"file_name"`
	Path     string `json:"path" yaml:"path"`
}

type CalculatedConfig struct {
	Type       FieldType `yaml:"type" json:"type"`
	Expression string    `yaml:"expression" json:"expression"`
}

type PluginFieldConfig struct {
	Name   string          `json:"name" yaml:"name"`
	Config json.RawMessage `json:"config" yaml:"config"`
}

type ModelField struct {
	// Type of parsing
	ConnectorConfig *ConnectorConfig `yaml:"connector_config" json:"connector_config"`
	// Model of the response
	Model *Model `yaml:"model" json:"model"`

	Type       FieldType `yaml:"type" json:"type"`
	Path       string    `yaml:"path" json:"path"`
	Expression string    `yaml:"expression" json:"expression"`
}

type StaticGeneratedFieldConfig struct {
	Type  FieldType `yaml:"type" json:"type"`
	Value string    `json:"value" yaml:"value"`
}

type UUIDGeneratedFieldConfig struct {
	Regexp string `yaml:"regexp" json:"regexp"`
}

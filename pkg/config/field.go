package config

type FieldType string

const (
	Bool    FieldType = "boolean"
	String  FieldType = "string"
	Int     FieldType = "int"
	Int64   FieldType = "int64"
	Float   FieldType = "float"
	Float64 FieldType = "float64"
)

type Field struct {
	*BaseField
	ObjectConfig *EntryConfig `yaml:"object_config" yaml:"object_config"`
	ArrayConfig  *ArrayConfig `json:"array_config" yaml:"array_config"`
}

type BaseField struct {
	Type FieldType `yaml:"type" json:"type"`
	Path string    `yaml:"path" json:"path"`
}

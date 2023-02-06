package config

type FieldType string

const (
	Bool    FieldType = "boolean"
	String  FieldType = "string"
	Int     FieldType = "int"
	Int64   FieldType = "int64"
	Float   FieldType = "float"
	Float64 FieldType = "float64"
	Object  FieldType = "object"
	Array   FieldType = "array"
)

type Field struct {
	Type         FieldType    `yaml:"type" json:"type"`
	Path         FieldType    `yaml:"path" json:"path"`
	ObjectConfig *EntryConfig `yaml:"object_config" yaml:"object_config"`
	ArrayConfig  *ArrayConfig `json:"array_config" yaml:"array_config"`
}

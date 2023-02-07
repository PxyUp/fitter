package config

type FieldType string

const (
	Null    FieldType = "null"
	Bool    FieldType = "boolean"
	String  FieldType = "string"
	Int     FieldType = "int"
	Int64   FieldType = "int64"
	Float   FieldType = "float"
	Float64 FieldType = "float64"
)

type Field struct {
	*BaseField
	ObjectConfig *ObjectConfig `yaml:"object_config" yaml:"object_config"`
	ArrayConfig  *ArrayConfig  `json:"array_config" yaml:"array_config"`
}

type BaseField struct {
	Type FieldType `yaml:"type" json:"type"`
	Path string    `yaml:"path" json:"path"`

	Generated *GeneratedFieldConfig `yaml:"generated" json:"generated"`
}

type GeneratedFieldConfig struct {
	UUID   *UUIDGeneratedFieldConfig   `yaml:"uuid" json:"uuid"`
	Static *StaticGeneratedFieldConfig `yaml:"static" json:"static"`
}

type StaticGeneratedFieldConfig struct {
	Type  FieldType `yaml:"type" json:"type"`
	Value string    `json:"value" yaml:"value"`
}

type UUIDGeneratedFieldConfig struct {
	Regexp string `yaml:"regexp" json:"regexp"`
}

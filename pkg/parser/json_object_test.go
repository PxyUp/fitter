package parser_test

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestJsonV2Object(t *testing.T) {
	suite.Run(t, new(JsonV2ObjectSuite))
}

type JsonV2ObjectSuite struct {
	suite.Suite
	body   []byte
	parser parser.Parser
}

func (s *JsonV2ObjectSuite) SetupTest() {
	jsonFile, err := os.Open("json_example_object.json")
	require.NoError(s.T(), err)
	defer jsonFile.Close()

	jsonBody, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		require.NoError(s.T(), err)
	}
	s.body = jsonBody
	s.parser = parser.NewJson(s.body, logger.Null)
}

func (s *JsonV2ObjectSuite) Test_Return_BaseField_String() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "eyeColor",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "\"green\"", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_Return_BaseField_Number() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.Int,
			Path: "age",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "27", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_Return_BaseField_Calculated() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.Int,
			Path: "age",
			Generated: &config.GeneratedFieldConfig{
				Calculated: &config.CalculatedConfig{
					Type:       config.Bool,
					Expression: "fRes >= 27",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "true", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_FirstOf() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"title": {
					BaseField: &config.BaseField{
						FirstOf: []*config.BaseField{
							{
								Type: config.String,
								Path: "asdfasfasfasf",
							},
							{
								Type: config.String,
								Path: "name",
							},
						},
					},
				},
				"object": {
					FirstOf: []*config.Field{
						{
							ObjectConfig: &config.ObjectConfig{
								Fields: map[string]*config.Field{
									"title": {
										BaseField: &config.BaseField{
											Type: config.String,
											Path: "asdfasfasfasf",
										},
									},
								},
							},
						},
						{
							ObjectConfig: &config.ObjectConfig{
								Fields: map[string]*config.Field{
									"title": {
										BaseField: &config.BaseField{
											Type: config.String,
											Path: "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"object\": {\"title\": \"Henderson Gonzales\"},\"title\": \"Henderson Gonzales\"}\n", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_StaticArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Type: config.String,
							Path: "gender",
						},
					},
					1: {
						ObjectConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"test": {
									BaseField: &config.BaseField{
										Type: config.String,
										Path: "friends.0.name",
									},
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Type: config.String,
							Path: "gender",
							Generated: &config.GeneratedFieldConfig{
								Calculated: &config.CalculatedConfig{
									Type:       config.Bool,
									Expression: "fIndex == 2",
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"male\",{\"test\": \"Cooley Spence\"}, true]", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_ParseSimpleObject() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"address": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "address",
					},
				},
				"name": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "friends.0.name",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"address\": \"472 Cheever Place, Spelter, New Jersey, 5250\",\"name\": \"Cooley Spence\"}", res.ToJson())
}

func (s *JsonV2ObjectSuite) TestGeneratedField() {
	require.NoError(s.T(), os.Setenv("T_NUMBER", "99"))
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"uuid": {
					BaseField: &config.BaseField{
						Generated: &config.GeneratedFieldConfig{
							UUID: &config.UUIDGeneratedFieldConfig{},
						},
					},
				},
				"name": {
					BaseField: &config.BaseField{
						Generated: &config.GeneratedFieldConfig{
							Static: &config.StaticGeneratedFieldConfig{
								Type:  config.Int,
								Value: "5",
							},
						},
					},
				},
				"array": {
					BaseField: &config.BaseField{
						Generated: &config.GeneratedFieldConfig{
							Static: &config.StaticGeneratedFieldConfig{
								Type:  config.Array,
								Value: "[1,2,4, {{{FromExp=int({{{FromEnv=T_NUMBER}}})}}}]",
							},
						},
					},
				},
				"string": {
					BaseField: &config.BaseField{
						Generated: &config.GeneratedFieldConfig{
							Static: &config.StaticGeneratedFieldConfig{
								Type: config.Array,
								Raw:  []byte("\"{{{FromEnv=T_NUMBER}}}\""),
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(res.ToJson()), &jsonMap)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(jsonMap["uuid"].(string)) > 0)
	assert.Equal(s.T(), float64(5), jsonMap["name"])
	assert.Equal(s.T(), "99", jsonMap["string"])
	assert.Equal(s.T(), []any{float64(1), float64(2), float64(4), float64(99)}, jsonMap["array"])
	require.NoError(s.T(), os.Unsetenv("T_NUMBER"))
}

func (s *JsonV2ObjectSuite) Test_ReturnSimpleArray_Concat() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"prices": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "friends.#.meals|@flatten.#.price",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{
								Type: config.Int,
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"prices\": [292,357,695,315,279,336,594,821,791]}\n", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_ReturnSimpleArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "tags",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.String,
					Path: "",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"tempor\",\"magna\",\"ullamco\",\"Lorem\",\"sunt\",\"irure\",\"et\"]", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_ReturnSimpleArrayOfArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "tags_nested",
			ItemConfig: &config.ObjectConfig{
				ArrayConfig: &config.ArrayConfig{
					ItemConfig: &config.ObjectConfig{
						Field: &config.BaseField{
							Type: config.String,
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[[\"tempor\"],[\"test\"]]\n", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_ReturnNestedArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "friends",
			ItemConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"name": {
						BaseField: &config.BaseField{
							Type: config.String,
							Path: "name",
						},
					},
					"meals": {
						ArrayConfig: &config.ArrayConfig{
							RootPath: "meals",
							ItemConfig: &config.ObjectConfig{
								Fields: map[string]*config.Field{
									"my_price": {
										BaseField: &config.BaseField{
											Type: config.Int,
											Path: "price",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[{\"name\": \"Cooley Spence\",\"meals\": [{\"my_price\": 292},{\"my_price\": 357},{\"my_price\": 695}]},{\"name\": \"Dixie Padilla\",\"meals\": [{\"my_price\": 315},{\"my_price\": 279},{\"my_price\": 336}]},{\"name\": \"Tanisha Kline\",\"meals\": [{\"my_price\": 594},{\"my_price\": 821},{\"my_price\": 791}]}]\n", res.ToJson())
}

func (s *JsonV2ObjectSuite) Test_ParseNestedObject() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"player": {
					ObjectConfig: &config.ObjectConfig{
						Fields: map[string]*config.Field{
							"name": {
								BaseField: &config.BaseField{
									Type: config.String,
									Path: "name",
								},
							},
							"isActive": {
								BaseField: &config.BaseField{
									Type: config.Bool,
									Path: "isActive",
								},
							},
							"null": {
								BaseField: &config.BaseField{
									Type: config.Null,
								},
							},
							"latitude": {
								BaseField: &config.BaseField{
									Type: config.Float,
									Path: "latitude",
								},
							},
							"player_meal": {
								ArrayConfig: &config.ArrayConfig{
									RootPath: "friends.0.meals",
									ItemConfig: &config.ObjectConfig{
										Fields: map[string]*config.Field{
											"my_price": {
												BaseField: &config.BaseField{
													Type: config.Int,
													Path: "price",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"tags": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "tags",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{
								Type: config.String,
							},
						},
					},
				},
				"player_meal": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "friends.0.meals",
						ItemConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"my_price": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Path: "price",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"player_meal\": [{\"my_price\": 292},{\"my_price\": 357},{\"my_price\": 695}],\"player\": {\"latitude\": 44.823498,\"player_meal\": [{\"my_price\": 292},{\"my_price\": 357},{\"my_price\": 695}],\"name\": \"Henderson Gonzales\",\"isActive\": true,\"null\": null},\"tags\": [\"tempor\",\"magna\",\"ullamco\",\"Lorem\",\"sunt\",\"irure\",\"et\"]}\n", res.ToJson())
}

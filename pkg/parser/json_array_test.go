package parser_test

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestJsonV2Array(t *testing.T) {
	suite.Run(t, new(JsonV2ArraySuite))
}

type JsonV2ArraySuite struct {
	suite.Suite
	body   []byte
	parser parser.Parser
}

func (s *JsonV2ArraySuite) SetupTest() {
	jsonFile, err := os.Open("json_example_array.json")
	require.NoError(s.T(), err)
	defer jsonFile.Close()

	jsonBody, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		require.NoError(s.T(), err)
	}
	s.body = jsonBody
	s.parser = parser.NewJson(s.body, logger.Null)
}

func (s *JsonV2ArraySuite) TestFileConnector() {
	body, err := connectors.NewFile(&config.FileConnectorConfig{
		Path: "json_example_array.json",
	}).Get(nil, nil, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.body, body)
}

func (s *JsonV2ArraySuite) Test_ParseSimpleObject() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"address_1": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "0.address",
					},
				},
				"address_2": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "1.address",
					},
				},
				"address_3": {
					BaseField: &config.BaseField{
						Type: config.String,
						Generated: &config.GeneratedFieldConfig{
							Formatted: &config.FormattedFieldConfig{
								Template: "{{{FromInput=.}}}",
							},
						},
					},
				},
			},
		},
	}, builder.Int(55))
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"address_1\": \"433 Bennet Court, Manchester, Rhode Island, 6346\",\"address_2\": \"472 Cheever Place, Spelter, New Jersey, 5250\", \"address_3\": \"55\"}\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnNestedArray_Concat() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "#.friends|@flatten.#.meals|@flatten.#.price",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.Int,
					Path: "",
				},
			},
		},
	}, nil)

	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[215,692,257,623,172,567,960,924,857,292,357,695,315,279,336,594,821,791]\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnSimpleArray_Concat() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "#.tags|@flatten",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.String,
					Path: "",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"veniam\",\"nostrud\",\"elit\",\"consequat\",\"mollit\",\"pariatur\",\"proident\",\"tempor\",\"magna\",\"ullamco\",\"Lorem\",\"sunt\",\"irure\",\"et\"]\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnSimpleArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.String,
					Path: "email",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"nolanlester@qimonk.com\",\"hendersongonzales@megall.com\"]", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnSimpleArrayReverse() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "",
			Reverse:  true,
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.String,
					Path: "email",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"hendersongonzales@megall.com\",\"nolanlester@qimonk.com\"]", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnSimpleArray_Index() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.String,
					Path: "email",
					Generated: &config.GeneratedFieldConfig{
						Formatted: &config.FormattedFieldConfig{
							Template: "EMAIL: {PL} INDEX: {INDEX}",
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"EMAIL: nolanlester@qimonk.com INDEX: 0\",\"EMAIL: hendersongonzales@megall.com INDEX: 1\"]\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnSimpleArrayOfArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "",
			ItemConfig: &config.ObjectConfig{
				ArrayConfig: &config.ArrayConfig{
					RootPath: "tags",
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
	assert.JSONEq(s.T(), "[[\"veniam\",\"nostrud\",\"elit\",\"consequat\",\"mollit\",\"pariatur\",\"proident\"],[\"tempor\",\"magna\",\"ullamco\",\"Lorem\",\"sunt\",\"irure\",\"et\"]]\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ReturnNestedArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath: "",
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
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[{\"name\": \"Nolan Lester\",\"meals\": [{\"my_price\": 215},{\"my_price\": 692},{\"my_price\": 257}]},{\"name\": \"Henderson Gonzales\",\"meals\": [{\"my_price\": 292},{\"my_price\": 357},{\"my_price\": 695}]}]\n", res.ToJson())
}

func (s *JsonV2ArraySuite) Test_ParseNestedObject() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"player": {
					ObjectConfig: &config.ObjectConfig{
						Fields: map[string]*config.Field{
							"name": {
								BaseField: &config.BaseField{
									Type: config.String,
									Path: "0.name",
								},
							},
							"isActive": {
								BaseField: &config.BaseField{
									Type: config.Bool,
									Path: "1.isActive",
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
									Path: "1.latitude",
								},
							},
							"player_meal": {
								ArrayConfig: &config.ArrayConfig{
									RootPath: "1.friends.1.meals",
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
						RootPath: "1.tags",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{
								Type: config.String,
								Path: "",
							},
						},
					},
				},
				"player_meal": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "1.friends.0.meals",
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
	assert.JSONEq(s.T(), "{\"player\": {\"null\": null,\"latitude\": 44.823498,\"player_meal\": [{\"my_price\": 315},{\"my_price\": 279},{\"my_price\": 336}],\"name\": \"Nolan Lester\",\"isActive\": false},\"tags\": [\"tempor\",\"magna\",\"ullamco\",\"Lorem\",\"sunt\",\"irure\",\"et\"],\"player_meal\": [{\"my_price\": 292},{\"my_price\": 357},{\"my_price\": 695}]}\n", res.ToJson())
}

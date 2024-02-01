package parser_test

import (
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

func TestNewXML(t *testing.T) {
	suite.Run(t, new(NewXMLSuite))
}

type NewXMLSuite struct {
	suite.Suite
	body   []byte
	parser parser.Parser
}

func (s *NewXMLSuite) SetupTest() {
	jsonFile, err := os.Open("index.xml")
	require.NoError(s.T(), err)
	defer jsonFile.Close()

	jsonBody, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		require.NoError(s.T(), err)
	}
	s.body = jsonBody
	s.parser = parser.NewXML(s.body, logger.Null)
}

func (s *NewXMLSuite) Test_Return_BaseField_String() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "/breakfast_menu/food/name",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "\"Belgian Waffles\"", res.ToJson())
}

func (s *NewXMLSuite) Test_Return_BaseField_Calculated() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.Int,
			Path: "/breakfast_menu/food/calories",
			Generated: &config.GeneratedFieldConfig{
				Calculated: &config.CalculatedConfig{
					Type:       config.Int,
					Expression: "fRes + 2",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "652", res.ToJson())
}

func (s *NewXMLSuite) Test_Return_BaseField_Number() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.Int,
			Path: "/breakfast_menu/food/calories",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "650", res.ToJson())
}
func (s *NewXMLSuite) Test_StaticArray() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Type: config.String,
							Path: "/breakfast_menu/food/name",
						},
					},
					1: {
						ObjectConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"intro": {
									BaseField: &config.BaseField{
										Type: config.String,
										Path: "/breakfast_menu/food[2]/description",
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
	assert.JSONEq(s.T(), "[\"Belgian Waffles\",{\"intro\": \"Light Belgian waffles covered with strawberries and whipped cream\"}]", res.ToJson())
}

func (s *NewXMLSuite) Test_ReturnSimpleArray() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"menu": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "/breakfast_menu/food",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{
								Type: config.String,
								Path: "name",
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"menu\": [\"Belgian Waffles\",\"Strawberry Belgian Waffles\",\"Berry-Berry Belgian Waffles\",\"French Toast\",\"Homestyle Breakfast\"]}\n", res.ToJson())
}

func (s *NewXMLSuite) Test_ReturnSimpleArrayOfArray_Index() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"menu": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "/breakfast_menu/food",
						ItemConfig: &config.ObjectConfig{
							ArrayConfig: &config.ArrayConfig{
								RootPath: "/cla/name",
								ItemConfig: &config.ObjectConfig{
									Field: &config.BaseField{
										Type: config.String,
										Generated: &config.GeneratedFieldConfig{
											Formatted: &config.FormattedFieldConfig{
												Template: "{PL} {INDEX}",
											},
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
	assert.JSONEq(s.T(), "{\"menu\": [[\"1 0\",\"4 1\"],[\"1 0\",\"3 1\"],[\"1 0\",\"5 1\"],[\"1 0\",\"6 1\"],[\"1 0\",\"7 1\"]]}\n", res.ToJson())
}

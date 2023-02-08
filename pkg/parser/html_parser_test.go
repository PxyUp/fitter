package parser_test

import (
	"encoding/json"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewHtmlObject(t *testing.T) {
	suite.Run(t, new(HTMLParserArraySuite))
}

type HTMLParserArraySuite struct {
	suite.Suite
	body   []byte
	parser parser.Parser
}

func (s *HTMLParserArraySuite) SetupTest() {
	jsonFile, err := os.Open("index.html")
	require.NoError(s.T(), err)
	defer jsonFile.Close()

	jsonBody, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		require.NoError(s.T(), err)
	}
	s.body = jsonBody
	s.parser = parser.NewHTML(s.body)
}

func (s *HTMLParserArraySuite) Test_ParseSimpleObject() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"title": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "h2",
					},
				},
				"intro": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: ".w3-main .intro",
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"title\": \"Tutorials\",\"intro\": \"HTML headings are titles or subtitles that you want to display on a webpage.\"}", res.Raw)
}

func (s *HTMLParserArraySuite) TestGeneratedField() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ObjectModel,
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
			},
		},
	})
	assert.NoError(s.T(), err)
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(res.Raw), &jsonMap)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(jsonMap["uuid"].(string)) > 0)
	assert.Equal(s.T(), float64(5), jsonMap["name"])
}

func (s *HTMLParserArraySuite) Test_ReturnSimpleArray() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"menu": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "#topnav .w3-bar.w3-left a",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{
								Type: config.String,
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"menu\": [\"\",\"\",\"HTML\",\"CSS\",\"JAVASCRIPT\",\"SQL\",\"PYTHON\",\"JAVA\",\"PHP\",\"BOOTSTRAP\",\"HOW TO\",\"W3.CSS\",\"C\",\"C++\",\"C#\",\"REACT\",\"R\",\"JQUERY\",\"DJANGO\",\"TYPESCRIPT\",\"NODEJS\",\"MYSQL\",\"\uE802\",\"\uE801\",\"\uE80B\"]}\n", res.Raw)
}

func (s *HTMLParserArraySuite) Test_ReturnSimpleArrayOfArray() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"menu": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "#test_array div",
						ItemConfig: &config.ObjectConfig{
							ArrayConfig: &config.ArrayConfig{
								RootPath: "a",
								ItemConfig: &config.ObjectConfig{
									Field: &config.BaseField{
										Type: config.String,
									},
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"menu\": [[\"TEST_1\",\"TEST_2\"],[\"TEST_3\",\"TEST_4\"]]}\n", res.Raw)
}

func (s *HTMLParserArraySuite) Test_ReturnNestedArray() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ArrayModel,
		ArrayConfig: &config.ArrayConfig{
			RootPath: "#nav_tutorials .w3-col",
			ItemConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"name": {
						BaseField: &config.BaseField{
							Type: config.String,
							Path: "h3",
						},
					},
					"tutorials": {
						ArrayConfig: &config.ArrayConfig{
							RootPath: "a",
							ItemConfig: &config.ObjectConfig{
								Fields: map[string]*config.Field{
									"name": {
										BaseField: &config.BaseField{
											Type: config.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[{\"name\": \"HTML and CSS\",\"tutorials\": [{\"name\": \"Learn HTML\"},{\"name\": \"Learn CSS\"},{\"name\": \"Learn RWD\"},{\"name\": \"Learn Bootstrap\"},{\"name\": \"Learn W3.CSS\"},{\"name\": \"Learn Colors\"},{\"name\": \"Learn Icons\"},{\"name\": \"Learn Graphics\"},{\"name\": \"Learn SVG\"},{\"name\": \"Learn Canvas\"},{\"name\": \"Learn How To\"},{\"name\": \"Learn Sass\"},{\"name\": \"Learn AI\"},{\"name\": \"Learn Machine Learning\"},{\"name\": \"Learn Data Science\"},{\"name\": \"Learn NumPy\"},{\"name\": \"Learn Pandas\"},{\"name\": \"Learn SciPy\"},{\"name\": \"Learn Matplotlib\"},{\"name\": \"Learn Statistics\"},{\"name\": \"Learn Excel\"},{\"name\": \"Learn XML\"},{\"name\": \"Learn XML AJAX\"},{\"name\": \"Learn XML DOM\"},{\"name\": \"Learn XML DTD\"},{\"name\": \"Learn XML Schema\"},{\"name\": \"Learn XSLT\"},{\"name\": \"Learn XPath\"},{\"name\": \"Learn XQuery\"}]},{\"name\": \"JavaScript\",\"tutorials\": [{\"name\": \"Learn JavaScript\"},{\"name\": \"Learn jQuery\"},{\"name\": \"Learn React\"},{\"name\": \"Learn AngularJS\"},{\"name\": \"Learn JSON\"},{\"name\": \"Learn AJAX\"},{\"name\": \"Learn AppML\"},{\"name\": \"Learn W3.JS\"},{\"name\": \"Learn Python\"},{\"name\": \"Learn Java\"},{\"name\": \"Learn C\"},{\"name\": \"Learn C++\"},{\"name\": \"Learn C#\"},{\"name\": \"Learn R\"},{\"name\": \"Learn Kotlin\"},{\"name\": \"Learn Go\"},{\"name\": \"Learn Django\"},{\"name\": \"Learn TypeScript\"}]},{\"name\": \"Server Side\",\"tutorials\": [{\"name\": \"Learn SQL\"},{\"name\": \"Learn MySQL\"},{\"name\": \"Learn PHP\"},{\"name\": \"Learn ASP\"},{\"name\": \"Learn Node.js\"},{\"name\": \"Learn Raspberry Pi\"},{\"name\": \"Learn Git\"},{\"name\": \"Learn MongoDB\"},{\"name\": \"Learn AWS Cloud\"},{\"name\": \"Create a Website NEW\"},{\"name\": \"Where To Start\"},{\"name\": \"Web Templates\"},{\"name\": \"Web Statistics\"},{\"name\": \"Web Certificates\"},{\"name\": \"Web Development\"},{\"name\": \"Code Editor\"},{\"name\": \"Test Your Typing Speed\"},{\"name\": \"Play a Code Game\"},{\"name\": \"Cyber Security\"},{\"name\": \"Accessibility\"},{\"name\": \"Join our Newsletter\"}]},{\"name\": \"Data Analytics\",\"tutorials\": [{\"name\": \"Learn AI\"},{\"name\": \"Learn Machine Learning\"},{\"name\": \"Learn Data Science\"},{\"name\": \"Learn NumPy\"},{\"name\": \"Learn Pandas\"},{\"name\": \"Learn SciPy\"},{\"name\": \"Learn Matplotlib\"},{\"name\": \"Learn Statistics\"},{\"name\": \"Learn Excel\"},{\"name\": \"Learn Google Sheets\"},{\"name\": \"Learn XML\"},{\"name\": \"Learn XML AJAX\"},{\"name\": \"Learn XML DOM\"},{\"name\": \"Learn XML DTD\"},{\"name\": \"Learn XML Schema\"},{\"name\": \"Learn XSLT\"},{\"name\": \"Learn XPath\"},{\"name\": \"Learn XQuery\"}]}]\n", res.Raw)
}

func (s *HTMLParserArraySuite) Test_ParseNestedObject() {
	res, err := s.parser.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"player": {
					ObjectConfig: &config.ObjectConfig{
						Fields: map[string]*config.Field{
							"name": {
								BaseField: &config.BaseField{
									Type: config.String,
									Path: "title",
								},
							},
							"isActive": {
								BaseField: &config.BaseField{
									Type: config.Bool,
									Path: "body > p.is_active",
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
									Path: "p.float",
								},
							},
							"player_meal": {
								ArrayConfig: &config.ArrayConfig{
									RootPath: "#myAccordion .w3-container a",
									ItemConfig: &config.ObjectConfig{
										Fields: map[string]*config.Field{
											"my_price": {
												BaseField: &config.BaseField{
													Type: config.String,
													Path: "b",
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
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"player\": {\"latitude\": 3.120000,\"player_meal\": [{\"my_price\": \"first\"},{\"my_price\": \"second\"},{\"my_price\": null},{\"my_price\": null},{\"my_price\": null},{\"my_price\": null},{\"my_price\": null},{\"my_price\": null},{\"my_price\": null}],\"name\": \"HTML Headings\",\"isActive\": true,\"null\": null}}\n", res.Raw)
}

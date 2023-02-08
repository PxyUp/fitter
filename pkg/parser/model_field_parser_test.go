package parser_test

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	jsonBodyObject = []byte(`{"postal_codes": [10101, 10102]}`)
	jsonBodyArray  = []byte(`[10101, 10102]`)
	htmlBody       = []byte(`<html><body><code>10101</code><code>10102</code></body></html>`)
)

type ModelFieldParserSuite struct {
	suite.Suite
	jsonParserObject parser.Parser
	jsonParserArray  parser.Parser
	htmlParser       parser.Parser

	server *httptest.Server
}

func TestModelFieldParserSuite(t *testing.T) {
	suite.Run(t, new(ModelFieldParserSuite))
}

type testHandler struct {
}

func (t *testHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/html") {
		writer.Header().Set("Content-Type", "text/html")
	}
	if strings.HasPrefix(request.URL.Path, "/html/10101") {
		fmt.Fprintf(writer, `<html><title>Here 10101</title></html>`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/html/10102") {
		fmt.Fprintf(writer, `<html><title>Here 10102</title></html>`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/html/10103") {
		fmt.Fprintf(writer, `<html><title>Here 10103</title></html>`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/html/10104") {
		fmt.Fprintf(writer, `<html><title>Here 10104</title></html>`)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if strings.HasPrefix(request.URL.Path, "/10101") {
		fmt.Fprintf(writer, `{"population": 1010110101}`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/10102") {
		fmt.Fprintf(writer, `{"population": 1010210102}`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/10100") {
		fmt.Fprintf(writer, `{"population": 1010010100}`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/10103") {
		fmt.Fprintf(writer, `{"population": 1010310103}`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/neighbour/10101") {
		fmt.Fprintf(writer, `{"neighbour": [10102, 10100]}`)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/neighbour/10102") {
		fmt.Fprintf(writer, `{"neighbour": [10101, 10103]}`)
		return
	}
}

func (s *ModelFieldParserSuite) SetupTest() {
	s.jsonParserObject = parser.NewJson(jsonBodyObject)
	s.jsonParserArray = parser.NewJson(jsonBodyArray)
	s.htmlParser = parser.NewHTML(htmlBody)
	s.server = httptest.NewServer(&testHandler{})
}

func (s *ModelFieldParserSuite) TearDownTest() {
	s.server.Close()
}

func (s *ModelFieldParserSuite) TestJSONObject_ModelFieldFetching() {
	res, err := s.jsonParserObject.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"codes": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "postal_codes",
						ItemConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"code": {
									BaseField: &config.BaseField{
										Type: config.Int,
									},
								},
								"title": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.String),
												Path: "title",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"title": {
																BaseField: &config.BaseField{
																	Type: config.String,
																	Path: "title",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.HTML,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/html", s.server.URL) + "/%s",
													},
												},
											},
										},
									},
								},
								"neighbour": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.Array,
												Path: "neighbour",
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/neighbour", s.server.URL) + "/%s",
													},
												},
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"neighbour": {
																ArrayConfig: &config.ArrayConfig{
																	RootPath: "neighbour",
																	ItemConfig: &config.ObjectConfig{
																		Fields: map[string]*config.Field{
																			"code": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																				},
																			},
																			"population": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																					Generated: &config.GeneratedFieldConfig{
																						Model: &config.ModelField{
																							Type: config.GeneratedFieldType(config.Int),
																							Path: "pop",
																							Model: &config.Model{
																								Type: config.ObjectModel,
																								ObjectConfig: &config.ObjectConfig{
																									Fields: map[string]*config.Field{
																										"pop": {
																											BaseField: &config.BaseField{
																												Type: config.Int,
																												Path: "population",
																											},
																										},
																									},
																								},
																							},
																							ConnectorConfig: &config.ConnectorConfig{
																								ConnectorType: config.Server,
																								ResponseType:  config.Json,
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
																									Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
													},
												},
											},
										},
									},
								},
								"population": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.Int),
												Path: "pop",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"pop": {
																BaseField: &config.BaseField{
																	Type: config.Int,
																	Path: "population",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"codes\": [{\"neighbour\": [{\"code\": 10102,\"population\": 1010210102},{\"population\": 1010010100,\"code\": 10100}],\"population\": 1010110101,\"code\": 10101,\"title\": \"Here 10101\"},{\"population\": 1010210102,\"code\": 10102,\"title\": \"Here 10102\",\"neighbour\": [{\"population\": 1010110101,\"code\": 10101},{\"code\": 10103,\"population\": 1010310103}]}]}\n", res.ToJson())
}

func (s *ModelFieldParserSuite) TestJSONArray_ModelFieldFetching() {
	res, err := s.jsonParserArray.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"codes": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "@flatten",
						ItemConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"code": {
									BaseField: &config.BaseField{
										Type: config.Int,
									},
								},
								"title": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.String),
												Path: "title",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"title": {
																BaseField: &config.BaseField{
																	Type: config.String,
																	Path: "title",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.HTML,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/html", s.server.URL) + "/%s",
													},
												},
											},
										},
									},
								},
								"neighbour": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.Array,
												Path: "neighbour",
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/neighbour", s.server.URL) + "/%s",
													},
												},
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"neighbour": {
																ArrayConfig: &config.ArrayConfig{
																	RootPath: "neighbour",
																	ItemConfig: &config.ObjectConfig{
																		Fields: map[string]*config.Field{
																			"code": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																				},
																			},
																			"population": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																					Generated: &config.GeneratedFieldConfig{
																						Model: &config.ModelField{
																							Type: config.GeneratedFieldType(config.Int),
																							Path: "pop",
																							Model: &config.Model{
																								Type: config.ObjectModel,
																								ObjectConfig: &config.ObjectConfig{
																									Fields: map[string]*config.Field{
																										"pop": {
																											BaseField: &config.BaseField{
																												Type: config.Int,
																												Path: "population",
																											},
																										},
																									},
																								},
																							},
																							ConnectorConfig: &config.ConnectorConfig{
																								ConnectorType: config.Server,
																								ResponseType:  config.Json,
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
																									Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
													},
												},
											},
										},
									},
								},
								"population": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.Int),
												Path: "pop",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"pop": {
																BaseField: &config.BaseField{
																	Type: config.Int,
																	Path: "population",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"codes\": [{\"neighbour\": [{\"code\": 10102,\"population\": 1010210102},{\"population\": 1010010100,\"code\": 10100}],\"population\": 1010110101,\"code\": 10101,\"title\": \"Here 10101\"},{\"population\": 1010210102,\"code\": 10102,\"title\": \"Here 10102\",\"neighbour\": [{\"population\": 1010110101,\"code\": 10101},{\"code\": 10103,\"population\": 1010310103}]}]}\n", res.ToJson())
}

func (s *ModelFieldParserSuite) TestHTTP_ModelFieldFetching() {
	res, err := s.htmlParser.Parse(&config.Model{
		Type: config.ObjectModel,
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"codes": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "code",
						ItemConfig: &config.ObjectConfig{
							Fields: map[string]*config.Field{
								"code": {
									BaseField: &config.BaseField{
										Type: config.Int,
									},
								},
								"title": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.String),
												Path: "title",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"title": {
																BaseField: &config.BaseField{
																	Type: config.String,
																	Path: "title",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.HTML,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/html", s.server.URL) + "/%s",
													},
												},
											},
										},
									},
								},
								"neighbour": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.Array,
												Path: "neighbour",
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s/neighbour", s.server.URL) + "/%s",
													},
												},
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"neighbour": {
																ArrayConfig: &config.ArrayConfig{
																	RootPath: "neighbour",
																	ItemConfig: &config.ObjectConfig{
																		Fields: map[string]*config.Field{
																			"code": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																				},
																			},
																			"population": {
																				BaseField: &config.BaseField{
																					Type: config.Int,
																					Generated: &config.GeneratedFieldConfig{
																						Model: &config.ModelField{
																							Type: config.GeneratedFieldType(config.Int),
																							Path: "pop",
																							Model: &config.Model{
																								Type: config.ObjectModel,
																								ObjectConfig: &config.ObjectConfig{
																									Fields: map[string]*config.Field{
																										"pop": {
																											BaseField: &config.BaseField{
																												Type: config.Int,
																												Path: "population",
																											},
																										},
																									},
																								},
																							},
																							ConnectorConfig: &config.ConnectorConfig{
																								ConnectorType: config.Server,
																								ResponseType:  config.Json,
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
																									Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
													},
												},
											},
										},
									},
								},
								"population": {
									BaseField: &config.BaseField{
										Type: config.Int,
										Generated: &config.GeneratedFieldConfig{
											Model: &config.ModelField{
												Type: config.GeneratedFieldType(config.Int),
												Path: "pop",
												Model: &config.Model{
													Type: config.ObjectModel,
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"pop": {
																BaseField: &config.BaseField{
																	Type: config.Int,
																	Path: "population",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ConnectorType: config.Server,
													ResponseType:  config.Json,
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
														Url:    fmt.Sprintf("%s", s.server.URL) + "/%s",
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
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "{\"codes\": [{\"neighbour\": [{\"code\": 10102,\"population\": 1010210102},{\"population\": 1010010100,\"code\": 10100}],\"population\": 1010110101,\"code\": 10101,\"title\": \"Here 10101\"},{\"population\": 1010210102,\"code\": 10102,\"title\": \"Here 10102\",\"neighbour\": [{\"population\": 1010110101,\"code\": 10101},{\"code\": 10103,\"population\": 1010310103}]}]}\n", res.ToJson())
}

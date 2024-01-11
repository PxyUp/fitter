package parser_test

import (
	"bytes"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

var (
	jsonDatesObject = []byte(`[{"from":"2023-07-10","to":"2023-07-14"},{"from":"2023-07-11","to":"2023-07-15"},{"from":"2023-07-07","to":"2023-07-11"},{"from":"2023-07-06","to":"2023-07-12"}]`)
	jsonBodyObject  = []byte(`{"postal_codes": [10101, 10102]}`)
	jsonBodyArray   = []byte(`[10101, 10102]`)
	htmlBody        = []byte(`<html><body><code>10101</code><code>10102</code></body></html>`)
)

type ModelFieldParserSuite struct {
	suite.Suite
	jsonParserObject parser.Parser
	jsonParserArray  parser.Parser
	htmlParser       parser.Parser
	xpathParser      parser.Parser
	jsonDatesParser  parser.Parser

	server             *httptest.Server
	tmpFilePath        string
	notExistingTempDir string
}

func TestModelFieldParserSuite(t *testing.T) {
	suite.Run(t, new(ModelFieldParserSuite))
}

type testHandler struct {
}

var (
	fileName   = "foo.pdf"
	fileBuffer = []byte{1, 2, 3, 4}
)

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

	if strings.HasPrefix(request.URL.Path, "/file") {
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		_, _ = writer.Write(fileBuffer)
	}
}

func (s *ModelFieldParserSuite) SetupSuite() {
	s.tmpFilePath = os.TempDir()
	s.notExistingTempDir = path.Join(os.TempDir(), uuid.New().String())
}

func (s *ModelFieldParserSuite) SetupTest() {
	s.jsonParserObject = parser.JsonFactory(jsonBodyObject, logger.Null)
	s.jsonParserArray = parser.JsonFactory(jsonBodyArray, logger.Null)
	s.htmlParser = parser.HTMLFactory(htmlBody, logger.Null)
	s.jsonDatesParser = parser.JsonFactory(jsonDatesObject, logger.Null)
	s.xpathParser = parser.XPathFactory(htmlBody, logger.Null)
	s.server = httptest.NewServer(&testHandler{})
}

func (s *ModelFieldParserSuite) TearDownTest() {
	s.server.Close()
}

func (s *ModelFieldParserSuite) TearDownSuite() {
	s.server.Close()
	err := os.Remove(path.Join(s.tmpFilePath, fileName))
	require.NoError(s.T(), err)
	err = os.Remove(path.Join(s.notExistingTempDir, fileName))
	require.NoError(s.T(), err)
}

func (s *ModelFieldParserSuite) TestFile() {
	res, err := s.jsonDatesParser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "",
			Generated: &config.GeneratedFieldConfig{
				File: &config.FileFieldConfig{
					Path: s.tmpFilePath,
					Url:  fmt.Sprintf("%s/file", s.server.URL),
					Config: &config.ServerConnectorConfig{
						Method: http.MethodGet,
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), fmt.Sprintf(`"%s"`, path.Join(s.tmpFilePath, fileName)), res.ToJson())
	assert.FileExists(s.T(), path.Join(s.tmpFilePath, fileName))
	file, err := os.OpenFile(path.Join(s.tmpFilePath, fileName), os.O_RDWR, 0755)
	require.NoError(s.T(), err)
	resp, err := io.ReadAll(file)
	require.NoError(s.T(), err)
	assert.True(s.T(), bytes.Equal(fileBuffer, resp))
}

func (s *ModelFieldParserSuite) TestFile_NotExistingDir() {
	res, err := s.jsonDatesParser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "",
			Generated: &config.GeneratedFieldConfig{
				File: &config.FileFieldConfig{
					Path: s.notExistingTempDir,
					Url:  fmt.Sprintf("%s/file", s.server.URL),
					Config: &config.ServerConnectorConfig{
						Method: http.MethodGet,
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), fmt.Sprintf(`"%s"`, path.Join(s.notExistingTempDir, fileName)), res.ToJson())
	assert.FileExists(s.T(), path.Join(s.notExistingTempDir, fileName))
	file, err := os.OpenFile(path.Join(s.notExistingTempDir, fileName), os.O_RDWR, 0755)
	require.NoError(s.T(), err)
	resp, err := io.ReadAll(file)
	require.NoError(s.T(), err)
	assert.True(s.T(), bytes.Equal(fileBuffer, resp))
}

func (s *ModelFieldParserSuite) TestJSONObject_ModelField_Formating() {
	res, err := s.jsonDatesParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{
					Type: config.Object,
					Generated: &config.GeneratedFieldConfig{
						Formatted: &config.FormattedFieldConfig{
							Template: "From: {{{from}}} To: {{{to}}}",
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `["From: 2023-07-10 To: 2023-07-14","From: 2023-07-11 To: 2023-07-15","From: 2023-07-07 To: 2023-07-11","From: 2023-07-06 To: 2023-07-12"]`, res.ToJson())
}

func (s *ModelFieldParserSuite) TestJSONObject_ModelFieldFetching() {
	res, err := s.jsonParserObject.Parse(&config.Model{
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
												Type: config.String,
												Path: "title",
												Model: &config.Model{
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
													ResponseType: config.HTML,
													Url:          fmt.Sprintf("%s/html", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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
													ResponseType: config.Json,
													Url:          fmt.Sprintf("%s/neighbour", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
													},
												},
												Model: &config.Model{
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
																							Type: config.Int,
																							Path: "pop",
																							Model: &config.Model{
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
																								ResponseType: config.Json,
																								Url:          s.server.URL + "/{PL}",
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
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
												Type: config.Int,
												Path: "pop",
												Model: &config.Model{
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
													ResponseType: config.Json,
													Url:          s.server.URL + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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
												Type: config.String,
												Path: "title",
												Model: &config.Model{
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
													ResponseType: config.HTML,
													Url:          fmt.Sprintf("%s/html", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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
													ResponseType: config.Json,
													Url:          fmt.Sprintf("%s/neighbour", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
													},
												},
												Model: &config.Model{
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
																							Type: config.Int,
																							Path: "pop",
																							Model: &config.Model{
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
																								ResponseType: config.Json,
																								Url:          s.server.URL + "/{PL}",
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
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
												Type: config.Int,
												Path: "pop",
												Model: &config.Model{
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
													ResponseType: config.Json,
													Url:          s.server.URL + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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

func (s *ModelFieldParserSuite) TestHTML_ModelFieldFetching() {
	res, err := s.htmlParser.Parse(&config.Model{
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
												Type: config.String,
												Path: "title",
												Model: &config.Model{
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
													ResponseType: config.HTML,
													Url:          fmt.Sprintf("%s/html", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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
													ResponseType: config.Json,
													Url:          fmt.Sprintf("%s/neighbour", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
													},
												},
												Model: &config.Model{
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
																							Type: config.Int,
																							Path: "pop",
																							Model: &config.Model{
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
																								ResponseType: config.Json,
																								Url:          s.server.URL + "/{PL}",
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
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
												Type: config.Int,
												Path: "pop",
												Model: &config.Model{
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
													ResponseType: config.Json,
													Url:          s.server.URL + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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

func (s *ModelFieldParserSuite) TestXPath_ModelFieldFetching() {
	res, err := s.xpathParser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"codes": {
					ArrayConfig: &config.ArrayConfig{
						RootPath: "//code",
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
												Type: config.String,
												Path: "title",
												Model: &config.Model{
													ObjectConfig: &config.ObjectConfig{
														Fields: map[string]*config.Field{
															"title": {
																BaseField: &config.BaseField{
																	Type: config.String,
																	Path: "//title",
																},
															},
														},
													},
												},
												ConnectorConfig: &config.ConnectorConfig{
													ResponseType: config.XPath,
													Url:          fmt.Sprintf("%s/html", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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
													ResponseType: config.Json,
													Url:          fmt.Sprintf("%s/neighbour", s.server.URL) + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
													},
												},
												Model: &config.Model{
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
																							Type: config.Int,
																							Path: "pop",
																							Model: &config.Model{
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
																								ResponseType: config.Json,
																								Url:          s.server.URL + "/{PL}",
																								ServerConfig: &config.ServerConnectorConfig{
																									Method: "GET",
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
												Type: config.Int,
												Path: "pop",
												Model: &config.Model{
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
													ResponseType: config.Json,
													Url:          s.server.URL + "/{PL}",
													ServerConfig: &config.ServerConnectorConfig{
														Method: "GET",
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

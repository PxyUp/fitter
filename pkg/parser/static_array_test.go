package parser_test

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type StaticArraySuite struct {
	suite.Suite
	jsonParser  parser.Parser
	xpathParser parser.Parser
	htmlParser  parser.Parser
}

func TestStaticArraySuite(t *testing.T) {
	suite.Run(t, new(StaticArraySuite))
}

func (s *StaticArraySuite) SetupTest() {
	s.jsonParser = parser.JsonFactory(jsonBodyObject, logger.Null)
	s.xpathParser = parser.XPathFactory(htmlBody, logger.Null)
	s.htmlParser = parser.HTMLFactory(htmlBody, logger.Null)
}

func (s *StaticArraySuite) Test_JSON_Custom_Length() {
	res, err := s.jsonParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
				Length: 3,
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",null,null]", res.Raw)
}

func (s *StaticArraySuite) Test_HTML_Custom_Length() {
	res, err := s.htmlParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
				Length: 3,
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",null,null]", res.Raw)
}

func (s *StaticArraySuite) Test_XPATH_Custom_Length() {
	res, err := s.xpathParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
				Length: 3,
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",null,null]", res.Raw)
}

func (s *StaticArraySuite) Test_JSON() {
	res, err := s.jsonParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",\"Page: 1\",\"Page: 2\"]\n", res.Raw)
}

func (s *StaticArraySuite) Test_HTML() {
	res, err := s.htmlParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",\"Page: 1\",\"Page: 2\"]\n", res.Raw)
}

func (s *StaticArraySuite) Test_XPATH() {
	res, err := s.xpathParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 0\",\"Page: 1\",\"Page: 2\"]\n", res.Raw)
}

func (s *StaticArraySuite) Test_JSON_HUMAN() {
	res, err := s.jsonParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 1\",\"Page: 2\",\"Page: 3\"]\n", res.Raw)
}

func (s *StaticArraySuite) Test_HTML_HUMAN() {
	res, err := s.htmlParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 1\",\"Page: 2\",\"Page: 3\"]\n", res.Raw)
}

func (s *StaticArraySuite) Test_XPATH_HUMAN() {
	res, err := s.xpathParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					1: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
					2: {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								Formatted: &config.FormattedFieldConfig{
									Template: "Page: {HUMAN_INDEX}",
								},
							},
						},
					},
				},
			},
		},
	})
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), "[\"Page: 1\",\"Page: 2\",\"Page: 3\"]\n", res.Raw)
}

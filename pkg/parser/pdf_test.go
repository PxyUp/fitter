package parser_test

import (
	"os"
	"strings"
	"testing"

	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

func TestNewPDF(t *testing.T) {
	suite.Run(t, new(NewPDFSuite))
}

type NewPDFSuite struct {
	suite.Suite
	parser parser.Parser
}

func (s *NewPDFSuite) SetupTest() {
	body, err := os.ReadFile("sample.pdf")
	require.NoError(s.T(), err)
	s.parser = parser.NewPDF(body, logger.Null)
}

func (s *NewPDFSuite) Test_Return_BaseField_Text() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "text",
		},
	}, nil)
	assert.NoError(s.T(), err)
	text := gjson.Parse(res.ToJson()).String()
	assert.True(s.T(), strings.Contains(text, "Sample PDF"))
	assert.True(s.T(), strings.Contains(text, "Lorem ipsum dolor sit amet"))
}

func (s *NewPDFSuite) Test_Return_BaseField_Page() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "pages.0",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.True(s.T(), strings.Contains(gjson.Parse(res.ToJson()).String(), "Sample PDF"))
}

func (s *NewPDFSuite) Test_Return_ObjectField() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"total_pages": {
					BaseField: &config.BaseField{
						Type: config.Int,
						Path: "total_pages",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"total_pages": 1}`, res.ToJson())
}

func (s *NewPDFSuite) Test_InvalidPDF_ReturnsNull() {
	invalidParser := parser.NewPDF([]byte("not a pdf at all"), logger.Null)
	res, err := invalidParser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type: config.String,
			Path: "text",
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `""`, res.ToJson())
}

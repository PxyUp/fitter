package utils_test

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/PxyUp/fitter/pkg/references"
	"github.com/PxyUp/fitter/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type TestFormatterSuite struct {
	suite.Suite
}

func TestRun(t *testing.T) {
	suite.Run(t, new(TestFormatterSuite))
}

func (s *TestFormatterSuite) TestInvalidValue() {
	assert.Equal(s.T(), "", utils.Format("{{{asfasf}}}", nil, nil, nil))
	assert.Equal(s.T(), "", utils.Format("{{{asfasf", nil, nil, nil))
	assert.Equal(s.T(), "", utils.Format("{{{FromEnv=", nil, nil, nil))
	assert.Equal(s.T(), "FromEnv=}}}test", utils.Format("FromEnv=}}}test", nil, nil, nil))
	assert.Equal(s.T(), "FromEnv=}}}testFromEnv=}}}test", utils.Format("FromEnv=}}}testFromEnv=}}}test", nil, nil, nil))
}

func (s *TestFormatterSuite) TestDeepFormatter() {
	assert.Equal(s.T(), "testhello", utils.Format("{{{FromExp=\"{{{FromEnv=TEST_VAL}}}\" + \"hello\"}}}", nil, nil, nil))
}

func (s *TestFormatterSuite) TestInputWithPath() {
	assert.Equal(s.T(), "3", utils.Format("{{{FromInput=index}}}", nil, nil, builder.Object(map[string]builder.Interfacable{
		"index": builder.Number(3),
	})))
}

func (s *TestFormatterSuite) TestFormatter() {
	assert.Equal(s.T(), "", utils.Format("", nil, nil, nil))

	index := uint32(8)
	assert.Equal(s.T(), "TokenRef=my_token and TokenObjectRef=my_token Object=value kek {\"value\": \"value kek\"} Env=test 8 9 5", utils.Format("TokenRef={{{RefName=TokenRef}}} and TokenObjectRef={{{RefName=TokenObjectRef token}}} Object={{{value}}} {PL} Env={{{FromEnv=TEST_VAL}}} {INDEX} {HUMAN_INDEX} {{{FromInput=.}}}", builder.Object(map[string]builder.Interfacable{
		"value": builder.String("value kek"),
	}), &index, builder.Number(5)))
}

func (s *TestFormatterSuite) TestFile() {
	assert.Equal(s.T(), "hi test_content end", utils.Format("{{{FromExp='hi ' + '{{{FromFile=./test_file.log}}}' + ' end'}}}", nil, nil, nil))
	assert.Equal(s.T(), "hi 14 test end", utils.Format("{{{FromExp='hi ' + '{{{FromFile=./test_formatted_file.log}}}' + ' end'}}}", nil, nil, nil))
}

func (s *TestFormatterSuite) TestNewLineSeparator() {
	assert.Equal(s.T(), "\n", utils.Format("$__FLINE__$", nil, nil, nil))
}

func (s *TestFormatterSuite) TestExpr() {
	index := uint32(1)
	assert.Equal(s.T(), "8", utils.Format("{{{FromExp=fRes + 5 + fIndex}}}", builder.Number(2), &index, nil))
}

func (s *TestFormatterSuite) TestFromURL() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`14`))
	}))
	defer server.Close()

	assert.Equal(s.T(), "22", utils.Format(fmt.Sprintf("{{{FromExp=fRes + 5 + int('{{{FromURL=%s}}}')}}}", server.URL), builder.Number(3), nil, nil))
}

func (s *TestFormatterSuite) TearDownSuite() {
	_ = os.Unsetenv("TEST_VAL")
}
func (s *TestFormatterSuite) SetupSuite() {
	references.SetReference(map[string]*config.Reference{
		"TokenRef": {
			ModelField: &config.ModelField{
				ConnectorConfig: &config.ConnectorConfig{
					ResponseType: config.Json,
					StaticConfig: &config.StaticConnectorConfig{
						Value: builder.Object(map[string]builder.Interfacable{
							"token": builder.String("my_token"),
						}).ToJson(),
					},
				},
				Model: &config.Model{
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "token",
					},
				},
			},
		},
		"TokenObjectRef": {
			ModelField: &config.ModelField{
				ConnectorConfig: &config.ConnectorConfig{
					ResponseType: config.Json,
					StaticConfig: &config.StaticConnectorConfig{
						Value: builder.Object(map[string]builder.Interfacable{
							"token": builder.String("my_token"),
						}).ToJson(),
					},
				},
				Model: &config.Model{
					ObjectConfig: &config.ObjectConfig{
						Fields: map[string]*config.Field{
							"token": {
								BaseField: &config.BaseField{
									Type: config.String,
									Path: "token",
								},
							},
						},
					},
				},
			},
		},
	}, func(_ string, model *config.ModelField) (builder.Jsonable, error) {
		return parser.NewEngine(model.ConnectorConfig, logger.Null).Get(model.Model, nil, nil, nil)
	})
	os.Setenv("TEST_VAL", "test")
}

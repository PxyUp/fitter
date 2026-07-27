package parser_test

import (
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCondition(t *testing.T) {
	suite.Run(t, new(ConditionSuite))
}

type ConditionSuite struct {
	suite.Suite
	parser parser.Parser
}

func (s *ConditionSuite) SetupTest() {
	s.parser = parser.NewJson([]byte(`{
	"name": "fitter",
	"age": 27,
	"on_sale": false,
	"discount_pct": 15.5,
	"products": [
		{"title": "first", "price": 10, "featured": true},
		{"title": "second", "price": 0, "featured": true},
		{"title": "third", "price": 5, "featured": false}
	]
}`), logger.Null)
}

func (s *ConditionSuite) Test_BaseField_ConditionTrue_FieldKept() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"age": {
					BaseField: &config.BaseField{
						Type:      config.Int,
						Path:      "age",
						Condition: "fRes >= 18",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"age": 27}`, res.ToJson())
}

func (s *ConditionSuite) Test_BaseField_ConditionFalse_KeyOmitted() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"name": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "name",
					},
				},
				"discount": {
					BaseField: &config.BaseField{
						Type:      config.Float,
						Path:      "discount_pct",
						Condition: "fRes > 50",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"name": "fitter"}`, res.ToJson())
}

func (s *ConditionSuite) Test_BaseField_ConditionFalse_SkipsGenerated() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"generated": {
					BaseField: &config.BaseField{
						Type:      config.Bool,
						Path:      "on_sale",
						Condition: "fRes == true",
						Generated: &config.GeneratedFieldConfig{
							Static: &config.StaticGeneratedFieldConfig{
								Type:  config.String,
								Value: "should-not-appear",
							},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{}`, res.ToJson())
}

func (s *ConditionSuite) Test_BaseField_RootCondition_False_NullResult() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			Type:      config.String,
			Path:      "name",
			Condition: `fRes == "other"`,
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `null`, res.ToJson())
}

func (s *ConditionSuite) Test_BaseField_InvalidCondition_KeyOmitted() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"name": {
					BaseField: &config.BaseField{
						Type: config.String,
						Path: "name",
					},
				},
				"broken": {
					BaseField: &config.BaseField{
						Type:      config.Int,
						Path:      "age",
						Condition: "fRes >>> nonsense",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"name": "fitter"}`, res.ToJson())
}

func (s *ConditionSuite) Test_ArrayConfig_ItemCondition_FiltersItems() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "products",
			ItemCondition: "fRes.price > 0",
			ItemConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"title": {BaseField: &config.BaseField{Type: config.String, Path: "title"}},
					"price": {BaseField: &config.BaseField{Type: config.Float, Path: "price"}},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `[{"title": "first", "price": 10}, {"title": "third", "price": 5}]`, res.ToJson())
}

func (s *ConditionSuite) Test_ArrayConfig_ItemCondition_ByIndex() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "products",
			ItemCondition: "fIndex != 1",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{Type: config.String, Path: "title"},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `["first", "third"]`, res.ToJson())
}

func (s *ConditionSuite) Test_ArrayConfig_ItemCondition_ScalarItems() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "products.#.price",
			ItemCondition: "fRes > 0",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{Type: config.Float},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `[10, 5]`, res.ToJson())
}

func (s *ConditionSuite) Test_ItemCondition_FSrc_UnextractedSourceField() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "products",
			ItemCondition: "fSrc.featured == true && fRes.price > 0",
			ItemConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"title": {BaseField: &config.BaseField{Type: config.String, Path: "title"}},
					"price": {BaseField: &config.BaseField{Type: config.Float, Path: "price"}},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `[{"title": "first", "price": 10}]`, res.ToJson())
}

func (s *ConditionSuite) Test_BaseField_FSrc_SiblingAccess() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"name": {
					BaseField: &config.BaseField{
						Type:      config.String,
						Path:      "name",
						Condition: "fSrc.age > 18",
					},
				},
				"discount": {
					BaseField: &config.BaseField{
						Type:      config.Float,
						Path:      "discount_pct",
						Condition: "fSrc.on_sale == true",
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"name": "fitter"}`, res.ToJson())
}

func (s *ConditionSuite) Test_HTML_ItemCondition_FSrc() {
	htmlParser := parser.NewHTML([]byte(`<html><body>
		<ul>
			<li class="price">10</li>
			<li class="price">0</li>
			<li class="price">5</li>
		</ul>
	</body></html>`), logger.Null)

	res, err := htmlParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "ul li.price",
			ItemCondition: "fSrc > 0",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{Type: config.String},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `["10", "5"]`, res.ToJson())
}

func (s *ConditionSuite) Test_ArrayConfig_Condition_False_KeyOmitted() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"name": {
					BaseField: &config.BaseField{Type: config.String, Path: "name"},
				},
				"products": {
					ArrayConfig: &config.ArrayConfig{
						RootPath:  "products",
						Condition: "fRes.on_sale == true",
						ItemConfig: &config.ObjectConfig{
							Field: &config.BaseField{Type: config.String, Path: "title"},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"name": "fitter"}`, res.ToJson())
}

func (s *ConditionSuite) Test_ObjectConfig_Condition_False_KeyOmitted() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"name": {
					BaseField: &config.BaseField{Type: config.String, Path: "name"},
				},
				"sale_info": {
					ObjectConfig: &config.ObjectConfig{
						Condition: "fRes.on_sale == true",
						Fields: map[string]*config.Field{
							"discount": {BaseField: &config.BaseField{Type: config.Float, Path: "discount_pct"}},
						},
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"name": "fitter"}`, res.ToJson())
}

func (s *ConditionSuite) Test_ObjectConfig_RootCondition_False_NullResult() {
	res, err := s.parser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Condition: "fRes.age > 100",
			Fields: map[string]*config.Field{
				"name": {BaseField: &config.BaseField{Type: config.String, Path: "name"}},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `null`, res.ToJson())
}

func (s *ConditionSuite) Test_FirstOf_ConditionFalse_FallsThrough() {
	res, err := s.parser.Parse(&config.Model{
		BaseField: &config.BaseField{
			FirstOf: []*config.BaseField{
				{
					Type:      config.String,
					Path:      "name",
					Condition: "fRes == \"other\"",
				},
				{
					Type: config.Int,
					Path: "age",
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `27`, res.ToJson())
}

func (s *ConditionSuite) Test_StaticArray_ConditionFalse_KeepsNullSlot() {
	res, err := s.parser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			StaticConfig: &config.StaticArrayConfig{
				Items: map[uint32]*config.Field{
					0: {BaseField: &config.BaseField{Type: config.String, Path: "name"}},
					1: {BaseField: &config.BaseField{
						Type:      config.Int,
						Path:      "age",
						Condition: "fRes > 100",
					}},
					2: {BaseField: &config.BaseField{Type: config.Int, Path: "age"}},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `["fitter", null, 27]`, res.ToJson())
}

func (s *ConditionSuite) Test_HTML_ItemCondition_FiltersByText() {
	htmlParser := parser.NewHTML([]byte(`<html><body>
		<ul>
			<li class="price">10</li>
			<li class="price">0</li>
			<li class="price">5</li>
		</ul>
	</body></html>`), logger.Null)

	res, err := htmlParser.Parse(&config.Model{
		ArrayConfig: &config.ArrayConfig{
			RootPath:      "ul li.price",
			ItemCondition: "fRes > 0",
			ItemConfig: &config.ObjectConfig{
				Field: &config.BaseField{Type: config.Int},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `[10, 5]`, res.ToJson())
}

func (s *ConditionSuite) Test_HTML_BaseField_ConditionOnText() {
	htmlParser := parser.NewHTML([]byte(`<html><body>
		<div class="title">hello</div>
		<div class="status">inactive</div>
	</body></html>`), logger.Null)

	res, err := htmlParser.Parse(&config.Model{
		ObjectConfig: &config.ObjectConfig{
			Fields: map[string]*config.Field{
				"title": {
					BaseField: &config.BaseField{Type: config.String, Path: "div.title"},
				},
				"status": {
					BaseField: &config.BaseField{
						Type:      config.String,
						Path:      "div.status",
						Condition: `fRes == "active"`,
					},
				},
			},
		},
	}, nil)
	assert.NoError(s.T(), err)
	assert.JSONEq(s.T(), `{"title": "hello"}`, res.ToJson())
}

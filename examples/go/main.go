package main

import (
	"fmt"
	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"log"
	"net/http"
)

func main() {
	res, err := lib.Parse(&config.Item{
		ConnectorConfig: &config.ConnectorConfig{
			ResponseType: config.Json,
			Url:          "https://random-data-api.com/api/appliance/random_appliance",
			ServerConfig: &config.ServerConnectorConfig{
				Method: http.MethodGet,
			},
		},
		Model: &config.Model{
			ObjectConfig: &config.ObjectConfig{
				Fields: map[string]*config.Field{
					"my_id": {
						BaseField: &config.BaseField{
							Type: config.Int,
							Path: "id",
						},
					},
					"generated_id": {
						BaseField: &config.BaseField{
							Generated: &config.GeneratedFieldConfig{
								UUID: &config.UUIDGeneratedFieldConfig{},
							},
						},
					},
					"generated_array": {
						ArrayConfig: &config.ArrayConfig{
							RootPath: "@this|@keys",
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
	}, nil, nil, nil, logger.Null)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.ToJson())
}

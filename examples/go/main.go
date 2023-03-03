package main

import (
	"fmt"
	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/config"
	"log"
	"net/http"
)

func main() {
	res, err := lib.Parse(&config.Item{
		ConnectorConfig: &config.ConnectorConfig{
			ConnectorType: config.Server,
			ResponseType:  config.Json,
			ServerConfig: &config.ServerConnectorConfig{
				Method: http.MethodGet,
				Url:    "https://random-data-api.com/api/appliance/random_appliance",
			},
		},
		Model: &config.Model{
			Type: config.ObjectModel,
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
	}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.ToJson())
}

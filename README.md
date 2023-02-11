# Fitter (development)

Fitter - new way for collect information from the API's/Websites

# Way to collect information

1. **Server** - parsing response from some API's or http request
2. **Browser** - emulate real browser using cypress and get DOM information

# Format which can be parsed

1. **JSON** - parsing JSON to get specific information
2. **XML** - parsing xml tree to get specific information
3. **HTML** - parsing dom tree to get specific information
4. **XPath** - parsing dom tree to get specific information but by xpath

# Environment variables
1. **FITTER_HTTP_WORKER** - int[1000] - default concurrent HTTP workers

# How to run locally

### Arguments
1. **--path** - string[config.yaml] - path for the configuration of the Fitter

```bash
go run cmd/fitter/main.go --path=./examples/config_api.json
go run cmd/fitter/main.go --path=./examples/config_web.json
```

# How to use Fitter_CLI

### Arguments
1. **--path** - string[config.yaml] - path for the configuration of the Fitter_CLI
2. **--copy** - bool[false] - copy information into clipboard
3. **--pretty** - bool[true] - make readable result(also affect on copy)

```bash
go run cmd/cli/main.go --path=./examples/cli/config_cli.json --copy=true
```

```json
{
  "connector_config": {
    "response_type": "json",
    "connector_type": "server",
    "server_config": {
      "method": "GET",
      "url": "https://hacker-news.firebaseio.com/v0/beststories.json?print=pretty&limitToFirst=10&orderBy=%22$key%22"
    }
  },
  "model": {
    "type": "object",
    "object_config": {
      "fields": {
        "response_id": {
          "base_field": {
            "generated": {
              "uuid": {}
            }
          }
        },
        "sky_news": {
          "base_field": {
            "generated": {
              "model": {
                "type": "array",
                "model": {
                  "type": "array",
                  "array_config": {
                    "root_path": "//div[@class='fc-item__standfirst']",
                    "item_config": {
                      "field": {
                        "type": "string",
                        "path": "text()"
                      }
                    }
                  }
                },
                "connector_config": {
                  "response_type": "xpath",
                  "connector_type": "server",
                  "server_config": {
                    "method": "GET",
                    "url": "https://www.theguardian.com/world"
                  }
                }
              }
            }
          }
        },
        "quotes": {
          "base_field": {
            "generated": {
              "model": {
                "type": "array",
                "model": {
                  "type": "array",
                  "array_config": {
                    "root_path": "#content dt.quote > a",
                    "item_config": {
                      "field": {
                        "type": "string"
                      }
                    }
                  }
                },
                "connector_config": {
                  "response_type": "HTML",
                  "connector_type": "server",
                  "server_config": {
                    "method": "GET",
                    "url": "http://www.quotationspage.com/random.php"
                  }
                }
              }
            }
          }
        },
        "hackernews": {
          "array_config": {
            "item_config": {
              "fields": {
                "id": {
                  "base_field": {
                    "type": "int"
                  }
                },
                "internal_url": {
                  "base_field": {
                    "type": "int",
                    "generated": {
                      "formatted": {
                        "template": "https://news.ycombinator.com/item?id=%s"
                      }
                    }
                  }
                },
                "content": {
                  "base_field": {
                    "type": "int",
                    "generated": {
                      "model": {
                        "type": "object",
                        "model": {
                          "type": "object",
                          "object_config": {
                            "fields": {
                              "title": {
                                "base_field": {
                                  "type": "string",
                                  "path": "title"
                                }
                              },
                              "score": {
                                "base_field": {
                                  "type": "int",
                                  "path": "score"
                                }
                              }
                            }
                          }
                        },
                        "connector_config": {
                          "response_type": "json",
                          "connector_type": "server",
                          "server_config": {
                            "method": "GET",
                            "url": "https://hacker-news.firebaseio.com/v0/item/%s.json?print=pretty"
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

# Roadmap

1. XML/HTML parsers - 1.0
2. Browser - emulation - 1.0
3. Notification: Webhook - 1.0
4. Trigger: Webhook/Queue - 1.0

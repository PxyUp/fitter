# Fitter (development)

Fitter - new way for collect information from the API's/Websites

# Way to collect information

1. **Server** - parsing response from some API's or http request
2. **Browser** - emulate real browser using cypress and get DOM information

# Format which can be parsed

1. **JSON** - parsing JSON to get specific information
2. **XML** - parsing xml tree to get specific information
3. **HTML** - parsing dom tree to get specific information

# How to run locally

```bash
go run cmd/fitter/main.go --path=./examples/config_api.json
go run cmd/fitter/main.go --path=./examples/config_web.json
```

# Roadmap

1. XML/HTML parsers - 1.0
2. Browser - emulation - 1.0
3. Notification: Webhook - 1.0
4. Trigger: Webhook/Queue - 1.0

# Plugins

Fitter and Fitter CLI can have external plugins which should implement interface

# Usage

--plugins - looking for all files with ".so" extension in provided folder(subdirs excluded)


```bash
./fitter_cli_${VERSION} --path=./examples/plugin/plugin_cli.json --plugins=./examples/plugin --copy=true
```

```bash
./fitter_${VERSION} --path=./examples/plugin/plugin.json --plugins=./examples/plugin
```

# Example of plugin


Build plugin
```bash
go build -buildmode=plugin -o examples/plugin/hardcoder.so examples/plugin/hardcoder/hardcoder.go
```

Make sure you export **Plugin** variable which implements **pl.Plugin** interface

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	pl "github.com/PxyUp/fitter/pkg/plugins/plugin"
)

var (
	_ pl.Plugin = &plugin{}

	Plugin plugin
)

type plugin struct {
	Name string `json:"name" yaml:"name"`
}

func (pl *plugin) Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	if field.Config != nil {
		err := json.Unmarshal(field.Config, pl)
		if err != nil {
			logger.Errorw("cant unmarshal plugin configuration", "error", err.Error())
			return builder.Null()
		}
		return builder.String(fmt.Sprintf("Hello %s", pl.Name))
	}

	return builder.String(fmt.Sprintf("Hello %s", parsedValue.ToJson()))
}

```
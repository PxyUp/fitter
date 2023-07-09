package store

import (
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/parser/builder"
	"github.com/PxyUp/fitter/pkg/plugins/plugin"
	"os"
	"path"
	pl "plugin"
	"strings"
	"sync"
)

type nullPlugin struct {
}

func (n *nullPlugin) Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	return builder.Null()
}

var (
	null = &nullPlugin{}

	Store = &store{
		kv: make(map[string]plugin.Plugin),
	}
)

type store struct {
	kv map[string]plugin.Plugin

	mutex sync.Mutex
}

func (s *store) Add(name string, plugin plugin.Plugin) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.kv[name] = plugin
}

func (s *store) Get(name string, log logger.Logger) plugin.Plugin {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if pl, exists := s.kv[name]; exists {
		return pl
	}

	log.Infof("Cant find plugin with name: %s", name)

	return null
}

func PluginInitialize(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".so") {

			errPlugin := processPlugin(path.Join(dirPath, e.Name()))
			if errPlugin != nil {
				return errPlugin
			}
		}
	}

	return nil
}

func processPlugin(fileName string) error {
	plug, err := pl.Open(fileName)
	if err != nil {
		return err
	}

	symPlugin, err := plug.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("%s not implement plugin interface", fileName)
	}

	localPlugin, ok := symPlugin.(plugin.Plugin)
	if !ok {
		return fmt.Errorf("%s not implement plugin interface plugin.Plugin", fileName)
	}

	Store.Add(strings.TrimSuffix(path.Base(fileName), path.Ext(fileName)), localPlugin)

	return nil
}

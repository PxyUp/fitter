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

type nullFieldPlugin struct {
}

func (n *nullFieldPlugin) Format(parsedValue builder.Jsonable, field *config.PluginFieldConfig, logger logger.Logger, index *uint32) builder.Jsonable {
	return builder.Null()
}

type nullConnectorPlugin struct {
}

func (n *nullConnectorPlugin) SetConfig(cfg *config.PluginConnectorConfig, _ logger.Logger) {
	return
}

func (n *nullConnectorPlugin) Get(parsedValue builder.Jsonable, index *uint32) ([]byte, error) {
	return []byte{}, nil
}

var (
	nullField     plugin.FieldPlugin     = &nullFieldPlugin{}
	nullConnector plugin.ConnectorPlugin = &nullConnectorPlugin{}

	Store = &store{
		fieldPlugins:     make(map[string]plugin.FieldPlugin),
		connectorPlugins: make(map[string]plugin.ConnectorPlugin),
	}
)

type store struct {
	fieldPlugins map[string]plugin.FieldPlugin

	connectorPlugins map[string]plugin.ConnectorPlugin

	mField     sync.Mutex
	mConnector sync.Mutex
}

func (s *store) AddFieldPlugin(name string, plugin plugin.FieldPlugin) {
	s.mField.Lock()
	defer s.mField.Unlock()

	s.fieldPlugins[name] = plugin
}

func (s *store) GetFieldPlugin(name string, log logger.Logger) plugin.FieldPlugin {
	s.mField.Lock()
	defer s.mField.Unlock()

	if pl, exists := s.fieldPlugins[name]; exists {
		return pl
	}

	log.Infof("Cant find plugin with name: %s", name)

	return nullField
}

func (s *store) AddConnectorPlugin(name string, plugin plugin.ConnectorPlugin) {
	s.mConnector.Lock()
	defer s.mConnector.Unlock()

	s.connectorPlugins[name] = plugin
}

func (s *store) GetConnectorPlugin(name string, cfg *config.PluginConnectorConfig, log logger.Logger) plugin.ConnectorPlugin {
	s.mConnector.Lock()
	defer s.mConnector.Unlock()

	if pl, exists := s.connectorPlugins[name]; exists {
		pl.SetConfig(cfg, log)
		return pl
	}

	log.Infof("Cant find connector plugin with name: %s", name)

	return nullConnector
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

	localFieldPlugin, okField := symPlugin.(plugin.FieldPlugin)
	if okField {
		Store.AddFieldPlugin(strings.TrimSuffix(path.Base(fileName), path.Ext(fileName)), localFieldPlugin)
		return nil
	}

	localConnectorPlugin, okConnector := symPlugin.(plugin.ConnectorPlugin)
	if okConnector {
		Store.AddConnectorPlugin(strings.TrimSuffix(path.Base(fileName), path.Ext(fileName)), localConnectorPlugin)
		return nil
	}

	return fmt.Errorf("%s not implement plugin interface plugin.ConnectorPlugin or plugin.FieldPlugin", fileName)
}

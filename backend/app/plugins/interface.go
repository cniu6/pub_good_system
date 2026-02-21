package plugins

import (
	"github.com/gin-gonic/gin"
)

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string
	
	// Version returns the version of the plugin
	Version() string
	
	// Init initializes the plugin (e.g., database migrations)
	Init() error
	
	// RegisterRoutes allows the plugin to register its own routes
	RegisterRoutes(router *gin.RouterGroup)
}

// PluginManager manages all registered plugins
type PluginManager struct {
	plugins map[string]Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) Register(p Plugin) {
	pm.plugins[p.Name()] = p
}

func (pm *PluginManager) GetPlugins() map[string]Plugin {
	return pm.plugins
}

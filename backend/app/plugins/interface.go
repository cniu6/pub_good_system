package plugins

import (
	"github.com/gin-gonic/gin"
)

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// ========================================
	// 基础信息
	// ========================================

	// Name returns the unique name of the plugin
	Name() string

	// Version returns the version of the plugin
	Version() string

	// Description returns the description of the plugin
	Description() string

	// ========================================
	// 生命周期管理
	// ========================================

	// Priority returns the priority of the plugin (lower = loaded first)
	// Default: 100
	Priority() int

	// Dependencies returns the list of plugin names this plugin depends on
	// These plugins will be loaded before this plugin
	Dependencies() []string

	// Configure receives configuration from the system
	// Called before Init()
	Configure(config map[string]interface{}) error

	// Migrate performs database migrations
	// Called after Init()
	Migrate() error

	// Init initializes the plugin (e.g., database connections, caches)
	Init() error

	// RegisterRoutes allows the plugin to register its own routes
	RegisterRoutes(router *gin.RouterGroup)

	// Shutdown performs cleanup when the system shuts down
	Shutdown() error
}

// BasePlugin 插件基类，提供默认实现
// 可以嵌入到插件结构体中，避免实现所有方法
type BasePlugin struct {
	name         string
	version      string
	description  string
	priority     int
	dependencies []string
}

// NewBasePlugin 创建插件基类
func NewBasePlugin(name, version, description string) BasePlugin {
	return BasePlugin{
		name:        name,
		version:     version,
		description: description,
		priority:    100, // 默认优先级
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Version() string {
	return p.version
}

func (p *BasePlugin) Description() string {
	return p.description
}

func (p *BasePlugin) Priority() int {
	return p.priority
}

func (p *BasePlugin) SetPriority(priority int) {
	p.priority = priority
}

func (p *BasePlugin) Dependencies() []string {
	return p.dependencies
}

func (p *BasePlugin) SetDependencies(deps []string) {
	p.dependencies = deps
}

func (p *BasePlugin) Configure(config map[string]interface{}) error {
	return nil // 默认不做任何事
}

func (p *BasePlugin) Migrate() error {
	return nil // 默认不做任何事
}

func (p *BasePlugin) Init() error {
	return nil // 默认不做任何事
}

func (p *BasePlugin) RegisterRoutes(router *gin.RouterGroup) {
	// 默认不注册任何路由
}

func (p *BasePlugin) Shutdown() error {
	return nil // 默认不做任何事
}

// PluginInfo 插件信息（用于展示）
type PluginInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Priority     int      `json:"priority"`
	Dependencies []string `json:"dependencies"`
	Status       string   `json:"status"` // "active", "inactive", "error"
}

// PluginConfig 插件配置
type PluginConfig map[string]interface{}

// PluginManager manages all registered plugins
type PluginManager struct {
	plugins    map[string]Plugin
	configs    map[string]PluginConfig
	init_order []string // 初始化顺序
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]PluginConfig),
	}
}

// Register registers a plugin
func (pm *PluginManager) Register(p Plugin) {
	pm.plugins[p.Name()] = p
}

// RegisterWithConfig registers a plugin with configuration
func (pm *PluginManager) RegisterWithConfig(p Plugin, config PluginConfig) {
	pm.plugins[p.Name()] = p
	pm.configs[p.Name()] = config
}

// GetPlugin gets a plugin by name
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	p, ok := pm.plugins[name]
	return p, ok
}

// GetPlugins returns all registered plugins
func (pm *PluginManager) GetPlugins() map[string]Plugin {
	return pm.plugins
}

// GetPluginInfos returns plugin information list
func (pm *PluginManager) GetPluginInfos() []PluginInfo {
	var infos []PluginInfo
	for _, p := range pm.plugins {
		info := PluginInfo{
			Name:         p.Name(),
			Version:      p.Version(),
			Description:  p.Description(),
			Priority:     p.Priority(),
			Dependencies: p.Dependencies(),
			Status:       "active",
		}
		infos = append(infos, info)
	}
	return infos
}

// SetConfig sets configuration for a plugin
func (pm *PluginManager) SetConfig(name string, config PluginConfig) {
	pm.configs[name] = config
}

// GetConfig gets configuration for a plugin
func (pm *PluginManager) GetConfig(name string) PluginConfig {
	if config, ok := pm.configs[name]; ok {
		return config
	}
	return make(PluginConfig)
}

// HasPlugin checks if a plugin exists
func (pm *PluginManager) HasPlugin(name string) bool {
	_, ok := pm.plugins[name]
	return ok
}

// Count returns the number of registered plugins
func (pm *PluginManager) Count() int {
	return len(pm.plugins)
}

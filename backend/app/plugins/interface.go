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

// PluginConfig 插件配置
type PluginConfig map[string]interface{}


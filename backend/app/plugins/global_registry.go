package plugins

import (
	"fst/backend/pkg/pluginregistry"
	"sync"
)

// globalPluginRegistry 本地插件注册表
var globalPluginRegistry = &pluginRegistry{
	plugins: make(map[string]Plugin),
}

// pluginRegistry 插件注册表
type pluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// RegisterPlugin 注册插件到全局注册表
// 在插件的 init() 函数中调用此函数
func RegisterPlugin(p Plugin) {
	globalPluginRegistry.mu.Lock()
	defer globalPluginRegistry.mu.Unlock()
	globalPluginRegistry.plugins[p.Name()] = p
}

// AutoRegisterAll 将所有已注册的插件添加到插件管理器
// 在 main.go 中调用此函数
func AutoRegisterAll(mgr *Manager) {
	// 从 pluginregistry 获取所有通过 init() 注册的插件
	for _, p := range pluginregistry.GetAll() {
		// 类型断言获取完整 Plugin 接口
		if fullPlugin, ok := p.(Plugin); ok {
			mgr.Register(fullPlugin)
		}
	}

	// 也从本地注册表获取（兼容旧代码）
	globalPluginRegistry.mu.RLock()
	for _, p := range globalPluginRegistry.plugins {
		mgr.Register(p)
	}
	globalPluginRegistry.mu.RUnlock()
}

// GetRegisteredPlugins 获取已注册的插件列表
func GetRegisteredPlugins() []Plugin {
	// 从本地注册表获取
	globalPluginRegistry.mu.RLock()
	defer globalPluginRegistry.mu.RUnlock()

	plugins := make([]Plugin, 0, len(globalPluginRegistry.plugins))
	for _, p := range globalPluginRegistry.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// GetPluginByName 从全局注册表获取插件
func GetPluginByName(name string) Plugin {
	globalPluginRegistry.mu.RLock()
	defer globalPluginRegistry.mu.RUnlock()

	if p, ok := globalPluginRegistry.plugins[name]; ok {
		return p
	}
	return nil
}

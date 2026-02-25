// Package pluginregistry 插件注册表
// 独立包，避免循环导入问题
package pluginregistry

import (
	"sync"
)

// Plugin 插件接口 - 最小化定义
// 插件需要实现这个接口才能被注册
type Plugin interface {
	Name() string
	Version() string
}

// 全局注册表
var (
	mu      sync.RWMutex
	plugins = make(map[string]Plugin)
)

// Register 注册插件
// 在插件的 init() 函数中调用
func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins[p.Name()] = p
}

// GetAll 获取所有已注册的插件
func GetAll() []Plugin {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Plugin, 0, len(plugins))
	for _, p := range plugins {
		result = append(result, p)
	}
	return result
}

// Count 获取已注册插件数量
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(plugins)
}

// Clear 清空注册表（用于测试）
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	plugins = make(map[string]Plugin)
}

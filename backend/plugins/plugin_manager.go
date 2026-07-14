package plugins

import (
	"errors"
	"log"
	"sync"
)

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]Plugin // 插件实例映射
	mu      sync.RWMutex      // 读写锁
}

// manager 全局插件管理器实例
var manager *PluginManager
var once sync.Once

// GetPluginManager 获取插件管理器单例
func GetPluginManager() *PluginManager {
	once.Do(func() {
		manager = &PluginManager{
			plugins: make(map[string]Plugin),
		}
	})
	return manager
}

// Register 注册插件
// 参数: plugin-插件实例
// 返回: 错误信息
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pluginID := plugin.GetID()
	if pluginID == "" {
		return errors.New("插件ID不能为空")
	}

	// 检查插件是否已注册
	if _, exists := pm.plugins[pluginID]; exists {
		return errors.New("插件已存在: " + pluginID)
	}

	// 初始化插件
	if err := plugin.Init(plugin.GetConfig()); err != nil {
		return errors.New("插件初始化失败: " + err.Error())
	}

	// 注册插件
	pm.plugins[pluginID] = plugin
	log.Printf("插件注册成功: %s(%s)", plugin.GetName(), pluginID)

	return nil
}

// Unregister 注销插件
// 参数: pluginID-插件ID
// 返回: 错误信息
func (pm *PluginManager) Unregister(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return errors.New("插件不存在: " + pluginID)
	}

	// 卸载插件
	if err := plugin.Uninstall(); err != nil {
		log.Printf("插件卸载失败: %s, 错误: %v", pluginID, err)
	}

	delete(pm.plugins, pluginID)
	log.Printf("插件注销成功: %s", pluginID)

	return nil
}

// Get 获取插件
// 参数: pluginID-插件ID
// 返回: 插件实例和是否找到
func (pm *PluginManager) Get(pluginID string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	return plugin, exists
}

// GetAll 获取所有插件
// 返回: 插件列表
func (pm *PluginManager) GetAll() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetByType 根据类型获取插件
// 参数: pluginType-插件类型
// 返回: 插件列表
func (pm *PluginManager) GetByType(pluginType PluginType) []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, plugin := range pm.plugins {
		if plugin.GetType() == pluginType {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

// Enable 启用插件
// 参数: pluginID-插件ID
// 返回: 错误信息
func (pm *PluginManager) Enable(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return errors.New("插件不存在: " + pluginID)
	}

	return plugin.Enable()
}

// Disable 禁用插件
// 参数: pluginID-插件ID
// 返回: 错误信息
func (pm *PluginManager) Disable(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return errors.New("插件不存在: " + pluginID)
	}

	return plugin.Disable()
}

// GetPaymentPlugin 获取支付插件
// 参数: pluginID-插件ID
// 返回: 支付插件和是否找到
func (pm *PluginManager) GetPaymentPlugin(pluginID string) (PaymentPlugin, bool) {
	plugin, exists := pm.Get(pluginID)
	if !exists {
		return nil, false
	}
	if paymentPlugin, ok := plugin.(PaymentPlugin); ok {
		return paymentPlugin, true
	}
	return nil, false
}

// GetShippingPlugin 获取物流插件
// 参数: pluginID-插件ID
// 返回: 物流插件和是否找到
func (pm *PluginManager) GetShippingPlugin(pluginID string) (ShippingPlugin, bool) {
	plugin, exists := pm.Get(pluginID)
	if !exists {
		return nil, false
	}
	if shippingPlugin, ok := plugin.(ShippingPlugin); ok {
		return shippingPlugin, true
	}
	return nil, false
}

// GetNotificationPlugin 获取通知插件
// 参数: pluginID-插件ID
// 返回: 通知插件和是否找到
func (pm *PluginManager) GetNotificationPlugin(pluginID string) (NotificationPlugin, bool) {
	plugin, exists := pm.Get(pluginID)
	if !exists {
		return nil, false
	}
	if notificationPlugin, ok := plugin.(NotificationPlugin); ok {
		return notificationPlugin, true
	}
	return nil, false
}

// GetProductExtension 获取产品扩展插件
// 参数: pluginID-插件ID
// 返回: 产品扩展插件和是否找到
func (pm *PluginManager) GetProductExtension(pluginID string) (ProductExtension, bool) {
	plugin, exists := pm.Get(pluginID)
	if !exists {
		return nil, false
	}
	if productExt, ok := plugin.(ProductExtension); ok {
		return productExt, true
	}
	return nil, false
}

package plugins

import (
	"fst/backend/pkg/pluginregistry"
)

// AutoRegisterAll 将 pluginregistry 中已注册的插件装载到管理器。
// 各插件在 init() 里通过 pluginregistry.Register 注册，此处统一装载。
func AutoRegisterAll(mgr *Manager) {
	for _, p := range pluginregistry.GetAll() {
		if fullPlugin, ok := p.(Plugin); ok {
			mgr.Register(fullPlugin)
		}
	}
}

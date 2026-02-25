package plugins

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

// Manager 插件管理器（增强版）
type Manager struct {
	pm          *PluginManager
	mu          sync.RWMutex
	initialized bool
	shutdown    bool
	errors      map[string]error // 插件初始化错误记录
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		pm:     NewPluginManager(),
		errors: make(map[string]error),
	}
}

// Register 注册插件
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pm.Register(p)
}

// RegisterWithConfig 注册带配置的插件
func (m *Manager) RegisterWithConfig(p Plugin, config PluginConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pm.RegisterWithConfig(p, config)
}

// LoadAll 加载所有插件
// 按优先级和依赖关系排序后依次初始化
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("插件已经初始化")
	}

	// 解析依赖
	if err := m.resolveDependencies(); err != nil {
		return err
	}

	// 按优先级排序
	sorted_plugins := m.sortByPriority()

	// 依次初始化
	for _, name := range sorted_plugins {
		p := m.pm.plugins[name]

		// 1. 配置
		config := m.pm.GetConfig(name)
		if err := p.Configure(config); err != nil {
			m.errors[name] = fmt.Errorf("配置失败: %v", err)
			log.Printf("[Plugin] %s 配置失败: %v", name, err)
			continue
		}

		// 2. 初始化
		if err := p.Init(); err != nil {
			m.errors[name] = fmt.Errorf("初始化失败: %v", err)
			log.Printf("[Plugin] %s 初始化失败: %v", name, err)
			continue
		}

		// 3. 数据库迁移
		if err := p.Migrate(); err != nil {
			m.errors[name] = fmt.Errorf("迁移失败: %v", err)
			log.Printf("[Plugin] %s 迁移失败: %v", name, err)
			continue
		}

		log.Printf("[Plugin] %s v%s 加载成功", p.Name(), p.Version())
	}

	m.initialized = true
	return nil
}

// RegisterAllRoutes 注册所有插件的路由
func (m *Manager) RegisterAllRoutes(router *gin.RouterGroup) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sorted_plugins := m.sortByPriority()

	for _, name := range sorted_plugins {
		p := m.pm.plugins[name]

		// 跳过初始化失败的插件
		if _, has_error := m.errors[name]; has_error {
			continue
		}

		p.RegisterRoutes(router)
		log.Printf("[Plugin] %s 路由注册完成", name)
	}
}

// ShutdownAll 关闭所有插件
func (m *Manager) ShutdownAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shutdown {
		return nil
	}

	// 按优先级逆序关闭
	sorted_plugins := m.sortByPriority()
	for i := len(sorted_plugins) - 1; i >= 0; i-- {
		name := sorted_plugins[i]
		p := m.pm.plugins[name]

		if err := p.Shutdown(); err != nil {
			log.Printf("[Plugin] %s 关闭失败: %v", name, err)
		} else {
			log.Printf("[Plugin] %s 已关闭", name)
		}
	}

	m.shutdown = true
	return nil
}

// resolveDependencies 解析依赖关系
func (m *Manager) resolveDependencies() error {
	// 检查依赖是否存在
	for name, p := range m.pm.plugins {
		for _, dep := range p.Dependencies() {
			if !m.pm.HasPlugin(dep) {
				return fmt.Errorf("插件 %s 依赖的 %s 不存在", name, dep)
			}
		}
	}

	// 检查循环依赖
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	for name := range m.pm.plugins {
		if err := m.checkCycle(name, visited, visiting); err != nil {
			return err
		}
	}

	return nil
}

// checkCycle 检查循环依赖（DFS）
func (m *Manager) checkCycle(name string, visited, visiting map[string]bool) error {
	if visited[name] {
		return nil
	}

	if visiting[name] {
		return fmt.Errorf("检测到循环依赖: %s", name)
	}

	visiting[name] = true

	p, ok := m.pm.plugins[name]
	if !ok {
		return nil
	}

	for _, dep := range p.Dependencies() {
		if err := m.checkCycle(dep, visited, visiting); err != nil {
			return err
		}
	}

	visiting[name] = false
	visited[name] = true
	return nil
}

// sortByPriority 按优先级和依赖关系排序
func (m *Manager) sortByPriority() []string {
	plugins := make([]Plugin, 0, len(m.pm.plugins))
	for _, p := range m.pm.plugins {
		plugins = append(plugins, p)
	}

	// 拓扑排序 + 优先级排序
	result := m.topologicalSort(plugins)

	return result
}

// topologicalSort 拓扑排序
func (m *Manager) topologicalSort(plugins []Plugin) []string {
	// 构建入度表
	in_degree := make(map[string]int)
	adj := make(map[string][]string)

	for _, p := range plugins {
		name := p.Name()
		in_degree[name] = 0
		adj[name] = []string{}
	}

	// 构建邻接表
	for _, p := range plugins {
		for _, dep := range p.Dependencies() {
			adj[dep] = append(adj[dep], p.Name())
			in_degree[p.Name()]++
		}
	}

	// 按优先级排序的同级节点
	var queue []string
	for name, degree := range in_degree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// 按优先级排序入度为0的节点
	sort.Slice(queue, func(i, j int) bool {
		return m.pm.plugins[queue[i]].Priority() < m.pm.plugins[queue[j]].Priority()
	})

	var result []string
	for len(queue) > 0 {
		// 取出第一个
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 更新邻接节点的入度
		var next_zero []string
		for _, neighbor := range adj[current] {
			in_degree[neighbor]--
			if in_degree[neighbor] == 0 {
				next_zero = append(next_zero, neighbor)
			}
		}

		// 按优先级排序新增的入度为0的节点
		sort.Slice(next_zero, func(i, j int) bool {
			return m.pm.plugins[next_zero[i]].Priority() < m.pm.plugins[next_zero[j]].Priority()
		})

		queue = append(queue, next_zero...)
	}

	return result
}

// GetPlugin 获取插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	return m.pm.GetPlugin(name)
}

// GetPlugins 获取所有插件
func (m *Manager) GetPlugins() map[string]Plugin {
	return m.pm.GetPlugins()
}

// GetPluginInfos 获取插件信息列表
func (m *Manager) GetPluginInfos() []PluginInfo {
	infos := m.pm.GetPluginInfos()

	// 更新状态
	for i := range infos {
		if err, ok := m.errors[infos[i].Name]; ok {
			infos[i].Status = "error: " + err.Error()
		} else if m.initialized {
			infos[i].Status = "active"
		} else {
			infos[i].Status = "inactive"
		}
	}

	return infos
}

// GetErrors 获取插件错误
func (m *Manager) GetErrors() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]error)
	for k, v := range m.errors {
		result[k] = v
	}
	return result
}

// IsInitialized 检查是否已初始化
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// Count 获取插件数量
func (m *Manager) Count() int {
	return m.pm.Count()
}

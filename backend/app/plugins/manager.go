package plugins

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

// Manager 插件管理器：注册、依赖解析、加载、路由与关闭
type Manager struct {
	plugins     map[string]Plugin
	configs     map[string]PluginConfig
	mu          sync.RWMutex
	initialized bool
	shutdown    bool
	errors      map[string]error // 插件初始化错误记录
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]PluginConfig),
		errors:  make(map[string]error),
	}
}

// Register 注册插件
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

// pluginConfig 取插件配置，无则返回空 map
func (m *Manager) pluginConfig(name string) PluginConfig {
	if config, ok := m.configs[name]; ok {
		return config
	}
	return make(PluginConfig)
}

// LoadAll 加载所有插件（按优先级与依赖排序后依次初始化）
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("插件已经初始化")
	}

	if err := m.resolveDependencies(); err != nil {
		return err
	}

	sorted_plugins := m.sortByPriority()

	for _, name := range sorted_plugins {
		p := m.plugins[name]

		config := m.pluginConfig(name)
		if err := p.Configure(config); err != nil {
			m.errors[name] = fmt.Errorf("配置失败: %v", err)
			log.Printf("[Plugin] %s 配置失败: %v", name, err)
			continue
		}

		if err := p.Init(); err != nil {
			m.errors[name] = fmt.Errorf("初始化失败: %v", err)
			log.Printf("[Plugin] %s 初始化失败: %v", name, err)
			continue
		}

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
		p := m.plugins[name]

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

	sorted_plugins := m.sortByPriority()
	for i := len(sorted_plugins) - 1; i >= 0; i-- {
		name := sorted_plugins[i]
		p := m.plugins[name]

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
	for name, p := range m.plugins {
		for _, dep := range p.Dependencies() {
			if _, ok := m.plugins[dep]; !ok {
				return fmt.Errorf("插件 %s 依赖的 %s 不存在", name, dep)
			}
		}
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	for name := range m.plugins {
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

	p, ok := m.plugins[name]
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
	list := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		list = append(list, p)
	}
	return m.topologicalSort(list)
}

// topologicalSort 拓扑排序
func (m *Manager) topologicalSort(list []Plugin) []string {
	in_degree := make(map[string]int)
	adj := make(map[string][]string)

	for _, p := range list {
		name := p.Name()
		in_degree[name] = 0
		adj[name] = []string{}
	}

	for _, p := range list {
		for _, dep := range p.Dependencies() {
			adj[dep] = append(adj[dep], p.Name())
			in_degree[p.Name()]++
		}
	}

	var queue []string
	for name, degree := range in_degree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	sort.Slice(queue, func(i, j int) bool {
		return m.plugins[queue[i]].Priority() < m.plugins[queue[j]].Priority()
	})

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		var next_zero []string
		for _, neighbor := range adj[current] {
			in_degree[neighbor]--
			if in_degree[neighbor] == 0 {
				next_zero = append(next_zero, neighbor)
			}
		}

		sort.Slice(next_zero, func(i, j int) bool {
			return m.plugins[next_zero[i]].Priority() < m.plugins[next_zero[j]].Priority()
		})

		queue = append(queue, next_zero...)
	}

	return result
}

// Count 获取插件数量
func (m *Manager) Count() int {
	return len(m.plugins)
}

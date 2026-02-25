package demo

import (
	"fst/backend/app/plugins"
	"fst/backend/pkg/pluginregistry"
	"fst/backend/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	// 自动注册插件到全局注册表
	pluginregistry.Register(NewPlugin())
}

// DemoPlugin 示例插件
// 展示如何实现完整的插件接口
type DemoPlugin struct {
	plugins.BasePlugin
	config map[string]interface{}
}

// NewPlugin 创建示例插件实例
func NewPlugin() plugins.Plugin {
	p := &DemoPlugin{
		BasePlugin: plugins.NewBasePlugin(
			"demo-plugin",
			"1.0.0",
			"示例插件，展示插件系统的完整功能",
		),
		config: make(map[string]interface{}),
	}

	// 设置优先级（可选）
	p.BasePlugin.SetPriority(100)

	return p
}

// Configure 接收配置
func (p *DemoPlugin) Configure(config map[string]interface{}) error {
	if config != nil {
		p.config = config
		log.Printf("[DemoPlugin] 配置已加载: %v", config)
	}
	return nil
}

// Migrate 数据库迁移
func (p *DemoPlugin) Migrate() error {
	// 示例：创建插件所需的数据表
	// 这里只是演示，实际项目中应该执行真实的数据库迁移
	log.Println("[DemoPlugin] 数据库迁移完成（示例）")
	return nil
}

// Init 初始化插件
func (p *DemoPlugin) Init() error {
	log.Println("[DemoPlugin] 初始化完成")
	return nil
}

// RegisterRoutes 注册路由
func (p *DemoPlugin) RegisterRoutes(router *gin.RouterGroup) {
	// 示例路由
	router.GET("/demo/hello", p.helloHandler)
	router.GET("/demo/info", p.infoHandler)
	router.POST("/demo/echo", p.echoHandler)

	log.Println("[DemoPlugin] 路由注册完成")
}

// Shutdown 关闭插件
func (p *DemoPlugin) Shutdown() error {
	log.Println("[DemoPlugin] 已关闭")
	return nil
}

// ========================================
// 路由处理器
// ========================================

// helloHandler 示例Hello接口
// @Summary Demo插件Hello
// @Description 示例插件的Hello接口
// @Tags Plugin-Demo
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/demo/hello [get]
func (p *DemoPlugin) helloHandler(c *gin.Context) {
	utils.Success(c, gin.H{
		"message": "Hello from Demo Plugin!",
		"version": p.Version(),
	})
}

// infoHandler 获取插件信息
// @Summary Demo插件信息
// @Description 获取Demo插件的详细信息
// @Tags Plugin-Demo
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/demo/info [get]
func (p *DemoPlugin) infoHandler(c *gin.Context) {
	utils.Success(c, gin.H{
		"name":         p.Name(),
		"version":      p.Version(),
		"description":  p.Description(),
		"priority":     p.Priority(),
		"dependencies": p.Dependencies(),
	})
}

// echoHandler Echo接口
// @Summary Demo插件Echo
// @Description 回显请求数据
// @Tags Plugin-Demo
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "请求数据"
// @Success 200 {object} utils.Response
// @Router /api/v1/demo/echo [post]
func (p *DemoPlugin) echoHandler(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Fail(c, 400, "请求体格式错误")
		return
	}

	utils.Success(c, gin.H{
		"echo":   body,
		"config": p.config,
	})
}

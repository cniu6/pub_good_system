package demo

import (
	"fst/backend/app/plugins"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DemoPlugin struct{}

func (p *DemoPlugin) Name() string {
	return "demo-plugin"
}

func (p *DemoPlugin) Version() string {
	return "1.0.0"
}

func (p *DemoPlugin) Init() error {
	log.Println("Demo plugin initialized")
	return nil
}

func (p *DemoPlugin) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/demo/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from Demo Plugin!"})
	})
}

func NewPlugin() plugins.Plugin {
	return &DemoPlugin{}
}

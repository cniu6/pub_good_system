package plugins_test

import (
	"testing"

	"fst/backend/app/plugins"

	"github.com/gin-gonic/gin"
)

type stubPlugin struct {
	plugins.BasePlugin
}

func newStub(name string) *stubPlugin {
	return &stubPlugin{BasePlugin: plugins.NewBasePlugin(name, "0.0.1", "stub")}
}

func TestPluginManager_RegisterLoadShutdown(t *testing.T) {
	m := plugins.NewManager()
	m.Register(newStub("stub-a"))
	m.Register(newStub("stub-b"))

	if m.Count() != 2 {
		t.Fatalf("Count=%d want 2", m.Count())
	}
	if err := m.LoadAll(); err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/plugins")
	m.RegisterAllRoutes(g)

	if err := m.ShutdownAll(); err != nil {
		t.Fatalf("ShutdownAll: %v", err)
	}
}

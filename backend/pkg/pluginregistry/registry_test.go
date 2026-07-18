package pluginregistry_test

import (
	"testing"

	"fst/backend/pkg/pluginregistry"
)

type mockPlugin struct {
	name, version string
}

func (m mockPlugin) Name() string    { return m.name }
func (m mockPlugin) Version() string { return m.version }

func TestRegisterGetAllClear(t *testing.T) {
	pluginregistry.Clear()
	defer pluginregistry.Clear()

	if pluginregistry.Count() != 0 {
		t.Fatalf("清空后 Count=%d", pluginregistry.Count())
	}

	pluginregistry.Register(mockPlugin{name: "demo-a", version: "1.0.0"})
	pluginregistry.Register(mockPlugin{name: "demo-b", version: "2.0.0"})
	if pluginregistry.Count() != 2 {
		t.Fatalf("Count=%d want 2", pluginregistry.Count())
	}

	all := pluginregistry.GetAll()
	if len(all) != 2 {
		t.Fatalf("GetAll len=%d", len(all))
	}

	// 同名覆盖
	pluginregistry.Register(mockPlugin{name: "demo-a", version: "1.1.0"})
	if pluginregistry.Count() != 2 {
		t.Fatalf("覆盖后 Count 应变 2，实际 %d", pluginregistry.Count())
	}
	found := false
	for _, p := range pluginregistry.GetAll() {
		if p.Name() == "demo-a" && p.Version() == "1.1.0" {
			found = true
		}
	}
	if !found {
		t.Fatal("同名插件未覆盖版本")
	}
}

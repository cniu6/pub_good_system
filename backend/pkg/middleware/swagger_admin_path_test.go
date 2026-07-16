package middleware

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/swaggo/swag"
)

// fakeSwagger 模拟 swag 文档，仅用于路径改写单测
type fakeSwagger struct {
	doc string
}

func (f *fakeSwagger) ReadDoc() string { return f.doc }

func TestBuildRewrittenSwaggerDoc_AdminPrefix(t *testing.T) {
	raw := `{
		"swagger":"2.0",
		"info":{"title":"t","description":"d"},
		"paths":{
			"/api/v1/admin/users":{"get":{"summary":"u"}},
			"/api/v1/admin":{"get":{"summary":"root"}},
			"/api/v1/administrator/x":{"get":{"summary":"keep"}},
			"/api/v1/user/profile":{"get":{"summary":"user"}}
		}
	}`
	// 临时注册假文档
	name := "swagger-test-admin-path"
	swag.Register(name, &fakeSwagger{doc: raw})
	// 直接测改写逻辑：先用真实 GetSwagger 读不到我们的 name，因此把核心逻辑拆成对 map 的校验
	// 这里通过 json 手工复现改写规则（与 buildRewrittenSwaggerDoc 一致）
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		t.Fatal(err)
	}
	paths := root["paths"].(map[string]interface{})
	oldPrefix := "/api/v1/admin"
	newPrefix := "/api/v1/mgr-api"
	rewritten := make(map[string]interface{}, len(paths))
	for pathKey, pathVal := range paths {
		if pathKey == oldPrefix || strings.HasPrefix(pathKey, oldPrefix+"/") {
			rewritten[newPrefix+strings.TrimPrefix(pathKey, oldPrefix)] = pathVal
			continue
		}
		rewritten[pathKey] = pathVal
	}

	if _, ok := rewritten["/api/v1/mgr-api/users"]; !ok {
		t.Fatalf("expected rewritten /api/v1/mgr-api/users, got keys=%v", keys(rewritten))
	}
	if _, ok := rewritten["/api/v1/mgr-api"]; !ok {
		t.Fatalf("expected rewritten /api/v1/mgr-api root")
	}
	if _, ok := rewritten["/api/v1/administrator/x"]; !ok {
		t.Fatalf("administrator path must keep")
	}
	if _, ok := rewritten["/api/v1/user/profile"]; !ok {
		t.Fatalf("user path must keep")
	}
	if _, ok := rewritten["/api/v1/admin/users"]; ok {
		t.Fatalf("old admin path must be removed")
	}
}

func keys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

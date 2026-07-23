package docs_test

import (
	"encoding/json"
	"strings"
	"testing"

	"fst/backend/docs"
)

func TestSwaggerInfoPresent(t *testing.T) {
	if docs.SwaggerInfo == nil {
		t.Fatal("SwaggerInfo 不应为 nil")
	}
	if strings.TrimSpace(docs.SwaggerInfo.Title) == "" {
		t.Fatal("SwaggerInfo.Title 不应为空")
	}
	if strings.TrimSpace(docs.SwaggerInfo.Version) == "" {
		t.Fatal("SwaggerInfo.Version 不应为空")
	}
	if strings.TrimSpace(docs.SwaggerInfo.BasePath) == "" {
		t.Fatal("SwaggerInfo.BasePath 不应为空")
	}

	raw := docs.SwaggerInfo.ReadDoc()
	if strings.TrimSpace(raw) == "" {
		t.Fatal("Swagger ReadDoc 不应返回空文档")
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("Swagger JSON 无法解析: %v", err)
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		t.Fatal("Swagger paths 应包含至少一个路由")
	}
}

package docs_test

import (
	"testing"

	"fst/backend/docs"
)

func TestSwaggerInfoPresent(t *testing.T) {
	if docs.SwaggerInfo == nil {
		t.Fatal("SwaggerInfo 不应为 nil")
	}
	if docs.SwaggerInfo.Title == "" && docs.SwaggerInfo.Version == "" {
		// 生成文件通常有 Title/Version；至少保证结构可访问
		t.Logf("SwaggerInfo=%+v", docs.SwaggerInfo)
	}
}

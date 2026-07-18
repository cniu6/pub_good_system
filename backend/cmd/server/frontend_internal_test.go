package server

import (
	"net/http"
	"testing"
)

func TestRequestAcceptsGzip(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	if requestAcceptsGzip(req) {
		t.Fatal("无 Accept-Encoding 不应接受 gzip")
	}
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	if !requestAcceptsGzip(req) {
		t.Fatal("应识别 gzip")
	}
}

func TestIsStaticAssetRequest(t *testing.T) {
	if !isStaticAssetRequest("assets/app.js") {
		t.Fatal("js 应视为静态资源")
	}
	if isStaticAssetRequest("index") {
		t.Fatal("无扩展名不应视为静态资源")
	}
}

func TestFrontendContentType(t *testing.T) {
	ct := frontendContentType("app.css")
	if ct == "" {
		t.Fatal("content-type 不应为空")
	}
}

func TestOpenExternalDist(t *testing.T) {
	fsys := openExternalDist()
	if fsys == nil {
		t.Fatal("openExternalDist 不应返回 nil")
	}
}

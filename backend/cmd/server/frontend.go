package server

import (
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// openExternalDist 使用磁盘上的 ./dist 目录作为前端资源根（external 打包模式）。
func openExternalDist() fs.FS {
	return os.DirFS("dist")
}

func requestAcceptsGzip(r *http.Request) bool {
	if r == nil {
		return false
	}
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(strings.ToLower(acceptEncoding), "gzip")
}

func isStaticAssetRequest(relPath string) bool {
	return strings.TrimSpace(filepath.Ext(relPath)) != ""
}

func frontendContentType(relPath string) string {
	cleanPath := strings.TrimSuffix(relPath, ".gz")
	contentType := mime.TypeByExtension(filepath.Ext(cleanPath))
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func serveFrontendFile(c *gin.Context, rawFS fs.FS, relPath string) bool {
	if c == nil || rawFS == nil {
		return false
	}

	// 优先尝试预压缩的 .gz 文件（构建产物常见）
	if requestAcceptsGzip(c.Request) {
		gzipPath := relPath + ".gz"
		if data, err := fs.ReadFile(rawFS, gzipPath); err == nil {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
			c.Data(http.StatusOK, frontendContentType(gzipPath), data)
			return true
		}
	}

	data, err := fs.ReadFile(rawFS, relPath)
	if err != nil {
		return false
	}

	c.Data(http.StatusOK, frontendContentType(relPath), data)
	return true
}

// registerFrontendNoRoute 注册 SPA 回退：
// 非 /api 路径先找静态文件；带扩展名找不到 → 404；否则回退 index.html。
func registerFrontendNoRoute(router *gin.Engine, rawFS fs.FS) {
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			utils.Fail(c, 404, "API not found")
			return
		}

		relPath := strings.TrimPrefix(path, "/")
		if relPath != "" {
			if serveFrontendFile(c, rawFS, relPath) {
				return
			}

			// 带扩展名的静态资源找不到 → 404，不回退到 index.html
			if isStaticAssetRequest(relPath) {
				c.Status(http.StatusNotFound)
				return
			}
		}

		if !serveFrontendFile(c, rawFS, "index.html") {
			log.Printf("[Frontend] index.html 读取失败")
			c.Status(http.StatusNotFound)
			return
		}

		log.Printf("[Frontend] 返回 index.html for path: %s", path)
	})
}

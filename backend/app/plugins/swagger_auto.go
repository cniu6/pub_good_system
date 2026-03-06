//go:build !prod
// +build !prod

package plugins

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	swaggerOnce     sync.Once
	swaggerLastGen  int64
	swaggerLockFile string
)

func init() {
	// 在开发模式下自动更新 Swagger
	swaggerOnce.Do(func() {
		if os.Getenv("GO_ENV") == "production" || os.Getenv("SKIP_AUTO_SWAGGER") == "true" {
			return
		}

		go autoUpdateSwagger()
	})
}

// autoUpdateSwagger 自动更新 Swagger 文档
func autoUpdateSwagger() {
	// 获取项目路径
	projectDir, backendDir := getProjectPaths()
	if projectDir == "" {
		return
	}

	swaggerLockFile = filepath.Join(backendDir, "docs", ".swagger-lock")

	// 检查是否需要更新
	if !needSwaggerUpdate(projectDir, backendDir) {
		return
	}

	log.Println("[Swagger] 自动更新文档...")

	// 1. 运行插件生成器
	if err := runPluginGenerator(projectDir); err != nil {
		log.Printf("[Swagger] 插件生成失败: %v", err)
	}

	// 2. 运行 swag init
	if err := runSwagInit(backendDir); err != nil {
		log.Printf("[Swagger] 文档生成失败: %v", err)
		return
	}

	// 3. 更新锁文件
	os.WriteFile(swaggerLockFile, []byte(time.Now().Format(time.RFC3339)), 0644)

	log.Println("[Swagger] 文档已更新")
}

// getProjectPaths 获取项目路径
func getProjectPaths() (projectDir, backendDir string) {
	// 方法1: 通过调用者路径推断
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		pluginsDir := filepath.Dir(filename)
		backendDir = filepath.Dir(pluginsDir)
		projectDir = filepath.Dir(backendDir)
		return projectDir, backendDir
	}

	// 方法2: 通过工作目录
	wd, err := os.Getwd()
	if err != nil {
		return "", ""
	}

	// 检查是否在 backend 目录
	if filepath.Base(wd) == "backend" {
		return filepath.Dir(wd), wd
	}

	// 检查是否有 backend 子目录
	backendPath := filepath.Join(wd, "backend")
	if _, err := os.Stat(backendPath); err == nil {
		return wd, backendPath
	}

	return "", ""
}

// needSwaggerUpdate 检查是否需要更新 Swagger
func needSwaggerUpdate(projectDir, backendDir string) bool {
	// 获取插件目录修改时间
	pluginsDir := filepath.Join(backendDir, "app", "plugins")
	pluginsModTime := getDirModTime(pluginsDir)

	// 获取控制器目录修改时间
	controllersDir := filepath.Join(backendDir, "app", "controllers")
	controllersModTime := getDirModTime(controllersDir)

	// 取最新的修改时间
	latestModTime := pluginsModTime
	if controllersModTime > latestModTime {
		latestModTime = controllersModTime
	}

	// 获取锁文件时间
	lockTime := int64(0)
	if info, err := os.Stat(swaggerLockFile); err == nil {
		lockTime = info.ModTime().Unix()
	}

	// 如果代码比锁文件新，需要更新
	return latestModTime > lockTime
}

// getDirModTime 获取目录最新修改时间
func getDirModTime(dir string) int64 {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	var latest int64
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if entry.IsDir() {
			subTime := getDirModTime(fullPath)
			if subTime > latest {
				latest = subTime
			}
		} else {
			if info.ModTime().Unix() > latest {
				latest = info.ModTime().Unix()
			}
		}
	}
	return latest
}

// runPluginGenerator 运行插件生成器
func runPluginGenerator(projectDir string) error {
	genPath := filepath.Join("backend", "app", "plugins", "gen_plugins.go")

	cmd := exec.Command("go", "run", genPath)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runSwagInit 运行 swag init
func runSwagInit(backendDir string) error {
	// 检查 swag 是否安装
	swagPath, err := exec.LookPath("swag")
	if err != nil {
		log.Println("[Swagger] swag 未安装，跳过自动生成")
		log.Println("[Swagger] 安装命令: go install github.com/swaggo/swag/cmd/swag@latest")
		return nil
	}

	args := []string{
		"init",
		"-g", "cmd/main.go",
		"-o", "docs",
		"--parseDependency",
		"--parseInternal",
		"--quiet",
	}

	cmd := exec.Command(swagPath, args...)
	cmd.Dir = backendDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetSwaggerStatus 获取 Swagger 状态
func GetSwaggerStatus() map[string]interface{} {
	projectDir, backendDir := getProjectPaths()
	if projectDir == "" {
		return map[string]interface{}{"error": "无法获取项目路径"}
	}

	pluginsDir := filepath.Join(backendDir, "app", "plugins")
	pluginsModTime := time.Unix(getDirModTime(pluginsDir), 0)

	lockContent, _ := os.ReadFile(swaggerLockFile)
	lockTime := strings.TrimSpace(string(lockContent))

	// 统计插件数量
	pluginCount := 0
	entries, _ := os.ReadDir(pluginsDir)
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			pluginFile := filepath.Join(pluginsDir, e.Name(), "plugin.go")
			if _, err := os.Stat(pluginFile); err == nil {
				pluginCount++
			}
		}
	}

	return map[string]interface{}{
		"plugins_dir":     pluginsDir,
		"plugins_mod":     pluginsModTime.Format(time.RFC3339),
		"last_gen":        lockTime,
		"need_update":     needSwaggerUpdate(projectDir, backendDir),
		"plugin_count":    pluginCount,
		"swagger_enabled": os.Getenv("GO_ENV") != "production",
	}
}

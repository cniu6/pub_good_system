//go:build ignore
// +build ignore

// Plugin Import Generator
// 自动扫描插件目录并更新 main.go 中的插件导入
//
// 使用方法:
//   go run backend/app/plugins/gen_plugins.go

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	wd, _ := os.Getwd()

	// 查找插件目录
	pluginDir := ""
	candidates := []string{
		filepath.Join(wd, "backend", "app", "plugins"),
		filepath.Join(wd, "app", "plugins"),
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			pluginDir = c
			break
		}
	}

	if pluginDir == "" {
		fmt.Fprintln(os.Stderr, "[Plugin] 无法找到插件目录")
		os.Exit(1)
	}

	fmt.Printf("[Plugin] 扫描插件目录: %s\n", pluginDir)

	// 扫描插件
	entries, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Plugin] 读取目录失败: %v\n", err)
		os.Exit(1)
	}

	var plugins []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}

		pluginFile := filepath.Join(pluginDir, name, "plugin.go")
		content, err := ioutil.ReadFile(pluginFile)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "NewPlugin()") && strings.Contains(contentStr, "pluginregistry.Register") {
			plugins = append(plugins, name)
			fmt.Printf("[Plugin]   发现: %s\n", name)
		}
	}

	if len(plugins) == 0 {
		fmt.Println("[Plugin] 未发现任何插件")
	}

	// 查找并更新 main.go
	mainFile := filepath.Join(wd, "backend", "cmd", "main.go")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		mainFile = filepath.Join(filepath.Dir(pluginDir), "cmd", "main.go")
	}

	content, err := ioutil.ReadFile(mainFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Plugin] 无法读取 main.go: %v\n", err)
		os.Exit(1)
	}

	// 构建新的插件导入
	var imports strings.Builder
	for _, name := range plugins {
		imports.WriteString(fmt.Sprintf("\t_ \"fst/backend/app/plugins/%s\"\n", name))
	}

	// 使用正则替换插件导入区域
	pattern := regexp.MustCompile(`(?s)// @plugins-start\n(.*?)// @plugins-end`)
	newContent := pattern.ReplaceAllString(string(content), fmt.Sprintf("// @plugins-start\n%s\t// @plugins-end", imports.String()))

	// 检查是否有变化
	if newContent == string(content) {
		fmt.Println("[Plugin] main.go 无需更新")
		return
	}

	// 写入文件
	if err := ioutil.WriteFile(mainFile, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "[Plugin] 写入 main.go 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[Plugin] 已更新 main.go，共 %d 个插件\n", len(plugins))
}

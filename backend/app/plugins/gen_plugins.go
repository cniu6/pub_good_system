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
		// 跳过隐藏目录、下划线目录，以及 demo（需 -tags demo，由 main_demo_plugins.go 单独引入）
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "demo" {
			continue
		}

		pluginFile := filepath.Join(pluginDir, name, "plugin.go")
		content, err := ioutil.ReadFile(pluginFile)
		if err != nil {
			continue
		}

		contentStr := string(content)
		// 带有非默认 build tag 的插件不写入 main.go，避免普通编译失败
		if hasNonDefaultBuildTag(contentStr) {
			fmt.Printf("[Plugin]   跳过(build tag): %s\n", name)
			continue
		}
		if strings.Contains(contentStr, "NewPlugin()") && strings.Contains(contentStr, "pluginregistry.Register") {
			plugins = append(plugins, name)
			fmt.Printf("[Plugin]   发现: %s\n", name)
		}
	}

	if len(plugins) == 0 {
		fmt.Println("[Plugin] 未发现任何插件")
	}

	// 查找并更新根入口 main.go（插件 blank import 区域）
	mainFile := filepath.Join(wd, "main.go")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		// 兼容从 backend/ 子目录运行的情况
		mainFile = filepath.Join(filepath.Dir(wd), "main.go")
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

// hasNonDefaultBuildTag 检测 plugin.go 是否带有限制性 build tag（如 demo / ignore）。
// 这类插件不应写入默认 main.go 的 blank import。
func hasNonDefaultBuildTag(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// 只看文件头部约束，遇到 package 行即停止
		if strings.HasPrefix(trimmed, "package ") {
			break
		}
		if strings.HasPrefix(trimmed, "//go:build ") || strings.HasPrefix(trimmed, "// +build ") {
			lower := strings.ToLower(trimmed)
			// ignore 本身表示本文件是生成器，不是插件；其它条件 tag 一律跳过
			if strings.Contains(lower, "ignore") {
				return true
			}
			// 纯否定条件如 !prod 仍可被默认构建收录；含有正向可选 tag（无 !）则跳过
			constraint := strings.TrimPrefix(trimmed, "//go:build ")
			constraint = strings.TrimPrefix(constraint, "// +build ")
			for _, tok := range strings.Fields(strings.ReplaceAll(constraint, ",", " ")) {
				tok = strings.TrimSpace(tok)
				if tok == "" || tok == "||" || tok == "&&" {
					continue
				}
				if !strings.HasPrefix(tok, "!") {
					return true
				}
			}
		}
	}
	return false
}


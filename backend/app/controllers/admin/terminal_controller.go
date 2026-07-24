package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	terminalTimeout   = 15 * time.Second
	terminalMaxOutput = 256 * 1024 // 256KB
)

// TerminalController 调试终端（仅 IsAdminDebugOpsEnabled 时注册）
type TerminalController struct{}

func NewTerminalController() *TerminalController {
	return &TerminalController{}
}

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // 已挂管理端 JWT + AdminOnly
	},
}

// 危险命令模式（子串，不区分大小写）
var dangerousCmdPatterns = []string{
	"rm -rf /",
	"rm -rf /*",
	"mkfs",
	"format ",
	"shutdown",
	"reboot",
	"dd if=",
	":(){", // fork bomb
	"fork(",
	">/dev/sd",
	"del /f /s /q",
}

func isDangerousCmd(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, p := range dangerousCmdPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func writeTerminalAudit(c *gin.Context, cmd, output string, status int) {
	uid, name := adminAuditUser(c)
	rb := truncateStr(cmd, 2000)
	sb := truncateStr(output, 4000)
	_ = models.CreateOperationLog(&models.OperationLog{
		UserID:       uid,
		Username:     name,
		Module:       "调试终端",
		Action:       "exec",
		Method:       c.Request.Method,
		Path:         c.FullPath(),
		IP:           utils.GetClientIP(c),
		UserAgent:    c.Request.UserAgent(),
		HandlerName:  "admin.(*TerminalController).exec",
		RequestBody:  &rb,
		ResponseBody: &sb,
		StatusCode:   status,
	})
}

// buildShellCommand 按平台构造 shell 命令（Windows: cmd /C；Unix: bash -lc 或 sh -c）
func buildShellCommand(ctx context.Context, cmdLine string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		// 先切 UTF-8 代码页；仍乱码时由 decodeTerminalOutput 按 GB18030 兜底
		return exec.CommandContext(ctx, "cmd", "/C", "chcp 65001 >nul & "+cmdLine)
	}
	if path, err := exec.LookPath("bash"); err == nil && path != "" {
		return exec.CommandContext(ctx, "bash", "-lc", cmdLine)
	}
	return exec.CommandContext(ctx, "sh", "-c", cmdLine)
}

func runShellCommand(cmdLine string) (string, int, error) {
	cmdLine = strings.TrimSpace(cmdLine)
	if cmdLine == "" {
		return "", 400, fmt.Errorf("命令不能为空")
	}
	if len(cmdLine) > 4000 {
		return "", 400, fmt.Errorf("命令过长")
	}
	if isDangerousCmd(cmdLine) {
		return "", 400, fmt.Errorf("禁止执行危险命令")
	}

	ctx, cancel := context.WithTimeout(context.Background(), terminalTimeout)
	defer cancel()

	cmd := buildShellCommand(ctx, cmdLine)
	var buf bytes.Buffer
	limited := &limitedWriter{w: &buf, limit: terminalMaxOutput}
	cmd.Stdout = limited
	cmd.Stderr = limited

	err := cmd.Run()
	// Windows cmd 默认 GBK/GB18030；按 UTF-8 硬转会变成 � 乱码
	out := decodeTerminalOutput(buf.Bytes())
	if limited.truncated {
		out += "\n…[output truncated at 256KB]"
	}
	if ctx.Err() == context.DeadlineExceeded {
		return out, 408, fmt.Errorf("执行超时（15s）")
	}
	if err != nil {
		// 仍返回输出，便于前端展示 stderr
		return out, 200, err
	}
	return out, 200, nil
}

// decodeTerminalOutput 将命令输出规范为 UTF-8 字符串。
// 中文 Windows 下 cmd 输出多为 GBK；非法 UTF-8 时按 GB18030 解码（兼容 GBK）。
func decodeTerminalOutput(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if utf8.Valid(b) {
		return string(b)
	}
	if runtime.GOOS == "windows" {
		if decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(b); err == nil && utf8.Valid(decoded) {
			return string(decoded)
		}
	}
	// 兜底：替换非法字节，避免再走 string([]rune) 把整段弄成 �
	return strings.ToValidUTF8(string(b), "\uFFFD")
}

type limitedWriter struct {
	w         io.Writer
	limit     int
	written   int
	truncated bool
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	if lw.written >= lw.limit {
		lw.truncated = true
		return len(p), nil
	}
	remain := lw.limit - lw.written
	if len(p) > remain {
		lw.truncated = true
		n, err := lw.w.Write(p[:remain])
		lw.written += n
		return len(p), err
	}
	n, err := lw.w.Write(p)
	lw.written += n
	return n, err
}

type terminalExecRequest struct {
	Cmd string `json:"cmd"`
}

// ExecHTTP POST /debug/terminal/exec
func (ctrl *TerminalController) ExecHTTP(c *gin.Context) {
	var req terminalExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	out, status, err := runShellCommand(req.Cmd)
	exitMsg := ""
	if err != nil {
		exitMsg = err.Error()
	}
	writeTerminalAudit(c, req.Cmd, out+"\n"+exitMsg, status)
	if status == 400 || status == 408 {
		utils.Fail(c, status, exitMsg)
		return
	}
	utils.Success(c, gin.H{
		"output":  out,
		"error":   exitMsg,
		"success": err == nil,
	})
}

// HandleWS GET /debug/terminal — WebSocket 终端
func (ctrl *TerminalController) HandleWS(c *gin.Context) {
	conn, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ADMIN][Terminal] websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	_ = conn.WriteJSON(gin.H{"type": "ready", "message": "terminal ready"})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		cmdLine := parseTerminalMessage(data)
		if cmdLine == "" {
			_ = conn.WriteJSON(gin.H{"type": "error", "message": "empty command"})
			continue
		}
		out, status, runErr := runShellCommand(cmdLine)
		exitMsg := ""
		if runErr != nil {
			exitMsg = runErr.Error()
		}
		writeTerminalAudit(c, cmdLine, out+"\n"+exitMsg, status)
		_ = conn.WriteJSON(gin.H{
			"type":    "result",
			"cmd":     cmdLine,
			"output":  out,
			"error":   exitMsg,
			"success": runErr == nil && status == 200,
		})
	}
}

func parseTerminalMessage(data []byte) string {
	s := strings.TrimSpace(string(data))
	if s == "" {
		return ""
	}
	var msg struct {
		Type string `json:"type"`
		Cmd  string `json:"cmd"`
	}
	if err := json.Unmarshal(data, &msg); err == nil && (msg.Type == "exec" || msg.Cmd != "") {
		return strings.TrimSpace(msg.Cmd)
	}
	// 纯文本一行
	return s
}

// RegisterRoutes 仅在调试运维开关开启时注册
func (ctrl *TerminalController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	if !config.IsAdminDebugOpsEnabled() {
		return
	}
	debug := adminGroup.Group("/debug")
	{
		debug.GET("/terminal", ctrl.HandleWS)
		debug.POST("/terminal/exec", ctrl.ExecHTTP)
	}
}

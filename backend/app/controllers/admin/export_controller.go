package admin

import (
	"encoding/csv"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportController 导入导出中心 MVP
type ExportController struct{}

func NewExportController() *ExportController {
	return &ExportController{}
}

func writeExportAudit(c *gin.Context, action, detail string) {
	uid, _ := c.Get("userID")
	uname, _ := c.Get("username")
	userID, _ := uid.(uint64)
	username, _ := uname.(string)
	now := time.Now().Unix()
	body := detail
	item := &models.OperationLog{
		UserID:      userID,
		Username:    username,
		Module:      "导入导出",
		Action:      action,
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		IP:          utils.GetClientIP(c),
		UserAgent:   c.GetHeader("User-Agent"),
		HandlerName: "ExportController",
		RequestBody: &body,
		StatusCode:  200,
		Duration:    0,
		CreateTime:  &now,
	}
	if err := models.CreateOperationLog(item); err != nil {
		log.Printf("[ADMIN][EXPORT] audit log failed: %v", err)
	}
}

// ExportUsers POST /export/users — 导出用户 CSV（流式响应）
func (ctrl *ExportController) ExportUsers(c *gin.Context) {
	type row struct {
		ID       uint64  `gorm:"column:id"`
		Username string  `gorm:"column:username"`
		Nickname string  `gorm:"column:nickname"`
		Email    string  `gorm:"column:email"`
		Mobile   string  `gorm:"column:mobile"`
		Role     string  `gorm:"column:role"`
		Status   uint8   `gorm:"column:status"`
		Money    float64 `gorm:"column:money"`
		Score    int64   `gorm:"column:score"`
	}
	var rows []row
	err := db.DB.Raw(`
		SELECT id, username, nickname, email, mobile, role, status, money, score
		FROM users WHERE delete_time IS NULL ORDER BY id ASC LIMIT 10000`).Scan(&rows).Error
	if err != nil {
		log.Printf("[ADMIN][EXPORT] query users failed: %v", err)
		utils.Fail(c, 500, "导出失败")
		return
	}

	filename := fmt.Sprintf("users_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	// Excel 识别 UTF-8 BOM
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"id", "username", "nickname", "email", "mobile", "role", "status", "money", "score"})
	for _, r := range rows {
		_ = w.Write([]string{
			strconv.FormatUint(r.ID, 10),
			r.Username,
			r.Nickname,
			r.Email,
			r.Mobile,
			r.Role,
			strconv.Itoa(int(r.Status)),
			fmt.Sprintf("%.2f", r.Money),
			strconv.FormatInt(r.Score, 10),
		})
	}
	w.Flush()
	writeExportAudit(c, "export_users", fmt.Sprintf("count=%d", len(rows)))
}

// DownloadUserTemplate GET /export/users/template — 用户导入模板
func (ctrl *ExportController) DownloadUserTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="users_import_template.csv"`)
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"username", "nickname", "email", "mobile", "password", "role"})
	_ = w.Write([]string{"demo_user", "演示用户", "demo@example.com", "13800138000", "ChangeMe123", "user"})
	w.Flush()
	writeExportAudit(c, "download_user_template", "")
}

// RegisterRoutes 注册导出路由
func (ctrl *ExportController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	exp := adminGroup.Group("/export")
	exp.Use(middleware.SimpleLogMiddleware("导入导出"))
	exp.Use(middleware.RequirePermission("user:read"))
	{
		// 导出含手机等 PII：已启用 TOTP 的管理员须带头 X-Totp-Code
		exp.POST("/users", middleware.RequireRecentTOTP(), ctrl.ExportUsers)
		exp.GET("/users/template", ctrl.DownloadUserTemplate)
	}
}

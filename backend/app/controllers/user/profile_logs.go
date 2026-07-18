package user

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================
// 用户统计
// ========================================

// GetUserStats 获取用户统计
// @Summary 获取用户统计
// @Description 获取当前用户的统计数据
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/stats [get]
func (ctrl *ProfileController) GetUserStats(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 获取登录次数
	login_count, _ := models.GetUserLoginCount(user.ID)

	utils.Success(c, gin.H{
		"joinTime":      user.JoinTime,
		"lastLoginTime": user.LastLoginTime,
		"lastLoginIp":   user.LastLoginIp,
		"loginCount":    login_count,
		"daysJoined":    calculateDaysJoined(user.JoinTime),
		"money":         user.Money,
		"score":         user.Score,
		"level":         user.Level,
	})
}

// ========================================
// 余额/积分日志（用户只能查看自己的）
// ========================================

// GetMoneyLogs 获取当前用户的余额变动日志
// @Summary 获取我的余额变动日志
// @Tags 用户中心
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/money-logs [get]
func (ctrl *ProfileController) GetMoneyLogs(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.DefaultQuery("keyword", "")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := services.GetUserMoneyLogList(uid, page, pageSize, keyword)
	if err != nil {
		log.Printf("[PROFILE] load money logs failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "获取余额日志失败")
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// GetScoreLogs 获取当前用户的积分变动日志
// @Summary 获取我的积分变动日志
// @Tags 用户中心
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/score-logs [get]
func (ctrl *ProfileController) GetScoreLogs(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.DefaultQuery("keyword", "")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := services.GetUserScoreLogList(uid, page, pageSize, keyword)
	if err != nil {
		log.Printf("[PROFILE] load score logs failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "获取积分日志失败")
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// ========================================
// 用户仪表盘
// ========================================

// GetDashboard 获取用户仪表盘数据
// @Summary 获取用户仪表盘
// @Description 返回用户统计概览、公告、快捷操作入口
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/dashboard [get]
func (ctrl *ProfileController) GetDashboard(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}
	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	login_count, _ := models.GetUserLoginCount(uid)

	// 公告列表（可扩展为从数据库读取）
	announcements := []gin.H{
		{"id": 1, "type": "info", "title": "系统维护通知", "content": "系统将于本周六凌晨进行维护升级", "time": time.Now().Unix()},
		{"id": 2, "type": "success", "title": "新功能上线", "content": "用户中心新增设备管理和账号安全功能", "time": time.Now().Unix()},
		{"id": 3, "type": "warning", "title": "安全提醒", "content": "请定期修改密码以保障账号安全", "time": time.Now().Unix()},
	}

	utils.Success(c, gin.H{
		"user": gin.H{
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"email":    user.Email,
			"role":     user.Role,
			"level":    user.Level,
		},
		"stats": gin.H{
			"money":      user.Money,
			"score":      user.Score,
			"level":      user.Level,
			"loginCount": login_count,
			"daysJoined": calculateDaysJoined(user.JoinTime),
		},
		"announcements": announcements,
	})
}

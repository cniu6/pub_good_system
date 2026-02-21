package controllers

import (
	"fst/backend/app/services"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

type SystemController struct{}

func (ctrl *SystemController) GetUserPage(c *gin.Context) {
	utils.Success(c, []interface{}{})
}

// GetCleanupStatus 查询验证码清理任务的运行状态
// @Summary 获取清理任务状态
// @Description 返回验证码清理任务的运行状态、间隔、上次/下次执行时间
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system/cleanup-status [get]
func (ctrl *SystemController) GetCleanupStatus(c *gin.Context) {
	utils.Success(c, services.GetCleanupStatus())
}

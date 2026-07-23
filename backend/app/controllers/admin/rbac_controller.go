package admin

import (
	"fst/backend/app/models"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RBACController 管理端 RBAC（MVP：列表 + 分配角色）
type RBACController struct{}

func NewRBACController() *RBACController {
	return &RBACController{}
}

// ListRoles GET /roles
func (ctrl *RBACController) ListRoles(c *gin.Context) {
	list, err := models.ListRoles()
	if err != nil {
		log.Printf("[ADMIN][RBAC] list roles failed: %v", err)
		utils.Fail(c, 500, "获取角色失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// ListPermissions GET /permissions
func (ctrl *RBACController) ListPermissions(c *gin.Context) {
	list, err := models.ListPermissions()
	if err != nil {
		log.Printf("[ADMIN][RBAC] list permissions failed: %v", err)
		utils.Fail(c, 500, "获取权限失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// AssignUserRoleRequest 分配角色请求
type AssignUserRoleRequest struct {
	RoleID   *uint64 `json:"role_id"`
	RoleCode string  `json:"role_code"`
}

// AssignUserRole POST /users/:id/roles
func (ctrl *RBACController) AssignUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}
	var req AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	var role *models.Role
	if req.RoleID != nil && *req.RoleID > 0 {
		role, err = models.GetRoleByID(*req.RoleID)
	} else if strings.TrimSpace(req.RoleCode) != "" {
		role, err = models.GetRoleByCode(strings.TrimSpace(req.RoleCode))
	} else {
		utils.Fail(c, 400, "请提供 role_id 或 role_code")
		return
	}
	if err != nil || role == nil {
		utils.Fail(c, 404, "角色不存在")
		return
	}

	if _, err := models.GetUserByID(userID); err != nil {
		utils.Fail(c, 404, "用户不存在")
		return
	}

	if err := models.AssignUserRole(userID, role.ID); err != nil {
		log.Printf("[ADMIN][RBAC] assign role failed user_id=%d role_id=%d: %v", userID, role.ID, err)
		utils.Fail(c, 500, "分配角色失败")
		return
	}
	utils.Success(c, gin.H{"user_id": userID, "role": role})
}

// ListUserRoles GET /users/:id/roles
func (ctrl *RBACController) ListUserRoles(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}
	list, err := models.ListUserRoles(userID)
	if err != nil {
		utils.Fail(c, 500, "获取用户角色失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// RegisterRoutes 注册 RBAC 路由
func (ctrl *RBACController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	adminGroup.GET("/roles", ctrl.ListRoles)
	adminGroup.GET("/permissions", ctrl.ListPermissions)

	users := adminGroup.Group("/users")
	users.Use(middleware.SimpleLogMiddleware("RBAC角色"))
	{
		users.GET("/:id/roles", ctrl.ListUserRoles)
		users.POST("/:id/roles", ctrl.AssignUserRole)
	}
}

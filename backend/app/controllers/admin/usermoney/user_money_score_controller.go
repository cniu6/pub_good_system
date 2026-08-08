package usermoney

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserMoneyScoreController 管理员 - 用户余额/积分管理
type UserMoneyScoreController struct{}

// NewUserMoneyScoreController 创建控制器
func NewUserMoneyScoreController() *UserMoneyScoreController {
	return &UserMoneyScoreController{}
}

// ========================================
// 余额管理
// ========================================

// ChangeMoney 变更用户余额（增减）
// @Summary 变更用户余额
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/money/change [post]
func (ctrl *UserMoneyScoreController) ChangeMoney(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Money *float64 `json:"money" binding:"required"`
		Memo  string   `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if req.Money == nil {
		utils.Fail(c, 400, "金额不能为空")
		return
	}

	logEntry, err := services.ChangeUserMoney(userID, *req.Money, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "余额变更成功", "log": logEntry})
}

// SetMoney 直接设置用户余额
// @Summary 直接设置用户余额
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/money [put]
func (ctrl *UserMoneyScoreController) SetMoney(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Money float64 `json:"money"`
		Memo  string  `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	logEntry, err := services.SetUserMoney(userID, req.Money, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "余额设置成功", "log": logEntry})
}

// AddMoneyLog 仅添加余额变动日志（不修改余额）
// @Summary 仅添加余额变动日志
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/money/log [post]
func (ctrl *UserMoneyScoreController) AddMoneyLog(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Money *float64 `json:"money" binding:"required"`
		Memo  string   `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if req.Money == nil {
		utils.Fail(c, 400, "金额不能为空")
		return
	}

	logEntry, err := services.AddUserMoneyLogOnly(userID, *req.Money, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "余额日志添加成功", "log": logEntry})
}

// OperateMoney 统一余额操作（支持余额/日志/订单组合）
// @Summary 统一余额操作
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/money/operate [post]
func (ctrl *UserMoneyScoreController) OperateMoney(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Money       *float64 `json:"money"`
		Memo        string   `json:"memo"`
		Operation   string   `json:"operation" binding:"required"`
		OrderNo     string   `json:"order_no"`
		TradeNo     string   `json:"trade_no"`
		OrderStatus *int     `json:"order_status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	orderStatus := models.PaymentStatusPaid
	if req.OrderStatus != nil {
		orderStatus = *req.OrderStatus
	}

	amount := 0.0
	if req.Money != nil {
		amount = *req.Money
	}

	if req.Operation != "order_only" && req.Money == nil {
		utils.Fail(c, 400, "金额不能为空")
		return
	}

	result, err := services.OperateUserMoney(userID, services.MoneyOperationRequest{
		Amount:      amount,
		Memo:        req.Memo,
		Operation:   req.Operation,
		OrderNo:     req.OrderNo,
		TradeNo:     req.TradeNo,
		OrderStatus: orderStatus,
	})
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "余额组合操作成功", "result": result})
}

// MoneyLogList 获取余额变动日志列表（管理员可查看所有）
// @Summary 余额变动日志列表
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/money-logs [get]
func (ctrl *UserMoneyScoreController) MoneyLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	keyword := c.DefaultQuery("keyword", "")
	userIDFilter, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	logs, total, err := services.GetUserMoneyLogList(userIDFilter, page, pageSize, keyword)
	if err != nil {
		log.Printf("[ADMIN][MONEY] list money logs failed: %v", err)
		utils.Fail(c, 500, "获取余额日志失败")
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// MoneyLogDetail 获取单条余额变动记录
// @Summary 余额变动日志详情
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/money-logs/{id} [get]
func (ctrl *UserMoneyScoreController) MoneyLogDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	logEntry, err := models.GetUserMoneyLogByID(id)
	if err != nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}

	utils.Success(c, logEntry)
}

// MoneyLogDelete 删除余额变动记录（不影响用户余额）
// @Summary 删除余额变动日志
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/money-logs/{id} [delete]
func (ctrl *UserMoneyScoreController) MoneyLogDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	if err := models.DeleteUserMoneyLog(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Fail(c, 404, "记录不存在")
			return
		}
		log.Printf("[ADMIN][MONEY] delete money log failed id=%d: %v", id, err)
		utils.Fail(c, 500, "删除余额日志失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// ========================================
// 积分管理
// ========================================

// ChangeScore 变更用户积分（增减）
// @Summary 变更用户积分
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/score/change [post]
func (ctrl *UserMoneyScoreController) ChangeScore(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Score int64  `json:"score" binding:"required"`
		Memo  string `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	logEntry, err := services.ChangeUserScore(userID, req.Score, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "积分变更成功", "log": logEntry})
}

// SetScore 直接设置用户积分
// @Summary 直接设置用户积分
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/score [put]
func (ctrl *UserMoneyScoreController) SetScore(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Score int64  `json:"score"`
		Memo  string `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	logEntry, err := services.SetUserScore(userID, req.Score, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "积分设置成功", "log": logEntry})
}

// AddScoreLog 仅添加积分变动日志（不修改积分）
// @Summary 仅添加积分变动日志
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/users/{id}/score/log [post]
func (ctrl *UserMoneyScoreController) AddScoreLog(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "用户ID格式错误")
		return
	}

	var req struct {
		Score int64  `json:"score" binding:"required"`
		Memo  string `json:"memo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	logEntry, err := services.AddUserScoreLogOnly(userID, req.Score, req.Memo)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "积分日志添加成功", "log": logEntry})
}

// ScoreLogList 获取积分变动日志列表（管理员可查看所有）
// @Summary 积分变动日志列表
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/score-logs [get]
func (ctrl *UserMoneyScoreController) ScoreLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	keyword := c.DefaultQuery("keyword", "")
	userIDFilter, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	logs, total, err := services.GetUserScoreLogList(userIDFilter, page, pageSize, keyword)
	if err != nil {
		log.Printf("[ADMIN][SCORE] list score logs failed: %v", err)
		utils.Fail(c, 500, "获取积分日志失败")
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// ScoreLogDetail 获取单条积分变动记录
// @Summary 积分变动日志详情
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/score-logs/{id} [get]
func (ctrl *UserMoneyScoreController) ScoreLogDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	logEntry, err := models.GetUserScoreLogByID(id)
	if err != nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}

	utils.Success(c, logEntry)
}

// ScoreLogDelete 删除积分变动记录（不影响用户积分）
// @Summary 删除积分变动日志
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/score-logs/{id} [delete]
func (ctrl *UserMoneyScoreController) ScoreLogDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	if err := models.DeleteUserScoreLog(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Fail(c, 404, "记录不存在")
			return
		}
		log.Printf("[ADMIN][SCORE] delete score log failed id=%d: %v", id, err)
		utils.Fail(c, 500, "删除积分日志失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GenerateNos 生成订单号和交易号
// @Summary 生成订单号和交易号
// @Tags Admin-用户积分
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/generate-nos [get]
func (ctrl *UserMoneyScoreController) GenerateNos(c *gin.Context) {
	// 订单号：复用 models 中的生成逻辑
	orderNo := models.GenerateOrderNo()

	// 交易号：T + 年月日时分秒 + 6位密码学随机数
	now := time.Now()
	rnd, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	tradeNo := fmt.Sprintf("T%s%06d", now.Format("20060102150405"), rnd.Int64())

	utils.Success(c, gin.H{
		"order_no": orderNo,
		"trade_no": tradeNo,
	})
}

// RegisterRoutes 注册管理员余额/积分路由
func (ctrl *UserMoneyScoreController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	// 用户余额/积分操作（挂在 users/:id 下）
	// 资金/积分变更均要求幂等键，防止网络重试或双击导致重复加减款
	users := adminGroup.Group("/users")
	users.Use(middleware.SimpleLogMiddleware("资金调账"))
	{
		users.POST("/:id/money/change", middleware.RequireIdempotency("admin_money_change", 10*time.Minute), ctrl.ChangeMoney)
		users.PUT("/:id/money", middleware.RequireIdempotency("admin_money_set", 10*time.Minute), ctrl.SetMoney)
		users.POST("/:id/money/log", middleware.RequireIdempotency("admin_money_log", 10*time.Minute), ctrl.AddMoneyLog)
		users.POST("/:id/money/operate", middleware.RequireIdempotency("admin_money_operate", 10*time.Minute), ctrl.OperateMoney)
		users.POST("/:id/score/change", middleware.RequireIdempotency("admin_score_change", 10*time.Minute), ctrl.ChangeScore)
		users.PUT("/:id/score", middleware.RequireIdempotency("admin_score_set", 10*time.Minute), ctrl.SetScore)
		users.POST("/:id/score/log", middleware.RequireIdempotency("admin_score_log", 10*time.Minute), ctrl.AddScoreLog)
	}

	// 生成订单号/交易号
	adminGroup.GET("/generate-nos", ctrl.GenerateNos)

	// 余额日志（删除入口保留为软删，仍挂审计）
	moneyLogs := adminGroup.Group("/money-logs")
	moneyLogs.Use(middleware.SimpleLogMiddleware("余额日志"))
	{
		moneyLogs.GET("", ctrl.MoneyLogList)
		moneyLogs.GET("/:id", ctrl.MoneyLogDetail)
		moneyLogs.DELETE("/:id", ctrl.MoneyLogDelete)
	}

	// 积分日志
	scoreLogs := adminGroup.Group("/score-logs")
	scoreLogs.Use(middleware.SimpleLogMiddleware("积分日志"))
	{
		scoreLogs.GET("", ctrl.ScoreLogList)
		scoreLogs.GET("/:id", ctrl.ScoreLogDetail)
		scoreLogs.DELETE("/:id", ctrl.ScoreLogDelete)
	}
}

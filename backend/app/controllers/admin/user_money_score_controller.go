package admin

import (
	"crypto/rand"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
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
// POST /api/v1/admin/users/:id/money/change
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
// PUT /api/v1/admin/users/:id/money
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
// POST /api/v1/admin/users/:id/money/log
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
// POST /api/v1/admin/users/:id/money/operate
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
// GET /api/v1/admin/money-logs
func (ctrl *UserMoneyScoreController) MoneyLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.DefaultQuery("keyword", "")
	userIDFilter, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := services.GetUserMoneyLogList(userIDFilter, page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, "获取余额日志失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// MoneyLogDetail 获取单条余额变动记录
// GET /api/v1/admin/money-logs/:id
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
// DELETE /api/v1/admin/money-logs/:id
func (ctrl *UserMoneyScoreController) MoneyLogDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	if err := models.DeleteUserMoneyLog(id); err != nil {
		utils.Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// ========================================
// 积分管理
// ========================================

// ChangeScore 变更用户积分（增减）
// POST /api/v1/admin/users/:id/score/change
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
// PUT /api/v1/admin/users/:id/score
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
// POST /api/v1/admin/users/:id/score/log
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
// GET /api/v1/admin/score-logs
func (ctrl *UserMoneyScoreController) ScoreLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.DefaultQuery("keyword", "")
	userIDFilter, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := services.GetUserScoreLogList(userIDFilter, page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, "获取积分日志失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"list": logs, "total": total})
}

// ScoreLogDetail 获取单条积分变动记录
// GET /api/v1/admin/score-logs/:id
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
// DELETE /api/v1/admin/score-logs/:id
func (ctrl *UserMoneyScoreController) ScoreLogDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "记录ID格式错误")
		return
	}

	if err := models.DeleteUserScoreLog(id); err != nil {
		utils.Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GenerateNos 生成订单号和交易号
// GET /api/v1/admin/generate-nos
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
	users := adminGroup.Group("/users")
	{
		users.POST("/:id/money/change", ctrl.ChangeMoney)
		users.PUT("/:id/money", ctrl.SetMoney)
		users.POST("/:id/money/log", ctrl.AddMoneyLog)
		users.POST("/:id/money/operate", ctrl.OperateMoney)
		users.POST("/:id/score/change", ctrl.ChangeScore)
		users.PUT("/:id/score", ctrl.SetScore)
		users.POST("/:id/score/log", ctrl.AddScoreLog)
	}

	// 生成订单号/交易号
	adminGroup.GET("/generate-nos", ctrl.GenerateNos)

	// 余额日志
	moneyLogs := adminGroup.Group("/money-logs")
	{
		moneyLogs.GET("", ctrl.MoneyLogList)
		moneyLogs.GET("/:id", ctrl.MoneyLogDetail)
		moneyLogs.DELETE("/:id", ctrl.MoneyLogDelete)
	}

	// 积分日志
	scoreLogs := adminGroup.Group("/score-logs")
	{
		scoreLogs.GET("", ctrl.ScoreLogList)
		scoreLogs.GET("/:id", ctrl.ScoreLogDetail)
		scoreLogs.DELETE("/:id", ctrl.ScoreLogDelete)
	}
}

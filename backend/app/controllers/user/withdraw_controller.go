package user

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WithdrawController struct {
	withdrawService *services.WithdrawService
}

func NewWithdrawController() *WithdrawController {
	return &WithdrawController{
		withdrawService: services.NewWithdrawService(),
	}
}

type CreateWithdrawBody struct {
	Amount      float64 `json:"amount" binding:"required"`
	AccountType string  `json:"account_type"`
	AccountName string  `json:"account_name" binding:"required"`
	AccountNo   string  `json:"account_no" binding:"required"`
	RealName    string  `json:"real_name" binding:"required"`
	Remark      string  `json:"remark"`
}

func (ctrl *WithdrawController) Create(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}

	var req CreateWithdrawBody
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	req.AccountType = utils.Clean_XSS(req.AccountType)
	req.AccountName = utils.Clean_XSS(req.AccountName)
	req.AccountNo = utils.Clean_XSS(req.AccountNo)
	req.RealName = utils.Clean_XSS(req.RealName)
	req.Remark = utils.Clean_XSS(req.Remark)

	item, err := ctrl.withdrawService.Create(userID.(uint64), &services.CreateWithdrawRequest{
		Amount:      req.Amount,
		AccountType: req.AccountType,
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		RealName:    req.RealName,
		Remark:      req.Remark,
	})
	if err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[WITHDRAW] create request failed for user_id=%d: %v", userID.(uint64), err)
		utils.Fail(c, 500, "提现申请提交失败，请稍后重试")
		return
	}
	utils.SuccessMsg(c, "提现申请已提交，等待管理员审核", item)
}

func (ctrl *WithdrawController) List(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	statusStr := c.Query("status")
	var status *uint8
	if statusStr != "" {
		if v, err := strconv.ParseUint(statusStr, 10, 8); err == nil {
			val := uint8(v)
			status = &val
		}
	}

	result, err := ctrl.withdrawService.GetList(&models.WithdrawListQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID.(uint64),
		Status:   status,
	})
	if err != nil {
		utils.Fail(c, 500, "获取提现记录失败")
		return
	}
	utils.Success(c, result)
}

func (ctrl *WithdrawController) Detail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的ID")
		return
	}

	item, err := ctrl.withdrawService.GetByID(id)
	if err != nil {
		utils.Fail(c, 404, "提现记录不存在")
		return
	}
	if item.UserID != userID.(uint64) {
		utils.Fail(c, 403, "无权查看该提现记录")
		return
	}
	utils.Success(c, item)
}

func (ctrl *WithdrawController) RegisterRoutes(group *gin.RouterGroup) {
	withdraw := group.Group("/withdraw")
	{
		withdraw.POST("", middleware.RequireIdempotency("user_withdraw_create", 10*time.Minute), ctrl.Create)
		withdraw.GET("", ctrl.List)
		withdraw.GET("/:id", ctrl.Detail)
	}
}

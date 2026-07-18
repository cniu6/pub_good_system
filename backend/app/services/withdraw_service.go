package services

import (
	"database/sql"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"strings"
	"time"
	"unicode/utf8"
)

type WithdrawService struct{}

func NewWithdrawService() *WithdrawService {
	return &WithdrawService{}
}

type CreateWithdrawRequest struct {
	Amount      float64 `json:"amount"`
	AccountType string  `json:"account_type"`
	AccountName string  `json:"account_name"`
	AccountNo   string  `json:"account_no"`
	RealName    string  `json:"real_name"`
	Remark      string  `json:"remark"`
}

type ReviewWithdrawRequest struct {
	ID           uint64 `json:"id"`
	Status       uint8  `json:"status"`
	ReviewRemark string `json:"review_remark"`
}

type PayWithdrawRequest struct {
	ID             uint64 `json:"id"`
	TransferRemark string `json:"transfer_remark"`
}

func validateWithdrawTextField(value string, fieldName string, maxLen int) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxLen {
		return "", NewClientError(fmt.Sprintf("%s不能超过%d个字符", fieldName, maxLen))
	}
	return value, nil
}

func shouldRefundReservedBalance(reviewStatus uint8, balanceDeducted bool) bool {
	return reviewStatus == models.WithdrawStatusRejected && balanceDeducted
}

func shouldDeductBalanceOnWithdrawPay(balanceDeducted bool) bool {
	return !balanceDeducted
}

func (s *WithdrawService) Create(userID uint64, req *CreateWithdrawRequest) (*models.WithdrawRequest, error) {
	if userID == 0 {
		return nil, NewClientError("用户不存在")
	}
	if GlobalSettingsService != nil && !GlobalSettingsService.GetBoolWithDefault("withdraw_enabled", true) {
		return nil, NewClientError("提现功能暂未开启")
	}
	if req.Amount <= 0 {
		return nil, NewClientError("提现金额必须大于0")
	}
	// 提现金额先按分规范化
	if normalized, err := utils.NormalizeYuan(req.Amount); err != nil || normalized <= 0 {
		return nil, NewClientError("提现金额非法")
	} else {
		req.Amount = normalized
	}
	if GlobalSettingsService != nil {
		minAmount := parseJSONFloatWithDefault(GlobalSettingsService.GetWithDefault("withdraw_min_amount", "10"), 10)
		minFen := utils.MustYuanToFen(minAmount)
		reqFen := utils.MustYuanToFen(req.Amount)
		if reqFen < minFen {
			return nil, NewClientError(fmt.Sprintf("提现金额不能低于 %.2f", utils.FenToYuan(minFen)))
		}
	}

	accountType := strings.TrimSpace(req.AccountType)
	if accountType == "" {
		accountType = "bank"
	}
	var err error
	accountType, err = validateWithdrawTextField(accountType, "收款方式", 32)
	if err != nil {
		return nil, err
	}
	accountName, err := validateWithdrawTextField(req.AccountName, "账户名称", 100)
	if err != nil {
		return nil, err
	}
	accountNo, err := validateWithdrawTextField(req.AccountNo, "收款账号", 128)
	if err != nil {
		return nil, err
	}
	realName, err := validateWithdrawTextField(req.RealName, "收款人", 100)
	if err != nil {
		return nil, err
	}
	remark, err := validateWithdrawTextField(req.Remark, "备注", 255)
	if err != nil {
		return nil, err
	}
	if accountName == "" || accountNo == "" || realName == "" {
		return nil, NewClientError("请完整填写收款信息")
	}
	if GlobalSettingsService != nil {
		allowedTypes := parseJSONStringArrayWithDefault(GlobalSettingsService.GetWithDefault("withdraw_account_types", "[\"bank\",\"alipay\",\"wechat\",\"usdt\"]"), []string{"bank", "alipay", "wechat", "usdt"})
		typeAllowed := false
		for _, item := range allowedTypes {
			if strings.EqualFold(strings.TrimSpace(item), accountType) {
				typeAllowed = true
				break
			}
		}
		if !typeAllowed {
			return nil, NewClientError("当前收款方式不可用")
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user struct {
		Status uint8   `db:"status"`
		Money  float64 `db:"money"`
	}
	if err := tx.QueryRow("SELECT status, money FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE", userID).Scan(&user.Status, &user.Money); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewClientError("用户不存在")
		}
		return nil, err
	}
	if user.Status != 1 {
		return nil, NewClientError("当前用户状态不可提现")
	}
	// 余额与提现金额均按「分」比较，避免 float 误判不足/充足
	userFen, fenErr := utils.YuanToFen(user.Money)
	if fenErr != nil {
		return nil, NewClientError("账户余额异常")
	}
	reqFen, fenErr := utils.YuanToFen(req.Amount)
	if fenErr != nil || reqFen <= 0 {
		return nil, NewClientError("提现金额必须大于0")
	}
	// 写回规范化后的元，后续扣款/落库一致
	req.Amount = utils.FenToYuan(reqFen)
	if userFen < reqFen {
		return nil, NewClientError("账户余额不足")
	}

	var pendingCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM withdraw_requests WHERE user_id = ? AND status IN (?, ?) AND delete_time IS NULL FOR UPDATE", userID, models.WithdrawStatusPending, models.WithdrawStatusApproved).Scan(&pendingCount); err != nil {
		return nil, err
	}
	if pendingCount > 0 {
		return nil, NewClientError("您有待处理的提现申请，请勿重复提交")
	}

	if _, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: userID,
		Amount: -req.Amount,
		MemoI18n: map[string]string{
			"zhCN": "提交提现申请，系统预扣余额",
			"enUS": "Balance reserved for withdrawal request",
		},
	}, utils.OpChangeAndLog); err != nil {
		if strings.Contains(err.Error(), "扣款金额超出用户余额") {
			return nil, NewClientError("账户余额不足")
		}
		return nil, err
	}

	item := &models.WithdrawRequest{
		UserID:      userID,
		Amount:      req.Amount,
		AccountType: accountType,
		AccountName: accountName,
		AccountNo:   accountNo,
		RealName:    realName,
		Remark:      remark,
		Status:      models.WithdrawStatusPending,
		BalanceDeducted: true,
	}
	now := time.Now().Unix()
	item.CreateTime = now
	item.UpdateTime = now

	result, err := tx.Exec(
		`INSERT INTO withdraw_requests (user_id, amount, account_type, account_name, account_no, real_name, remark, status, balance_deducted, review_remark, transfer_remark, reviewed_at, reviewed_by, paid_at, paid_by, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.UserID, item.Amount, item.AccountType, item.AccountName, item.AccountNo, item.RealName, item.Remark,
		item.Status, item.BalanceDeducted, item.ReviewRemark, item.TransferRemark, item.ReviewedAt, item.ReviewedBy, item.PaidAt, item.PaidBy,
		item.CreateTime, item.UpdateTime, item.DeleteTime,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	item.ID = uint64(id)

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return models.GetWithdrawRequestByID(item.ID)
}

func (s *WithdrawService) GetList(query *models.WithdrawListQuery) (*models.WithdrawListResult, error) {
	return models.GetWithdrawRequestList(query)
}

func (s *WithdrawService) GetStats(query *models.WithdrawListQuery) (*models.WithdrawStatsResult, error) {
	return models.GetWithdrawRequestStats(query)
}

func (s *WithdrawService) GetByID(id uint64) (*models.WithdrawRequest, error) {
	return models.GetWithdrawRequestByID(id)
}

func (s *WithdrawService) Review(adminID uint64, req *ReviewWithdrawRequest) error {
	if req.Status != models.WithdrawStatusApproved && req.Status != models.WithdrawStatusRejected {
		return NewClientError("审核状态无效")
	}
	reviewRemark, err := validateWithdrawTextField(req.ReviewRemark, "审核备注", 255)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	item, err := models.GetWithdrawRequestByIDForUpdate(tx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return NewClientError("提现申请不存在")
		}
		return err
	}
	if item.Status != models.WithdrawStatusPending {
		return NewClientError("该提现申请已处理，无法重复审核")
	}
	if item.UserID == adminID {
		return NewClientError("不能审核自己的提现申请")
	}

	if shouldRefundReservedBalance(req.Status, item.BalanceDeducted) {
		if _, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
			UserID: item.UserID,
			Amount: item.Amount,
			MemoI18n: map[string]string{
				"zhCN": fmt.Sprintf("提现申请已拒绝，退回预扣余额-申请#%d", item.ID),
				"enUS": fmt.Sprintf("Withdrawal rejected, reserved balance released - Request#%d", item.ID),
			},
		}, utils.OpChangeAndLog); err != nil {
			return err
		}
	}

	if err := models.UpdateWithdrawReviewTx(tx, req.ID, req.Status, reviewRemark, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *WithdrawService) MarkPaid(adminID uint64, req *PayWithdrawRequest) error {
	transferRemark, err := validateWithdrawTextField(req.TransferRemark, "打款备注", 255)
	if err != nil {
		return err
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	item, err := models.GetWithdrawRequestByIDForUpdate(tx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return NewClientError("提现申请不存在")
		}
		return err
	}
	if item.Status != models.WithdrawStatusApproved {
		return NewClientError("仅已审核通过的提现申请可标记为已打款")
	}
	if item.UserID == adminID {
		return NewClientError("不能给自己的提现申请执行打款")
	}

	if shouldDeductBalanceOnWithdrawPay(item.BalanceDeducted) {
		if _, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
			UserID: item.UserID,
			Amount: -item.Amount,
			MemoI18n: map[string]string{
				"zhCN": fmt.Sprintf("人工提现打款-申请#%d", item.ID),
				"enUS": fmt.Sprintf("Manual withdrawal payout - Request#%d", item.ID),
			},
		}, utils.OpChangeAndLog); err != nil {
			if strings.Contains(err.Error(), "扣款金额超出用户余额") {
				return NewClientError("用户当前余额不足，无法执行提现打款")
			}
			return err
		}
	}

	if err := models.MarkWithdrawPaidTx(tx, item.ID, transferRemark, adminID); err != nil {
		return err
	}
	return tx.Commit()
}


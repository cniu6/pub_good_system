package services

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/db"
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
		return "", fmt.Errorf("%s不能超过%d个字符", fieldName, maxLen)
	}
	return value, nil
}

func (s *WithdrawService) Create(userID uint64, req *CreateWithdrawRequest) (*models.WithdrawRequest, error) {
	if userID == 0 {
		return nil, errors.New("用户不存在")
	}
	if GlobalSettingsService != nil && !GlobalSettingsService.GetBoolWithDefault("withdraw_enabled", true) {
		return nil, errors.New("提现功能暂未开启")
	}
	if req.Amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}
	if GlobalSettingsService != nil {
		minAmount := parseJSONFloatWithDefault(GlobalSettingsService.GetWithDefault("withdraw_min_amount", "10"), 10)
		if req.Amount < minAmount {
			return nil, fmt.Errorf("提现金额不能低于 %.2f", minAmount)
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
		return nil, errors.New("请完整填写收款信息")
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
			return nil, errors.New("当前收款方式不可用")
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
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	if user.Status != 1 {
		return nil, errors.New("当前用户状态不可提现")
	}
	if user.Money < req.Amount {
		return nil, errors.New("账户余额不足")
	}

	var pendingCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM withdraw_requests WHERE user_id = ? AND status IN (?, ?) AND delete_time IS NULL FOR UPDATE", userID, models.WithdrawStatusPending, models.WithdrawStatusApproved).Scan(&pendingCount); err != nil {
		return nil, err
	}
	if pendingCount > 0 {
		return nil, errors.New("您有待处理的提现申请，请勿重复提交")
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
	}
	now := time.Now().Unix()
	item.CreateTime = now
	item.UpdateTime = now

	result, err := tx.Exec(
		`INSERT INTO withdraw_requests (user_id, amount, account_type, account_name, account_no, real_name, remark, status, review_remark, transfer_remark, reviewed_at, reviewed_by, paid_at, paid_by, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.UserID, item.Amount, item.AccountType, item.AccountName, item.AccountNo, item.RealName, item.Remark,
		item.Status, item.ReviewRemark, item.TransferRemark, item.ReviewedAt, item.ReviewedBy, item.PaidAt, item.PaidBy,
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
		return errors.New("审核状态无效")
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
			return errors.New("提现申请不存在")
		}
		return err
	}
	if item.Status != models.WithdrawStatusPending {
		return errors.New("该提现申请已处理，无法重复审核")
	}
	if item.UserID == adminID {
		return errors.New("不能审核自己的提现申请")
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
			return errors.New("提现申请不存在")
		}
		return err
	}
	if item.Status != models.WithdrawStatusApproved {
		return errors.New("仅已审核通过的提现申请可标记为已打款")
	}
	if item.UserID == adminID {
		return errors.New("不能给自己的提现申请执行打款")
	}

	beforeMoney, err := models.GetUserMoneyForUpdate(tx, item.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if beforeMoney < item.Amount {
		return errors.New("用户当前余额不足，无法执行提现打款")
	}

	memo := utils.BuildMemo("", map[string]string{
		"zhCN": fmt.Sprintf("人工提现打款-申请#%d", item.ID),
		"enUS": fmt.Sprintf("Manual withdrawal payout - Request#%d", item.ID),
	})
	if _, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: item.UserID,
		Amount: -item.Amount,
		Memo:   memo,
	}, utils.OpChangeAndLog); err != nil {
		return err
	}

	if err := models.MarkWithdrawPaidTx(tx, item.ID, transferRemark, adminID); err != nil {
		return err
	}
	return tx.Commit()
}

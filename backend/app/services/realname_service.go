package services

import (
	"database/sql"
	"errors"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// RealnameService 实名认证服务
type RealnameService struct{}

func NewRealnameService() *RealnameService {
	return &RealnameService{}
}

// RealnameSubmitRequest 提交实名认证请求
type RealnameSubmitRequest struct {
	RealName         string `json:"real_name" binding:"required"`
	CertificateType  uint8  `json:"certificate_type" binding:"required,min=1,max=3"`
	CertificateNo    string `json:"certificate_no" binding:"required"`
	CertificateFront string `json:"certificate_front" binding:"required"`
	CertificateBack  string `json:"certificate_back" binding:"required"`
}

// RealnameReviewRequest 审核实名认证请求
type RealnameReviewRequest struct {
	ID           uint64 `json:"id" binding:"required"`
	Status       uint8  `json:"status" binding:"required,min=1,max=2"`
	RejectReason string `json:"reject_reason"`
}

// RealnameStatus 实名认证状态常量
type RealnameStatus struct{}

// Status constants
const (
	RealnameStatusPending  uint8 = 0 // 待审核
	RealnameStatusApproved uint8 = 1 // 已通过
	RealnameStatusRejected uint8 = 2 // 已拒绝
)

// CertificateType 证件类型常量
type CertificateType struct{}

// Certificate type constants
const (
	CertificateTypeIDCard   uint8 = 1 // 身份证
	CertificateTypePassport uint8 = 2 // 护照
	CertificateTypeOfficer  uint8 = 3 // 军官证
)

// Submit 提交实名认证
func (s *RealnameService) Submit(userID uint64, req *RealnameSubmitRequest) error {
	if GlobalSettingsService != nil && !GlobalSettingsService.GetBoolWithDefault("realname_enabled", true) {
		return NewClientError("实名认证功能暂未开启")
	}

	// 参数校验
	realName := strings.TrimSpace(req.RealName)
	if len(realName) < 2 || len(realName) > 50 {
		return NewClientError("姓名长度必须在2-50个字符之间")
	}

	// 姓名只能包含中文、英文字母和点号
	for _, r := range realName {
		if !unicode.IsLetter(r) && r != '.' && !unicode.IsSpace(r) {
			return NewClientError("姓名只能包含中文、英文字母和点号")
		}
	}

	if err := s.validateCertificateType(req.CertificateType); err != nil {
		return err
	}

	certNo := strings.TrimSpace(req.CertificateNo)
	if err := s.validateCertificateNo(req.CertificateType, certNo); err != nil {
		return err
	}

	if err := validateCertificateImageURL(req.CertificateFront, "证件正面照"); err != nil {
		return err
	}
	if err := validateCertificateImageURL(req.CertificateBack, "证件背面照"); err != nil {
		return err
	}

	verification := &models.RealnameVerification{
		UserID:           userID,
		RealName:         realName,
		CertificateType:  req.CertificateType,
		CertificateNo:    strings.ToUpper(certNo),
		CertificateFront: req.CertificateFront,
		CertificateBack:  req.CertificateBack,
	}

	// 当后台关闭“需要审核”时，提交后直接标记为已通过，避免仍然进入待审核状态。
	if GlobalSettingsService != nil && !GlobalSettingsService.GetBoolWithDefault("realname_review_required", true) {
		now := time.Now().Unix()
		reviewedBy := uint64(0)
		verification.Status = RealnameStatusApproved
		verification.RejectReason = ""
		verification.ReviewedAt = &now
		verification.ReviewedBy = &reviewedBy
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existing, err := models.GetRealnameVerificationByUserIDForUpdate(tx, userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 在同一事务里锁定并判断最新申请，避免并发下重复提交。
	if existing != nil && existing.Status == RealnameStatusPending {
		return NewClientError("您有待审核的实名认证申请，请等待审核完成")
	}
	if existing != nil && existing.Status == RealnameStatusApproved {
		return NewClientError("您已通过实名认证，无需再次提交")
	}
	if existing != nil && existing.Status == RealnameStatusRejected {
		if err := models.SoftDeleteRealnameVerificationTx(tx, existing.ID); err != nil {
			return errors.New("处理旧记录失败，请重试")
		}
	}

	// 证件号跨用户查重：避免同一证件号被多个账号占用实名认证
	dupCount, err := models.CountOtherUsersByCertificateNoTx(tx, verification.CertificateNo, userID)
	if err != nil {
		return errors.New("查重失败，请重试")
	}
	if dupCount > 0 {
		return NewClientError("该证件号已被其他账号实名认证，请核对后重试")
	}

	if err := models.CreateRealnameVerificationTx(tx, verification); err != nil {
		if db.IsDuplicateKeyError(err) {
			return NewClientError("该证件号已被其他账号实名认证，请核对后重试")
		}
		return err
	}

	return tx.Commit()
}

// GetUserVerification 获取用户的实名认证状态
func (s *RealnameService) GetUserVerification(userID uint64) (*models.RealnameVerification, error) {
	verification, err := models.GetRealnameVerificationByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return verification, nil
}

// GetByID 根据ID获取实名认证记录
func (s *RealnameService) GetByID(id uint64) (*models.RealnameVerification, error) {
	return models.GetRealnameVerificationByID(id)
}

// GetList 获取实名认证列表（管理员）
func (s *RealnameService) GetList(query *models.RealnameVerificationListQuery) (*models.RealnameVerificationListResult, error) {
	return models.GetRealnameVerificationList(query)
}

// Review 审核实名认证（管理员）
// 业务层错误统一使用 ClientError，便于控制器与内部错误区分。
func (s *RealnameService) Review(adminID uint64, req *RealnameReviewRequest) error {
	if req.Status != RealnameStatusApproved && req.Status != RealnameStatusRejected {
		return NewClientError("审核状态无效")
	}

	if req.Status == RealnameStatusRejected && strings.TrimSpace(req.RejectReason) == "" {
		return NewClientError("请填写拒绝原因")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	verification, err := models.GetRealnameVerificationByIDForUpdate(tx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return NewClientError("实名认证记录不存在")
		}
		return err
	}

	// 防止管理员审核自己的申请
	if adminID == verification.UserID {
		return NewClientError("不能审核自己的实名认证申请")
	}

	if verification.Status != RealnameStatusPending {
		return NewClientError("该申请已处理，无法重复审核")
	}

	rejectReason := ""
	if req.Status == RealnameStatusRejected {
		rejectReason = strings.TrimSpace(req.RejectReason)
	}

	if err := models.UpdateRealnameVerificationStatusTx(tx, req.ID, req.Status, rejectReason, adminID); err != nil {
		return err
	}

	return tx.Commit()
}

// validateCertificateImageURL 校验证件照 URL 协议，只允许 http(s)，
// 防止 javascript: / data: 等恶意 scheme 被存入数据库并在管理端渲染时执行。
func validateCertificateImageURL(rawURL, fieldName string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return NewClientError(fieldName + "不能为空")
	}
	if len(rawURL) > 1024 {
		return NewClientError(fieldName + "地址过长")
	}
	lower := strings.ToLower(rawURL)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return NewClientError(fieldName + "地址协议不合法，仅支持 http/https")
	}
	return nil
}

// validateCertificateType 验证证件类型
func (s *RealnameService) validateCertificateType(certType uint8) error {
	switch certType {
	case CertificateTypeIDCard, CertificateTypePassport, CertificateTypeOfficer:
		return nil
	default:
		return NewClientError("证件类型无效，支持: 1=身份证, 2=护照, 3=军官证")
	}
}

// validateCertificateNo 验证证件号码格式
func (s *RealnameService) validateCertificateNo(certType uint8, certNo string) error {
	certNo = strings.TrimSpace(certNo)
	if certNo == "" {
		return NewClientError("证件号码不能为空")
	}

	switch certType {
	case CertificateTypeIDCard:
		return s.validateIDCard(certNo)
	case CertificateTypePassport:
		return s.validatePassport(certNo)
	case CertificateTypeOfficer:
		return s.validateOfficerCert(certNo)
	default:
		return NewClientError("不支持的证件类型")
	}
}

// validateIDCard 验证身份证号码
func (s *RealnameService) validateIDCard(certNo string) error {
	length := len(certNo)
	if length != 15 && length != 18 {
		return NewClientError("身份证号码长度应为15位或18位")
	}

	// 15位身份证：全是数字
	if length == 15 {
		pattern := `^\d{15}$`
		match, _ := regexp.MatchString(pattern, certNo)
		if !match {
			return NewClientError("15位身份证号码格式不正确")
		}
		return nil
	}

	// 18位身份证：前17位为数字，最后一位可以是数字或X/x
	pattern := `^(\d{17}[\dXx])$`
	match, _ := regexp.MatchString(pattern, certNo)
	if !match {
		return NewClientError("18位身份证号码格式不正确")
	}

	// 校验码验证
	if !s.verifyIDCardChecksum(certNo) {
		return NewClientError("身份证号码校验码不正确")
	}

	return nil
}

// verifyIDCardChecksum 验证身份证校验码
func (s *RealnameService) verifyIDCardChecksum(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}

	// 加权因子
	weight := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// 校验码
	checkCode := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		digit := int(idCard[i] - '0')
		sum += digit * weight[i]
	}

	mod := sum % 11
	expectedCheckCode := checkCode[mod]

	lastChar := idCard[17]
	if lastChar == 'x' {
		lastChar = 'X'
	}

	return expectedCheckCode == lastChar
}

// validatePassport 验证护照号码
func (s *RealnameService) validatePassport(certNo string) error {
	length := len(certNo)
	if length < 6 || length > 20 {
		return NewClientError("护照号码长度应在6-20位之间")
	}
	// 护照号码：字母或数字
	pattern := `^[A-Za-z0-9]+$`
	match, _ := regexp.MatchString(pattern, certNo)
	if !match {
		return NewClientError("护照号码格式不正确，只能包含字母和数字")
	}
	return nil
}

// validateOfficerCert 验证军官证号码
func (s *RealnameService) validateOfficerCert(certNo string) error {
	length := len(certNo)
	if length < 5 || length > 20 {
		return NewClientError("军官证号码长度应在5-20位之间")
	}
	return nil
}

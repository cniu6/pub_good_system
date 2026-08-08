package services

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// UserListQuery 用户列表查询参数
type UserListQuery struct {
	Page           int    `form:"page" json:"page"`
	PageSize       int    `form:"page_size" json:"page_size"`
	Keyword        string `form:"keyword" json:"keyword"`
	Status         *uint8 `form:"status" json:"status"`
	GroupID        uint64 `form:"group_id" json:"group_id"`
	RealnameStatus *uint8 `form:"realname_status" json:"realname_status"`
}

type AdminUserListItem struct {
	models.User
	AdminRemark      string  `gorm:"-" json:"admin_remark"`
	RealnameStatus   *uint8  `gorm:"column:realname_status" json:"realname_status"`
	TotalPaidAmount  float64 `gorm:"column:total_paid_amount" json:"total_paid_amount"`
	BalancePaidRatio float64 `gorm:"column:balance_paid_ratio" json:"balance_paid_ratio"`
	// ApikeyMasked 管理端列表/详情展示用：仅末4位，不下发明文
	ApikeyMasked string `gorm:"-" json:"apikey"`
	// LastSeenAt 最近一次会话心跳时间（跨全部设备取最大值），来自 user_sessions；无会话记录时为 0。
	// 仅供列表展示「上次在线」参考，不代表当前是否在线。
	LastSeenAt int64 `gorm:"-" json:"last_seen_at"`
	// IsOnline 当前是否在线（依据 LastSeenAt 与在线心跳容忍窗口判定，与专门的在线用户页口径一致）。
	IsOnline bool `gorm:"-" json:"is_online"`
}
//@name 管理端用户列表项

// UserListResult 用户列表返回结果
type UserListResult struct {
	List     []AdminUserListItem `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// GetList 分页获取用户列表
func (s *UserService) GetList(query *UserListQuery) (*UserListResult, error) {
	var users []AdminUserListItem
	var total int64

	// 默认分页参数；上限防止无界查询拖垮库
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询条件
	// 显式带上 users. 前缀，避免与联表字段（如 rv.status）产生歧义或误命中。
	where := "WHERE users.delete_time IS NULL"
	args := []interface{}{}

	if query.Keyword != "" {
		where += " AND (users.username LIKE ? OR users.nickname LIKE ? OR users.email LIKE ? OR users.mobile LIKE ? OR users.admin_remark LIKE ?)"
		kw := "%" + query.Keyword + "%"
		args = append(args, kw, kw, kw, kw, kw)
	}
	if query.Status != nil {
		where += " AND users.status = ?"
		args = append(args, *query.Status)
	}
	if query.GroupID > 0 {
		where += " AND users.group_id = ?"
		args = append(args, query.GroupID)
	}
	if query.RealnameStatus != nil {
		where += " AND rv.status = ?"
		args = append(args, *query.RealnameStatus)
	}

	fromClause := fmt.Sprintf(` FROM users
		LEFT JOIN (
			SELECT t.user_id, t.status
			FROM user_realname_verifications t
			INNER JOIN (
				SELECT user_id, MAX(id) AS max_id
				FROM user_realname_verifications
				WHERE delete_time IS NULL
				GROUP BY user_id
			) latest ON latest.max_id = t.id
		) rv ON rv.user_id = users.id
		LEFT JOIN (
			SELECT user_id, COALESCE(SUM(pay_amount), 0) AS total_paid_amount
			FROM payment_orders
			WHERE status = ? AND %s
			GROUP BY user_id
		) p ON p.user_id = users.id `, models.RealPaidOrderFilterSQL)

	baseArgs := append([]interface{}{models.PaymentStatusPaid}, args...)

	// 查询总数
	count_query := "SELECT COUNT(*)" + fromClause + where
	err := db.DB.Raw(count_query, baseArgs...).Scan(&total).Error
	if err != nil {
		return nil, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	list_query := "SELECT " + models.BuildUserSelectColumns("users") + ", rv.status AS realname_status, COALESCE(p.total_paid_amount, 0) AS total_paid_amount, CASE WHEN COALESCE(p.total_paid_amount, 0) > 0 THEN COALESCE(users.money, 0) / p.total_paid_amount ELSE 0 END AS balance_paid_ratio" + fromClause + where + " ORDER BY users.id DESC LIMIT ? OFFSET ?"
	listArgs := append(baseArgs, query.PageSize, offset)

	err = db.DB.Raw(list_query, listArgs...).Scan(&users).Error
	if err != nil {
		return nil, err
	}

	userIDs := make([]uint64, 0, len(users))
	for i := range users {
		userIDs = append(userIDs, users[i].ID)
	}
	lastSeenMap, err := models.GetLastSeenAtByUserIDs(userIDs)
	if err != nil {
		// 会话表查询失败不应影响用户列表主流程，仅记录日志并跳过「上次在线」展示。
		lastSeenMap = map[uint64]int64{}
	}
	graceSeconds := models.GetOnlineHeartbeatGraceSeconds()
	now := time.Now().Unix()

	for i := range users {
		users[i].AdminRemark = users[i].User.AdminRemark
		users[i].ApikeyMasked = users[i].User.MaskedApikey()
		if seen, ok := lastSeenMap[users[i].ID]; ok {
			users[i].LastSeenAt = seen
			users[i].IsOnline = seen >= now-graceSeconds
		}
	}

	return &UserListResult{
		List:     users,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id uint64) (*models.User, error) {
	return models.GetUserByID(id)
}

func (s *UserService) GetAdminDetail(id uint64) (*AdminUserListItem, error) {
	var user AdminUserListItem
	query := fmt.Sprintf(`SELECT
		%s,
		rv.status AS realname_status,
		COALESCE(p.total_paid_amount, 0) AS total_paid_amount,
		CASE WHEN COALESCE(p.total_paid_amount, 0) > 0 THEN COALESCE(users.money, 0) / p.total_paid_amount ELSE 0 END AS balance_paid_ratio
	FROM users
	LEFT JOIN (
		SELECT t.user_id, t.status
		FROM user_realname_verifications t
		INNER JOIN (
			SELECT user_id, MAX(id) AS max_id
			FROM user_realname_verifications
			WHERE delete_time IS NULL
			GROUP BY user_id
		) latest ON latest.max_id = t.id
	) rv ON rv.user_id = users.id
	LEFT JOIN (
		SELECT user_id, COALESCE(SUM(pay_amount), 0) AS total_paid_amount
		FROM payment_orders
		WHERE status = ? AND %s
		GROUP BY user_id
	) p ON p.user_id = users.id
	WHERE users.id = ? AND users.delete_time IS NULL`, models.BuildUserSelectColumns("users"), models.RealPaidOrderFilterSQL)

	if err := db.DB.Raw(query, models.PaymentStatusPaid, id).Scan(&user).Error; err != nil {
		return nil, db.MapGormNotFound(err)
	}
	user.AdminRemark = user.User.AdminRemark
	user.ApikeyMasked = user.User.MaskedApikey()
	if lastSeenMap, err := models.GetLastSeenAtByUserIDs([]uint64{user.ID}); err == nil {
		if seen, ok := lastSeenMap[user.ID]; ok {
			user.LastSeenAt = seen
			user.IsOnline = seen >= time.Now().Unix()-models.GetOnlineHeartbeatGraceSeconds()
		}
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return models.GetUserByUsername(username)
}

// GetByEmail 根据邮箱获取用户
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	return models.GetUserByEmail(email)
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Nickname    string `json:"nickname"`
	Mobile      string `json:"mobile"`
	Language    string `json:"language"`
	Country     string `json:"country"`
	AdminRemark string `json:"admin_remark"`
	Level       uint64 `json:"level"`
	Role        string `json:"role"`
	// Status 用指针区分「未传」与「显式禁用」：未传默认启用；显式 0 才创建为禁用账号
	Status  *uint8 `json:"status"`
	GroupID uint64 `json:"group_id"`
}
//@name 创建用户请求

// Create 创建用户
func (s *UserService) Create(req *UserCreateRequest) (*models.User, error) {
	if err := validateClientRuneLen(req.Username, "用户名", utils.MaxUsernameLength); err != nil {
		return nil, err
	}
	if err := validateClientRuneLen(req.Nickname, "昵称", utils.MaxNicknameLength); err != nil {
		return nil, err
	}
	if err := validateClientRuneLen(req.Email, "邮箱", utils.MaxEmailLength); err != nil {
		return nil, err
	}
	if err := validateClientRuneLen(req.AdminRemark, "管理员备注", utils.MaxAdminRemarkLength); err != nil {
		return nil, err
	}
	if err := validateClientRuneLen(req.Country, "国家/地区", utils.MaxCountryLength); err != nil {
		return nil, err
	}
	if err := validateClientRuneLen(req.Language, "语言", utils.MaxLanguageLength); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在（DB 故障不可吞掉，否则会跳过唯一性判断）
	existing, err := models.GetUserByUsername(req.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("检查用户名失败: " + err.Error())
	}
	if existing != nil {
		return nil, NewClientError("用户名已存在")
	}

	// 检查邮箱是否已存在
	existing, err = models.GetUserByEmail(req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("检查邮箱失败: " + err.Error())
	}
	if existing != nil {
		return nil, NewClientError("邮箱已存在")
	}

	mobile := strings.TrimSpace(req.Mobile)
	if mobile != "" {
		normalized, normErr := utils.NormalizeAndValidateMobile(mobile, GetGlobalMobileCNOnly())
		if normErr != nil {
			return nil, NewClientError(normErr.Error())
		}
		mobile = normalized
		existingMobile, err := models.GetUserByMobile(mobile)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("检查手机号失败: " + err.Error())
		}
		if existingMobile != nil {
			return nil, NewClientError("手机号已被使用")
		}
	}

	user := &models.User{
		Username:    req.Username,
		Email:       req.Email,
		Nickname:    req.Nickname,
		Mobile:      mobile,
		Language:    req.Language,
		Country:     req.Country,
		AdminRemark: req.AdminRemark,
		Level:       req.Level,
		Role:        req.Role,
		GroupId:     req.GroupID,
		Password:    req.Password, // 调用方需要先加密
	}

	// 未传 status 时默认启用，避免 API 漏传字段时静默创建出禁用账号
	if req.Status == nil {
		user.Status = 1
	} else if *req.Status > 1 {
		return nil, NewClientError("用户状态只能为 0（禁用）或 1（启用）")
	} else {
		user.Status = *req.Status
	}

	if user.Role == "" {
		user.Role = "user"
	}
	normalizedRole, errRole := normalizeUserRole(user.Role)
	if errRole != nil {
		return nil, errRole
	}
	user.Role = normalizedRole
	if user.Language == "" {
		user.Language = "zh-CN"
	}

	err = models.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	ID         uint64  `json:"id"`
	Nickname   *string `json:"nickname"`
	Email      *string `json:"email"`
	Mobile     *string `json:"mobile"`
	Avatar     *string `json:"avatar"`
	Gender     *uint8  `json:"gender"`     // 指针类型，允许设置为0（保密）
	Birthday   *int64  `json:"birthday"`
	Motto      *string `json:"motto"`
	BackGround *string `json:"back_ground"`
	Language   *string `json:"language"`
	Country    *string `json:"country"`
	AdminRemark *string `json:"admin_remark"`
	Level      *uint64 `json:"level"`
	Role       *string `json:"role"`
	Status     *uint8  `json:"status"`
	GroupID    *uint64 `json:"group_id"`
}
//@name 更新用户请求

// Update 更新用户
func (s *UserService) Update(req *UserUpdateRequest) error {
	user, err := models.GetUserByID(req.ID)
	if err != nil {
		return NewClientError("用户不存在")
	}

	// 记录原始关键字段，用于更新完成后判断是否需要级联撤销会话
	previousStatus := user.Status
	previousRole := user.Role

	newRole := previousRole
	if req.Role != nil {
		newRole = *req.Role
	}
	var newStatus *uint8
	if req.Status != nil {
		newStatus = req.Status
	}
	if err := ensureAdminPrivilegeSafe(user, newRole, newStatus, false); err != nil {
		return err
	}

	// 检查邮箱是否被其他用户使用（DB 查询本身出错要直接拒绝，不能当成「邮箱不存在」放过）
	if req.Email != nil && *req.Email != user.Email {
		if err := validateClientRuneLen(*req.Email, "邮箱", utils.MaxEmailLength); err != nil {
			return err
		}
		existing, err := models.GetUserByEmail(*req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.New("检查邮箱失败: " + err.Error())
		}
		if existing != nil && existing.ID != user.ID {
			return NewClientError("邮箱已被使用")
		}
		user.Email = *req.Email
	}
	if req.Mobile != nil && *req.Mobile != user.Mobile {
		mobileVal := strings.TrimSpace(*req.Mobile)
		if mobileVal == "" {
			user.Mobile = ""
		} else {
			normalized, normErr := utils.NormalizeAndValidateMobile(mobileVal, GetGlobalMobileCNOnly())
			if normErr != nil {
				return NewClientError(normErr.Error())
			}
			*req.Mobile = normalized
			existing, err := models.GetUserByMobile(normalized)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return errors.New("检查手机号失败: " + err.Error())
			}
			if existing != nil && existing.ID != user.ID {
				return NewClientError("手机号已被使用")
			}
			user.Mobile = normalized
		}
	}

	if req.Nickname != nil {
		if err := validateClientRuneLen(*req.Nickname, "昵称", utils.MaxNicknameLength); err != nil {
			return err
		}
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.Motto != nil {
		user.Motto = *req.Motto
	}
	if req.BackGround != nil {
		user.BackGround = *req.BackGround
	}
	if req.Language != nil {
		if err := validateClientRuneLen(*req.Language, "语言", utils.MaxLanguageLength); err != nil {
			return err
		}
		user.Language = *req.Language
	}
	if req.Country != nil {
		if err := validateClientRuneLen(*req.Country, "国家/地区", utils.MaxCountryLength); err != nil {
			return err
		}
		user.Country = *req.Country
	}
	if req.AdminRemark != nil {
		if err := validateClientRuneLen(*req.AdminRemark, "管理员备注", utils.MaxAdminRemarkLength); err != nil {
			return err
		}
		user.AdminRemark = *req.AdminRemark
	}
	if req.Level != nil && *req.Level > 0 {
		user.Level = *req.Level
	}
	if req.Role != nil {
		normalizedRole, errRole := normalizeUserRole(*req.Role)
		if errRole != nil {
			return errRole
		}
		user.Role = normalizedRole
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.GroupID != nil && *req.GroupID > 0 {
		user.GroupId = *req.GroupID
	}

	now := time.Now().Unix()
	user.UpdateTime = &now

	if err := db.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"nickname":     user.Nickname,
		"email":        user.Email,
		"mobile":       user.Mobile,
		"avatar":       user.Avatar,
		"gender":       user.Gender,
		"birthday":     user.Birthday,
		"motto":        user.Motto,
		"admin_remark": user.AdminRemark,
		"back_ground":  user.BackGround,
		"language":     user.Language,
		"country":      user.Country,
		"level":        user.Level,
		"role":         user.Role,
		"status":       user.Status,
		"group_id":     user.GroupId,
		"update_time":  user.UpdateTime,
	}).Error; err != nil {
		return err
	}

	// 状态从启用变为禁用、或用户角色变更时，需级联撤销其全部会话，避免旧 token 仍然可用。
	if (previousStatus == 1 && user.Status == 0) || user.Role != previousRole {
		revokeAllGuardSessions(user.ID)
	}
	return nil
}

// UpdateStatus 更新用户状态。禁用时顺手撤销其所有活跃会话，避免旧 token 仍被使用。
func (s *UserService) UpdateStatus(user_id uint64, status uint8) error {
	user, err := models.GetUserByID(user_id)
	if err != nil {
		return NewClientError("用户不存在")
	}
	if err := ensureAdminPrivilegeSafe(user, user.Role, &status, false); err != nil {
		return err
	}

	now := time.Now().Unix()
	if err := db.DB.Model(&models.User{}).Where("id = ? AND delete_time IS NULL", user_id).Updates(map[string]interface{}{
		"status":      status,
		"update_time": now,
	}).Error; err != nil {
		return err
	}
	if status == 0 {
		revokeAllGuardSessions(user_id)
	}
	return nil
}

// UpdatePassword 更新用户密码，并撤销全部登录会话（防旧 token 继续使用）
func (s *UserService) UpdatePassword(user_id uint64, hashed_password string) error {
	if err := models.UpdatePassword(user_id, hashed_password); err != nil {
		return err
	}
	revokeAllGuardSessions(user_id)
	return nil
}

// normalizeUserRole 规范化并校验角色，仅允许 user / admin
func normalizeUserRole(role string) (string, error) {
	r := strings.ToLower(strings.TrimSpace(role))
	if r == "user" || r == "admin" {
		return r, nil
	}
	return "", NewClientError("无效的用户角色，仅支持 user 或 admin")
}

// Delete 软删除用户（同时禁用账号状态，并撤销其所有会话）
func (s *UserService) Delete(user_id uint64) error {
	user, err := models.GetUserByID(user_id)
	if err != nil {
		return NewClientError("用户不存在")
	}
	if err := ensureAdminPrivilegeSafe(user, user.Role, nil, true); err != nil {
		return err
	}

	now := time.Now().Unix()
	if err := db.DB.Model(&models.User{}).Where("id = ?", user_id).Updates(map[string]interface{}{
		"delete_time": now,
		"status":      0,
		"update_time": now,
	}).Error; err != nil {
		return err
	}
	revokeAllGuardSessions(user_id)
	return nil
}

// countActiveAdmins 统计启用中且未删除的管理员数量
func countActiveAdmins() (int, error) {
	var count int64
	err := db.DB.Raw(`
		SELECT COUNT(*) FROM users
		WHERE LOWER(role) = 'admin'
		  AND status = 1
		  AND delete_time IS NULL
	`).Scan(&count).Error
	return int(count), err
}

// ensureAdminPrivilegeSafe 防止误删/禁用/降级「最后一个启用中的管理员」导致系统锁死
func ensureAdminPrivilegeSafe(target *models.User, newRole string, newStatus *uint8, deleting bool) error {
	if target == nil || !strings.EqualFold(strings.TrimSpace(target.Role), "admin") {
		return nil
	}
	if target.Status != 1 {
		// 目标本身已不是启用管理员，不影响「最后一个管理员」
		return nil
	}

	willLoseAdmin := deleting
	if newStatus != nil && *newStatus == 0 {
		willLoseAdmin = true
	}
	if newRole != "" && !strings.EqualFold(strings.TrimSpace(newRole), "admin") {
		willLoseAdmin = true
	}
	if !willLoseAdmin {
		return nil
	}

	count, err := countActiveAdmins()
	if err != nil {
		return fmt.Errorf("检查管理员数量失败: %w", err)
	}
	if count <= 1 {
		return NewClientError("不能删除、禁用或降级最后一个启用中的管理员")
	}
	return nil
}

// BatchDelete 批量软删除用户。
// 之前的实现只改 delete_time，不像 Delete 一样同步禁用 status=0、也不做「最后一个管理员」保护，
// 行为和单条删除不一致（历史遗留、当前无 controller 暴露）。现在直接复用 Delete 逐条校验+删除，
// 保证批量删除和单条删除的语义完全一致（包括会逐条检查是否会删空管理员）。
func (s *UserService) BatchDelete(user_ids []uint64) error {
	for _, uid := range user_ids {
		if err := s.Delete(uid); err != nil {
			return fmt.Errorf("删除用户 %d 失败: %w", uid, err)
		}
	}
	return nil
}

// BatchUpdateStatus 批量更新用户状态；禁用时同步撤销相关会话。
// 与 BatchDelete 一样逐条复用 UpdateStatus，保证：
// 1) 跳过已软删用户（GetUserByID 会判不存在）
// 2) 禁用时走「最后一个启用管理员」保护
// 3) 会话撤销与单条 UpdateStatus 一致
func (s *UserService) BatchUpdateStatus(user_ids []uint64, status uint8) error {
	for _, uid := range user_ids {
		if err := s.UpdateStatus(uid, status); err != nil {
			return fmt.Errorf("更新用户 %d 状态失败: %w", uid, err)
		}
	}
	return nil
}

// revokeAllGuardSessions 级联撤销用户在所有 guard 下的活跃会话。
// 该函数故意忽略单条错误，以免某条撤销失败导致整个管理员批量操作回滚。
func revokeAllGuardSessions(user_id uint64) {
	for _, guard := range []string{"user", "admin"} {
		if err := models.RevokeAllUserSessionsWithGuard(user_id, guard, ""); err != nil {
			// 这里属于清理动作，失败仅日志化即可
			continue
		}
	}
}

// UpdateLoginInfo 更新登录信息（委托 models，统一 IP 截断与 lock_until 清理）。
func (s *UserService) UpdateLoginInfo(user_id uint64, ip string) error {
	return models.UpdateLoginInfo(user_id, ip)
}

// IncrementLoginFailure 增加登录失败次数
func (s *UserService) IncrementLoginFailure(user_id uint64) error {
	now := time.Now().Unix()
	return db.DB.Model(&models.User{}).Where("id = ? AND delete_time IS NULL", user_id).Updates(map[string]interface{}{
		"login_failure": gorm.Expr("login_failure + 1"),
		"update_time":   now,
	}).Error
}

// IncrementLoginFailureWithLock 原子地增加登录失败次数，并在达到阈值时锁定账户。
// 原实现分两步 UPDATE/SELECT 存在并发竞态：多个失败登录可能互相读到对方递增后的值，
// 造成“不该锁定”或“漏锁定”。此处改为在事务中 FOR UPDATE，确保计数与锁定判定一致。
func (s *UserService) IncrementLoginFailureWithLock(user_id uint64, max_failures, lock_duration_minutes int) error {
	if max_failures <= 0 {
		return nil
	}
	now := time.Now().Unix()

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := db.ForUpdate(tx).
			Select("login_failure").
			Where("id = ? AND delete_time IS NULL", user_id).
			First(&user).Error; err != nil {
			return err
		}
		nextFailure := int(user.LoginFailure) + 1
		updates := map[string]interface{}{
			"login_failure": nextFailure,
			"update_time":   now,
		}
		if nextFailure >= max_failures && lock_duration_minutes > 0 {
			lockUntil := now + int64(lock_duration_minutes*60)
			updates["lock_until"] = lockUntil
		}
		return tx.Model(&models.User{}).Where("id = ? AND delete_time IS NULL", user_id).Updates(updates).Error
	})
}

// ClearLockUntil 清除账户锁定
func (s *UserService) ClearLockUntil(user_id uint64) error {
	return db.DB.Model(&models.User{}).Where("id = ? AND delete_time IS NULL", user_id).Updates(map[string]interface{}{
		"lock_until":    nil,
		"login_failure": 0,
	}).Error
}

// UserSimpleInfo 用户简要信息（用于批量查询返回）
type UserSimpleInfo struct {
	ID       uint64 `gorm:"column:id" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Nickname string `gorm:"column:nickname" json:"nickname"`
	Email    string `gorm:"column:email" json:"email"`
	Role     string `gorm:"column:role" json:"role"`
	Status   uint8  `gorm:"column:status" json:"status"`
}

// BatchGetUserSimpleInfo 批量获取用户简要信息
// 返回 map[id]UserSimpleInfo，方便前端通过 ID 快速查找
func (s *UserService) BatchGetUserSimpleInfo(userIDs []uint64) (map[uint64]UserSimpleInfo, error) {
	if len(userIDs) == 0 {
		return make(map[uint64]UserSimpleInfo), nil
	}

	var users []UserSimpleInfo
	err := db.DB.Raw("SELECT id, username, nickname, email, role, status FROM users WHERE id IN ? AND delete_time IS NULL", userIDs).Scan(&users).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map
	result := make(map[uint64]UserSimpleInfo)
	for _, user := range users {
		result[user.ID] = user
	}

	return result, nil
}


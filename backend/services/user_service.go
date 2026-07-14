package services

import (
	"errors"
	"fst/internal/db"
	"fst/models"
	"fst/utils"
	"time"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
// 参数: req-注册请求
// 返回: 用户对象、token、错误
func (s *UserService) Register(req *models.UserRegisterRequest) (*models.User, string, error) {
	// 清理XSS
	utils.CleanXSSFields(&req.Username, &req.Nickname, &req.Email, &req.Mobile, &req.Country, &req.Region, &req.Language)

	// 检查用户名是否已存在
	var existingUser models.User
	err := db.GetDB().Table("users").Where("username", "=", req.Username).First(&existingUser)
	if err == nil {
		return nil, "", errors.New("用户名已被使用")
	}

	// 检查邮箱是否已存在
	err = db.GetDB().Table("users").Where("email", "=", req.Email).First(&existingUser)
	if err == nil {
		return nil, "", errors.New("邮箱已被使用")
	}

	// 哈希密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", errors.New("密码加密失败")
	}

	// 生成API密钥
	apiKey := utils.GenerateAPIKey()

	// 获取当前时间戳
	now := uint64(time.Now().Unix())

	// 创建用户
	user := &models.User{
		Username:   req.Username,
		Nickname:   req.Nickname,
		Email:      req.Email,
		Mobile:     req.Mobile,
		Password:   hashedPassword,
		Country:    req.Country,
		Region:     req.Region,
		Language:   req.Language,
		Role:       models.RoleUser,
		Status:     models.UserStatusActive,
		APIKey:     apiKey,
		JoinIP:     req.IP,
		JoinTime:   &now,
		CreateTime: &now,
		UpdateTime: &now,
	}

	// 保存用户
	if err := db.GetDB().Create(user); err != nil {
		return nil, "", errors.New("用户创建失败: " + err.Error())
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role), uint8(user.Status), user.Password)
	if err != nil {
		return nil, "", errors.New("令牌生成失败")
	}

	return user, token, nil
}

// Login 用户登录
// 参数: req-登录请求
// 返回: 用户对象、token、错误
func (s *UserService) Login(req *models.UserLoginRequest) (*models.User, string, error) {
	var user models.User
	dbInstance := db.GetDB()

	// 尝试用用户名或邮箱登录
	err := dbInstance.Table("users").
		Where("username", "=", req.Username).
		OrWhere("email", "=", req.Username).
		First(&user)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户名或密码错误")
		}
		return nil, "", errors.New("数据库查询失败")
	}

	// 检查用户状态
	if user.Status == models.UserStatusInactive {
		return nil, "", errors.New("账号已被禁用")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		// 增加登录失败次数
		dbInstance.Table("users").Where("id", "=", user.ID).Update(map[string]interface{}{
			"login_failure": user.LoginFailure + 1,
		})
		return nil, "", errors.New("用户名或密码错误")
	}

	// 登录成功，更新登录信息
	now := uint64(time.Now().Unix())
	err = dbInstance.Table("users").Where("id", "=", user.ID).Update(map[string]interface{}{
		"last_login_time": now,
		"last_login_ip":   req.IP,
		"login_failure":   0,
		"update_time":     now,
	})
	if err != nil {
		// 登录成功但更新失败，不影响登录流程
		// 记录日志即可
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role), uint8(user.Status), user.Password)
	if err != nil {
		return nil, "", errors.New("令牌生成失败")
	}

	return &user, token, nil
}

// GetUserByID 根据ID获取用户
// 参数: userID-用户ID
// 返回: 用户对象、错误
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	if userID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	var user models.User
	err := db.GetDB().Table("users").Where("id", "=", userID).First(&user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("查询用户失败")
	}

	return &user, nil
}

// UpdateUser 更新用户信息(用户自己)
// 参数: userID-用户ID, req-更新请求
// 返回: 错误
func (s *UserService) UpdateUser(userID uint, req *models.UserUpdateRequestForUser) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 检查用户是否存在
	var user models.User
	err := db.GetDB().Table("users").Where("id", "=", userID).First(&user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("查询用户失败")
	}

	// 清理XSS
	utils.CleanXSSFields(&req.Nickname, &req.Email, &req.Mobile, &req.Motto, &req.Country, &req.Region, &req.Language)

	// 检查邮箱是否被其他用户使用
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		err = db.GetDB().Table("users").Where("email", "=", req.Email).First(&existingUser)
		if err == nil && existingUser.ID != userID {
			return errors.New("邮箱已被其他用户使用")
		}
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Mobile != "" {
		updates["mobile"] = req.Mobile
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Background != "" {
		updates["background"] = req.Background
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		updates["birthday"] = *req.Birthday
	}
	if req.Motto != "" {
		updates["motto"] = req.Motto
	}
	if req.Country != "" {
		updates["country"] = req.Country
	}
	if req.Region != "" {
		updates["region"] = req.Region
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}

	// 更新时间戳
	now := uint64(time.Now().Unix())
	updates["update_time"] = now

	// 执行更新
	if len(updates) > 0 {
		err = db.GetDB().Table("users").Where("id", "=", userID).Update(updates)
		if err != nil {
			return errors.New("更新用户信息失败")
		}
	}

	return nil
}

// ChangePassword 修改密码
// 参数: userID-用户ID, req-修改密码请求
// 返回: 错误
func (s *UserService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 获取用户信息
	var user models.User
	err := db.GetDB().Table("users").Where("id", "=", userID).First(&user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("查询用户失败")
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return errors.New("旧密码不正确")
	}

	// 哈希新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	now := uint64(time.Now().Unix())
	err = db.GetDB().Table("users").Where("id", "=", userID).Update(map[string]interface{}{
		"password":    hashedPassword,
		"update_time": now,
	})
	if err != nil {
		return errors.New("密码更新失败")
	}

	return nil
}

// GetUserList 获取用户列表(管理员)
// 参数: page-页码, pageSize-每页数量, keyword-搜索关键词
// 返回: 用户列表、总数、错误
func (s *UserService) GetUserList(page, pageSize int, keyword string) ([]models.User, int64, error) {
	dbInstance := db.GetDB().Table("users")

	// 关键词搜索
	if keyword != "" {
		keyword = "%" + keyword + "%"
		dbInstance = dbInstance.Where("username", "LIKE", keyword).
			OrWhere("nickname", "LIKE", keyword).
			OrWhere("email", "LIKE", keyword).
			OrWhere("mobile", "LIKE", keyword)
	}

	// 统计总数
	total, err := dbInstance.Count()
	if err != nil {
		return nil, 0, errors.New("统计用户数量失败")
	}

	// 查询列表
	var users []models.User
	err = dbInstance.Page(page, pageSize).OrderBy("id DESC").Find(&users)
	if err != nil {
		return nil, 0, errors.New("查询用户列表失败")
	}

	return users, total, nil
}

// AdminUpdateUser 管理员更新用户信息
// 参数: userID-用户ID, req-更新请求
// 返回: 错误
func (s *UserService) AdminUpdateUser(userID uint, req *models.UserUpdateRequestForAdmin) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 检查用户是否存在
	var user models.User
	err := db.GetDB().Table("users").Where("id", "=", userID).First(&user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("查询用户失败")
	}

	// 清理XSS
	utils.CleanXSSFields(&req.Nickname, &req.Email, &req.Mobile, &req.Motto, &req.Country, &req.Region, &req.Language)

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Mobile != "" {
		updates["mobile"] = req.Mobile
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Background != "" {
		updates["background"] = req.Background
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		updates["birthday"] = *req.Birthday
	}
	if req.Motto != "" {
		updates["motto"] = req.Motto
	}
	if req.Country != "" {
		updates["country"] = req.Country
	}
	if req.Region != "" {
		updates["region"] = req.Region
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}
	if req.Money != nil {
		updates["money"] = *req.Money
	}
	if req.Score != nil {
		updates["score"] = *req.Score
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	// 更新时间戳
	now := uint64(time.Now().Unix())
	updates["update_time"] = now

	// 执行更新
	if len(updates) > 0 {
		err = db.GetDB().Table("users").Where("id", "=", userID).Update(updates)
		if err != nil {
			return errors.New("更新用户信息失败")
		}
	}

	return nil
}

// AdminResetPassword 管理员重置密码
// 参数: userID-用户ID, newPassword-新密码
// 返回: 错误
func (s *UserService) AdminResetPassword(userID uint, newPassword string) error {
	if userID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 哈希新密码
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	now := uint64(time.Now().Unix())
	err = db.GetDB().Table("users").Where("id", "=", userID).Update(map[string]interface{}{
		"password":    hashedPassword,
		"update_time": now,
	})
	if err != nil {
		return errors.New("密码重置失败")
	}

	return nil
}

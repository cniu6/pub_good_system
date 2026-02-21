package services

import (
	"errors"
	"fst/backend/app/models"
	"fst/backend/utils"
	"time"
)

// AuthService 认证服务
type AuthService struct {
	userService *UserService
}

func NewAuthService() *AuthService {
	return &AuthService{
		userService: NewUserService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"-"`
}

// LoginResult 登录结果
type LoginResult struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResult, error) {
	// 查找用户
	user, err := models.GetUserByUsernameOrEmail(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, errors.New("账号已被禁用")
	}

	// 检查登录失败次数
	if user.LoginFailure >= 5 {
		// 可以在这里添加锁定时间检查
		return nil, errors.New("登录失败次数过多，请稍后再试")
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		// 增加失败次数
		s.userService.IncrementLoginFailure(user.ID)
		return nil, errors.New("用户名或密码错误")
	}

	// 生成 Token
	access_token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	refresh_token, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成RefreshToken失败")
	}

	// 更新登录信息
	s.userService.UpdateLoginInfo(user.ID, req.IP)

	return &LoginResult{
		User:         user,
		AccessToken:  access_token,
		RefreshToken: refresh_token,
	}, nil
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	IP       string `json:"-"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	// 验证验证码
	valid, vcID, err := models.VerifyCode(req.Email, req.Code, "register")
	if err != nil || !valid {
		return nil, errors.New("验证码无效或已过期")
	}

	// 检查用户名是否已存在
	existing, _ := models.GetUserByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existing, _ = models.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashed_password, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	now := time.Now().Unix()
	user := &models.User{
		Username: req.Username,
		Password: hashed_password,
		Email:    req.Email,
		Nickname: req.Username,
		Role:     "user",
		Status:   1,
		JoinIp:   req.IP,
		JoinTime: &now,
	}

	err = models.CreateUser(user)
	if err != nil {
		return nil, errors.New("创建用户失败")
	}

	// 标记验证码为已使用
	if vcID > 0 {
		models.MarkVerificationCodeAsUsed(vcID)
	}

	return user, nil
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新 Token
func (s *AuthService) RefreshToken(refresh_token string) (*LoginResult, error) {
	// 解析 refresh token
	claims, err := utils.ParseRefreshToken(refresh_token)
	if err != nil {
		return nil, errors.New("无效的RefreshToken")
	}

	// 获取用户信息
	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, errors.New("账号已被禁用")
	}

	// 生成新 Token
	access_token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	new_refresh_token, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成RefreshToken失败")
	}

	return &LoginResult{
		User:         user,
		AccessToken:  access_token,
		RefreshToken: new_refresh_token,
	}, nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	UserID      uint64 `json:"-"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(req *ChangePasswordRequest) error {
	user, err := models.GetUserByID(req.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashed_password, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	return s.userService.UpdatePassword(req.UserID, hashed_password)
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(req *ResetPasswordRequest) error {
	// 验证验证码
	valid, vcID, err := models.VerifyCode(req.Email, req.Code, "reset_password")
	if err != nil || !valid {
		return errors.New("验证码无效或已过期")
	}

	// 获取用户
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 加密新密码
	hashed_password, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	err = s.userService.UpdatePassword(user.ID, hashed_password)
	if err != nil {
		return err
	}

	// 标记验证码为已使用
	if vcID > 0 {
		models.MarkVerificationCodeAsUsed(vcID)
	}

	return nil
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(token string) (*utils.Claims, error) {
	return utils.ParseToken(token)
}

// GetUserInfo 获取当前用户信息
func (s *AuthService) GetUserInfo(user_id uint64) (*models.User, error) {
	return models.GetUserByID(user_id)
}

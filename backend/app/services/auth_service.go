package services

import (
	"errors"
	"fst/backend/app/models"
	"fst/backend/internal/config"
	"fst/backend/utils"
	"fmt"
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

// ServiceError 服务层错误
type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{Code: code, Message: message}
}

// LoginResult 登录结果
type LoginResult struct {
	ID           uint64   `json:"id"`
	UserName     string   `json:"userName"`
	Email        string   `json:"email"`
	Role         []string `json:"role"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresAt    int64    `json:"expiresAt"`
}

// Login 用户登录
func (s *AuthService) Login(username, password, clientIP string) (*LoginResult, *ServiceError) {
	// 查找用户
	user, err := models.GetUserByUsernameOrEmail(username)
	if err != nil {
		return nil, NewServiceError(401, "Invalid account or password")
	}

	// 检查账户锁定
	now := time.Now().Unix()
	if user.LockUntil != nil && *user.LockUntil > now {
		remaining := (*user.LockUntil - now) / 60
		return nil, NewServiceError(403, fmt.Sprintf("Account is locked. Please try again in %d minutes", remaining))
	}

	// 清除过期锁定
	if user.LockUntil != nil && *user.LockUntil <= now {
		s.userService.ClearLockUntil(user.ID)
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, NewServiceError(403, "Account is inactive")
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		// 增加失败次数（带锁定）
		s.userService.IncrementLoginFailureWithLock(user.ID, config.GlobalConfig.LoginMaxFailureCount, config.GlobalConfig.LoginLockDurationMinutes)
		return nil, NewServiceError(401, "Invalid account or password")
	}

	// 更新登录信息
	s.userService.UpdateLoginInfo(user.ID, clientIP)

	// 生成 Token
	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second

	accessToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, accessTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	refreshToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, refreshTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	return &LoginResult{
		ID:           user.ID,
		UserName:     user.Username,
		Email:        user.Email,
		Role:         []string{user.Role},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Unix() + int64(accessTTL.Seconds()),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(user *models.User) error {
	// 加密密码
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.Password = hashedPassword

	// 设置默认值
	if user.Role == "" {
		user.Role = "user"
	}
	if user.Status == 0 {
		user.Status = 1
	}

	// 创建用户
	return models.CreateUser(user)
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResult, *ServiceError) {
	// 解析 token
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		return nil, NewServiceError(401, "Invalid or expired refresh token")
	}

	// 获取用户
	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		return nil, NewServiceError(401, "User not found")
	}

	// 检查状态
	if user.Status == 0 {
		return nil, NewServiceError(403, "Account is inactive")
	}

	// 生成新Token
	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second

	accessToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, accessTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	newRefreshToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, refreshTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	return &LoginResult{
		ID:           user.ID,
		UserName:     user.Username,
		Email:        user.Email,
		Role:         []string{user.Role},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Unix() + int64(accessTTL.Seconds()),
	}, nil
}

// UpdatePassword 更新密码
func (s *AuthService) UpdatePassword(userID uint64, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}
	return s.userService.UpdatePassword(userID, hashedPassword)
}

// ChangePassword 修改密码（需要验证旧密码）
func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	user, err := models.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return errors.New("incorrect old password")
	}

	return s.UpdatePassword(userID, newPassword)
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(token string) (*utils.Claims, error) {
	return utils.ParseToken(token)
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint64) (*models.User, error) {
	return models.GetUserByID(userID)
}

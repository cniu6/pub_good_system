package services

import (
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"log"
	"strings"
	"time"
)

// AuthService 认证服务
type AuthService struct {
	userService *UserService
}

// 管理员代登录(login-as / impersonation)会话的短时上限：无论是首次签发（LoginToUser）还是
// 后续 refresh 续期，最终生效的 TTL 都不能超过这两个值，避免代登 token 被 refresh 续成与
// 正常登录一样的长会话（安全修复，此前 RefreshToken 会统一按运行时默认 TTL 签发，导致代登
// 短时限制形同虚设）。
const (
	ImpersonationAccessTTLCap  = 15 * time.Minute
	ImpersonationRefreshTTLCap = 30 * time.Minute
)

// CapImpersonationTTL 按「取配置值与上限的较小值，且 refresh 不短于 access」的规则，
// 计算代登录场景下实际生效的 access/refresh TTL。configuredAccess/configuredRefresh <= 0
// 时视为未配置，直接用上限。LoginToUser 首次签发与 RefreshToken 续期共用同一份规则。
func CapImpersonationTTL(configuredAccess, configuredRefresh time.Duration) (accessTTL, refreshTTL time.Duration) {
	accessTTL = ImpersonationAccessTTLCap
	if configuredAccess > 0 && configuredAccess < accessTTL {
		accessTTL = configuredAccess
	}
	refreshTTL = ImpersonationRefreshTTLCap
	if configuredRefresh > 0 && configuredRefresh < refreshTTL {
		refreshTTL = configuredRefresh
	}
	if refreshTTL < accessTTL {
		refreshTTL = accessTTL
	}
	return
}

func NewAuthService() *AuthService {
	return &AuthService{
		userService: NewUserService(),
	}
}

func normalizeAuthGuard(authGuard string) (string, bool) {
	if authGuard == "" {
		return utils.UserAuthGuard, true
	}
	if authGuard == utils.UserAuthGuard || authGuard == utils.AdminAuthGuard {
		return authGuard, true
	}
	return "", false
}

// ServiceError 服务层错误
type ServiceError struct {
	Code    int
	Message string
}

type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}

func NewClientError(message string) error {
	return &ClientError{Message: message}
}

func IsClientError(err error) bool {
	var target *ClientError
	return errors.As(err, &target)
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{Code: code, Message: message}
}

// LoginResult 登录结果
type LoginRealnameSummary struct {
	HasVerification bool   `json:"hasVerification"`
	ID              uint64 `json:"id,omitempty"`
	Status          uint8  `json:"status,omitempty"`
	RealName        string `json:"realName,omitempty"`
	CertificateType uint8  `json:"certificateType,omitempty"`
	CertificateNo   string `json:"certificateNo,omitempty"`
	SubmittedAt     *int64 `json:"submittedAt,omitempty"`
	ReviewedAt      *int64 `json:"reviewedAt,omitempty"`
	RejectReason    string `json:"rejectReason,omitempty"`
}

type LoginResult struct {
	ID               uint64               `json:"id"`
	UserName         string               `json:"userName"`
	Email            string               `json:"email"`
	Role             []string             `json:"role"`
	// AuthGuard 本次会话签发的认证上下文（user/admin），前端必须按此值存会话并刷新，不可用 role 猜测。
	AuthGuard        string               `json:"authGuard"`
	AccessToken      string               `json:"accessToken"`
	RefreshToken     string               `json:"refreshToken"`
	ExpiresAt        int64                `json:"expiresAt"`
	RefreshExpiresAt int64                `json:"-"`
	Realname         LoginRealnameSummary `json:"realname,omitempty"`
}

func buildLoginRealnameSummary(userID uint64) LoginRealnameSummary {
	summary := LoginRealnameSummary{HasVerification: false}

	// 登录态也补齐实名摘要，避免前端首次登录后本地 userInfo 比 profile 接口少一拍。
	verification, err := models.GetRealnameVerificationByUserID(userID)
	if err != nil || verification == nil {
		return summary
	}

	summary.HasVerification = true
	summary.ID = verification.ID
	summary.Status = verification.Status
	summary.RealName = verification.RealName
	summary.CertificateType = verification.CertificateType
	// 用户端登录响应只回传掩码后的证件号，避免身份证号明文随登录态泄露
	summary.CertificateNo = utils.MaskCertificateNo(verification.CertificateNo)
	summary.SubmittedAt = verification.SubmittedAt
	summary.ReviewedAt = verification.ReviewedAt
	summary.RejectReason = verification.RejectReason
	return summary
}

// BuildLoginRealnameSummaryForAPI 复用登录态实名摘要构造，保持普通登录/刷新/管理员代登录返回一致。
func BuildLoginRealnameSummaryForAPI(userID uint64) LoginRealnameSummary {
	return buildLoginRealnameSummary(userID)
}

// validateUserAccessForGuard 统一校验当前用户状态与认证上下文是否匹配。
func validateUserAccessForGuard(user *models.User, authGuard string) *ServiceError {
	if user == nil {
		return NewServiceError(401, "User not found")
	}
	if user.Status == 0 {
		return NewServiceError(403, "Account is inactive")
	}
	if authGuard == utils.AdminAuthGuard && user.Role != utils.AdminAuthGuard {
		return NewServiceError(403, "Admin access only")
	}
	return nil
}

func getAuthRuntimeDefaults() (loginMaxFailureCount, loginLockDurationMinutes, accessExpireSeconds, refreshExpireSeconds int) {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 5, 10, 7200, 604800
	}

	loginMaxFailureCount = cfg.LoginMaxFailureCount
	if loginMaxFailureCount <= 0 {
		loginMaxFailureCount = 5
	}

	loginLockDurationMinutes = cfg.LoginLockDurationMinutes
	if loginLockDurationMinutes <= 0 {
		loginLockDurationMinutes = 10
	}

	accessExpireSeconds = cfg.JWTAccessExpire
	if accessExpireSeconds <= 0 {
		accessExpireSeconds = 7200
	}

	refreshExpireSeconds = cfg.JWTRefreshExpire
	if refreshExpireSeconds <= 0 {
		refreshExpireSeconds = 604800
	}

	return
}

// Login 用户登录
func (s *AuthService) Login(username, password, authGuard, clientIP string) (*LoginResult, *ServiceError) {
	rawGuard := authGuard
	var ok bool
	authGuard, ok = normalizeAuthGuard(authGuard)
	if !ok {
		log.Printf("[Auth] 登录拒绝: 非法 auth_guard=%q ip=%s", rawGuard, clientIP)
		return nil, NewServiceError(400, "Invalid auth guard")
	}

	// 关闭「允许用户登录」后：普通用户无法密码登录拿 JWT；管理员与 API Key 不受影响
	if authGuard == utils.UserAuthGuard && !GetGlobalAllowUserLogin() {
		log.Printf("[Auth] 登录拒绝: 用户登录已关闭 username=%s ip=%s", username, clientIP)
		return nil, NewServiceError(403, "User login is disabled")
	}

	// 查找用户
	user, err := models.GetUserByUsernameOrEmail(username)
	if err != nil {
		// 不区分「用户不存在 / 密码错误」，但审计侧需要失败尝试痕迹（不含密码）
		log.Printf("[Auth] 登录失败: 账号不存在或查询失败 username=%s guard=%s ip=%s", username, authGuard, clientIP)
		return nil, NewServiceError(401, "Invalid account or password")
	}

	// 检查账户锁定
	now := time.Now().Unix()
	if user.LockUntil != nil && *user.LockUntil > now {
		remaining := (*user.LockUntil - now) / 60
		log.Printf("[Auth] 登录拒绝: 账户锁定中 user_id=%d remaining_min=%d ip=%s", user.ID, remaining, clientIP)
		return nil, NewServiceError(403, fmt.Sprintf("Account is locked. Please try again in %d minutes", remaining))
	}

	// 清除过期锁定
	if user.LockUntil != nil && *user.LockUntil <= now {
		if clearErr := s.userService.ClearLockUntil(user.ID); clearErr != nil {
			log.Printf("[Auth] 清除过期锁定失败 user_id=%d: %v", user.ID, clearErr)
		}
	}

	if accessErr := validateUserAccessForGuard(user, authGuard); accessErr != nil {
		log.Printf("[Auth] 登录拒绝: 状态/角色不匹配 user_id=%d guard=%s code=%d msg=%s ip=%s",
			user.ID, authGuard, accessErr.Code, accessErr.Message, clientIP)
		return nil, accessErr
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		// 增加失败次数（带锁定）
		loginMaxFailureCount, loginLockDurationMinutes, _, _ := getAuthRuntimeDefaults()
		if incrErr := s.userService.IncrementLoginFailureWithLock(user.ID, loginMaxFailureCount, loginLockDurationMinutes); incrErr != nil {
			log.Printf("[Auth] 累加登录失败次数出错 user_id=%d: %v", user.ID, incrErr)
		}
		log.Printf("[Auth] 登录失败: 密码错误 user_id=%d username=%s guard=%s ip=%s", user.ID, user.Username, authGuard, clientIP)
		return nil, NewServiceError(401, "Invalid account or password")
	}

	// 更新登录信息
	if updErr := s.userService.UpdateLoginInfo(user.ID, clientIP); updErr != nil {
		// 登录信息写失败不阻断签发，但必须可观测
		log.Printf("[Auth] 更新登录信息失败 user_id=%d ip=%s: %v", user.ID, clientIP, updErr)
	}

	// 生成 Token
	_, _, accessExpireSeconds, refreshExpireSeconds := getAuthRuntimeDefaults()
	accessTTL := time.Duration(accessExpireSeconds) * time.Second
	refreshTTL := time.Duration(refreshExpireSeconds) * time.Second

	accessToken, err := utils.GenerateTokenForGuardWithTTL(user.ID, user.Role, authGuard, accessTTL)
	if err != nil {
		log.Printf("[Auth] 签发 access token 失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, authGuard, refreshTTL)
	if err != nil {
		log.Printf("[Auth] 签发 refresh token 失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	log.Printf("[Auth] 登录成功 user_id=%d username=%s guard=%s ip=%s", user.ID, user.Username, authGuard, clientIP)
	return &LoginResult{
		ID:               user.ID,
		UserName:         user.Username,
		Email:            user.Email,
		Role:             []string{user.Role},
		AuthGuard:        authGuard,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Unix() + int64(accessTTL.Seconds()),
		RefreshExpiresAt: time.Now().Unix() + int64(refreshTTL.Seconds()),
		Realname:         buildLoginRealnameSummary(user.ID),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(user *models.User) error {
	if err := validateClientRuneLen(user.Username, "用户名", utils.MaxUsernameLength); err != nil {
		return err
	}
	if err := validateClientRuneLen(user.Email, "邮箱", utils.MaxEmailLength); err != nil {
		return err
	}
	if err := validateClientRuneLen(user.Nickname, "昵称", utils.MaxNicknameLength); err != nil {
		return err
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// 设置默认值
	if user.Role == "" {
		user.Role = "user"
	}
	if user.Status == 0 {
		user.Status = 1
	}
	// 自助注册等级：后台「注册默认等级」，缺省 1（管理端建号可自行指定 level）
	if user.Level == 0 {
		user.Level = GetGlobalRegisterDefaultLevel()
	}
	// 注册时自动发放 API Key，避免用户中心显示「暂无」还要点重置
	if user.Apikey == nil || strings.TrimSpace(*user.Apikey) == "" {
		key := models.GenerateApiKey()
		user.Apikey = &key
	}

	// 创建用户
	return models.CreateUser(user)
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(refreshToken, authGuard, clientIP, userAgent, device string) (*LoginResult, *ServiceError) {
	rawGuard := authGuard
	var ok bool
	authGuard, ok = normalizeAuthGuard(authGuard)
	if !ok {
		log.Printf("[Auth] 刷新拒绝: 非法 auth_guard=%q ip=%s", rawGuard, clientIP)
		return nil, NewServiceError(400, "Invalid auth guard")
	}

	// 解析 token
	claims, err := utils.ParseRefreshTokenForGuard(refreshToken, authGuard)
	if err != nil {
		log.Printf("[Auth] 刷新失败: refresh token 无效/过期 guard=%s ip=%s", authGuard, clientIP)
		return nil, NewServiceError(401, "Invalid or expired refresh token")
	}

	// 获取用户
	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("[Auth] 刷新失败: 用户不存在 user_id=%d guard=%s ip=%s", claims.UserID, authGuard, clientIP)
		return nil, NewServiceError(401, "User not found")
	}

	if accessErr := validateUserAccessForGuard(user, authGuard); accessErr != nil {
		log.Printf("[Auth] 刷新拒绝: 状态/角色不匹配 user_id=%d guard=%s code=%d msg=%s ip=%s",
			user.ID, authGuard, accessErr.Code, accessErr.Message, clientIP)
		return nil, accessErr
	}

	// 刷新令牌重放检测：轮换前先确认该 refresh token 仍是某个活跃会话的当前令牌。
	// 若未命中（已被轮换过 / 已过期 / 已被踢下线），视为潜在的令牌泄露或重放攻击，
	// 主动吊销该用户在此 guard 下的全部会话，阻止攻击者用同一枚旧 refresh token 继续得手。
	// 提前到 Token 生成之前做这个检测：既避免无效请求也生成 Token 浪费开销，也方便下面
	// 借同一个 refreshTokenHash 判断是否为管理员代登录会话。
	refreshTokenHash := utils.HashToken(refreshToken)
	active, err := models.IsRefreshSessionActive(user.ID, authGuard, refreshTokenHash)
	if err != nil {
		log.Printf("[Auth] 校验 refresh 会话失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to validate refresh session")
	}
	if !active {
		if revokeErr := models.RevokeAllUserSessionsWithGuard(user.ID, authGuard, ""); revokeErr != nil {
			log.Printf("[Auth] 检测到 refresh token 重放，吊销会话失败 user_id=%d guard=%s: %v", user.ID, authGuard, revokeErr)
		} else {
			log.Printf("[Auth] 检测到 refresh token 重放/失效，已吊销该用户全部会话 user_id=%d guard=%s", user.ID, authGuard)
		}
		return nil, NewServiceError(401, "Refresh session expired or revoked")
	}

	// 生成新Token
	_, _, accessExpireSeconds, refreshExpireSeconds := getAuthRuntimeDefaults()
	accessTTL := time.Duration(accessExpireSeconds) * time.Second
	refreshTTL := time.Duration(refreshExpireSeconds) * time.Second

	// 安全修复：管理员代登录(login-as)会话续期时仍需保持短时上限，避免代登 token 通过
	// refresh 变成与正常登录一样的长会话（此前这里统一按运行时默认 TTL 签发，代登的短时
	// 限制形同虚设）。
	isImpersonation, err := models.IsImpersonationSession(user.ID, authGuard, refreshTokenHash)
	if err != nil {
		log.Printf("[Auth] 判断代登会话失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to validate refresh session")
	}
	// 代登会话轮换时必须保留 device 标记，否则第二次 refresh 会丢失 impersonation 身份，
	// 短时 TTL 上限只生效一次后又被拉回正常长会话。
	rotateDevice := device
	if isImpersonation {
		accessTTL, refreshTTL = CapImpersonationTTL(accessTTL, refreshTTL)
		rotateDevice = models.ImpersonationDeviceLabel
	}

	accessToken, err := utils.GenerateTokenForGuardWithTTL(user.ID, user.Role, authGuard, accessTTL)
	if err != nil {
		log.Printf("[Auth] 刷新签发 access token 失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	newRefreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, authGuard, refreshTTL)
	if err != nil {
		log.Printf("[Auth] 刷新签发 refresh token 失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	accessExpiresAt := time.Now().Unix() + int64(accessTTL.Seconds())
	refreshExpiresAt := time.Now().Unix() + int64(refreshTTL.Seconds())
	rotated, err := models.RotateUserSessionTokens(
		user.ID,
		authGuard,
		refreshTokenHash,
		utils.HashToken(accessToken),
		utils.HashToken(newRefreshToken),
		clientIP,
		userAgent,
		rotateDevice,
		accessExpiresAt,
		refreshExpiresAt,
	)
	if err != nil {
		log.Printf("[Auth] 会话令牌轮换失败 user_id=%d guard=%s: %v", user.ID, authGuard, err)
		return nil, NewServiceError(500, "Failed to rotate session tokens")
	}
	if !rotated {
		// 并发 refresh / 会话刚被踢：与「重放」不同，这里不再二次吊销（active 检测已通过）
		log.Printf("[Auth] 会话令牌轮换未命中(并发/已失效) user_id=%d guard=%s ip=%s impersonation=%v",
			user.ID, authGuard, clientIP, isImpersonation)
		return nil, NewServiceError(401, "Refresh session expired or revoked")
	}

	return &LoginResult{
		ID:               user.ID,
		UserName:         user.Username,
		Email:            user.Email,
		Role:             []string{user.Role},
		AuthGuard:        authGuard,
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		Realname:         buildLoginRealnameSummary(user.ID),
	}, nil
}

// UpdatePassword 更新密码
func (s *AuthService) UpdatePassword(userID uint64, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.userService.UpdatePassword(userID, hashedPassword)
}

// ChangePassword 修改密码（需要验证旧密码）
func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Printf("[Auth] 改密失败: 用户不存在 user_id=%d", userID)
		return NewClientError("用户不存在")
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		log.Printf("[Auth] 改密失败: 原密码不正确 user_id=%d", userID)
		return NewClientError("原密码不正确")
	}
	if oldPassword == newPassword {
		return NewClientError("新密码不能与旧密码相同")
	}

	if err := s.UpdatePassword(userID, newPassword); err != nil {
		log.Printf("[Auth] 改密写库失败 user_id=%d: %v", userID, err)
		return err
	}
	log.Printf("[Auth] 改密成功并已吊销全部会话 user_id=%d", userID)
	return nil
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(token string) (*utils.Claims, error) {
	return utils.ParseToken(token)
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint64) (*models.User, error) {
	return models.GetUserByID(userID)
}

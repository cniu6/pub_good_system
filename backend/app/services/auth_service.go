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
	AccessToken      string               `json:"accessToken"`
	RefreshToken     string               `json:"refreshToken"`
	ExpiresAt        int64                `json:"expiresAt"`
	RefreshExpiresAt int64                `json:"-"`
	Realname         LoginRealnameSummary `json:"realname,omitempty"`
	// 管理端 TOTP 第二步：NeedTOTP=true 时只回 temp_token，不签发正式 access/refresh
	NeedTOTP  bool   `json:"need_totp,omitempty"`
	TempToken string `json:"temp_token,omitempty"`
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
func (s *AuthService) Login(username, password, authGuard, clientType, clientIP string) (*LoginResult, *ServiceError) {
	var ok bool
	authGuard, ok = normalizeAuthGuard(authGuard)
	if !ok {
		return nil, NewServiceError(400, "Invalid auth guard")
	}

	// 「禁止网页端登录」开关：仅拦截普通用户（非管理员）且客户端类型判定为 web 的登录请求；
	// App/小程序等客户端只要在登录请求带上 client_type=app 即可绕过此限制，管理员登录完全不受影响。
	// 用于「仅通过小程序/App 对外提供服务，网页面板只留管理员使用」的场景。
	// 安全定位：client_type 由请求自身携带、不做客户端可信校验（无签名/证书校验），
	// 这是一个引导前端 UX 的软限制（正规客户端不带这个参数登录就会被拦），不是安全边界，
	// 不能用来防止「有意伪造 client_type=app 的攻击者」绕过网页限制。
	if authGuard == utils.UserAuthGuard && models.NormalizeClientType(clientType) == "web" && GetGlobalDisableWebLogin() {
		return nil, NewServiceError(403, "Web login is disabled")
	}

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

	if accessErr := validateUserAccessForGuard(user, authGuard); accessErr != nil {
		return nil, accessErr
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		// 增加失败次数（带锁定）
		loginMaxFailureCount, loginLockDurationMinutes, _, _ := getAuthRuntimeDefaults()
		s.userService.IncrementLoginFailureWithLock(user.ID, loginMaxFailureCount, loginLockDurationMinutes)
		return nil, NewServiceError(401, "Invalid account or password")
	}

	// 管理端已启用 TOTP：密码通过后不签发正式 Token，返回临时令牌走第二步
	if authGuard == utils.AdminAuthGuard && user.TotpEnabled {
		tempToken, terr := utils.GenerateTOTPPendingToken(user.ID, user.Role, 5*time.Minute)
		if terr != nil {
			return nil, NewServiceError(500, "Failed to generate temp token")
		}
		return &LoginResult{
			ID:        user.ID,
			UserName:  user.Username,
			Email:     user.Email,
			Role:      []string{user.Role},
			NeedTOTP:  true,
			TempToken: tempToken,
		}, nil
	}

	// 更新登录信息
	s.userService.UpdateLoginInfo(user.ID, clientIP)

	// 生成 Token
	_, _, accessExpireSeconds, refreshExpireSeconds := getAuthRuntimeDefaults()
	accessTTL := time.Duration(accessExpireSeconds) * time.Second
	refreshTTL := time.Duration(refreshExpireSeconds) * time.Second

	accessToken, err := utils.GenerateTokenForGuardWithTTL(user.ID, user.Role, authGuard, accessTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, authGuard, refreshTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	return &LoginResult{
		ID:               user.ID,
		UserName:         user.Username,
		Email:            user.Email,
		Role:             []string{user.Role},
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Unix() + int64(accessTTL.Seconds()),
		RefreshExpiresAt: time.Now().Unix() + int64(refreshTTL.Seconds()),
		Realname:         buildLoginRealnameSummary(user.ID),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(user *models.User) error {
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
	var ok bool
	authGuard, ok = normalizeAuthGuard(authGuard)
	if !ok {
		return nil, NewServiceError(400, "Invalid auth guard")
	}

	// 解析 token
	claims, err := utils.ParseRefreshTokenForGuard(refreshToken, authGuard)
	if err != nil {
		return nil, NewServiceError(401, "Invalid or expired refresh token")
	}

	// 获取用户
	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		return nil, NewServiceError(401, "User not found")
	}

	if accessErr := validateUserAccessForGuard(user, authGuard); accessErr != nil {
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
		return nil, NewServiceError(500, "Failed to generate access token")
	}

	newRefreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, authGuard, refreshTTL)
	if err != nil {
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
		return nil, NewServiceError(500, "Failed to rotate session tokens")
	}
	if !rotated {
		return nil, NewServiceError(401, "Refresh session expired or revoked")
	}

	return &LoginResult{
		ID:               user.ID,
		UserName:         user.Username,
		Email:            user.Email,
		Role:             []string{user.Role},
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
		return NewClientError("用户不存在")
	}

	// 验证旧密码
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return NewClientError("原密码不正确")
	}
	if oldPassword == newPassword {
		return NewClientError("新密码不能与旧密码相同")
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

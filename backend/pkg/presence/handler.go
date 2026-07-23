package presence

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const heartbeatPersistInterval = 20 * time.Second

// HandlePresence 建立独立的在线状态 WebSocket；不挂通用 AuthMiddleware，避免 Header-only 鉴权限制。
func HandlePresence(c *gin.Context) {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin != "" {
		allowed, _ := middleware.IsOriginAllowed(origin, middleware.ResolveWSCorsAllowlist(c))
		if !allowed {
			c.Abort()
			utils.Fail(c, http.StatusForbidden, "origin not allowed")
			return
		}
	}

	ip := c.ClientIP()
	if !handshakeLimiter.allow(ip) {
		c.Abort()
		utils.Fail(c, http.StatusTooManyRequests, "too many websocket handshakes")
		return
	}

	ticket := strings.TrimSpace(c.Query("ticket"))
	if ticket == "" {
		// 兼容过渡期：仍拒绝裸 JWT 出现在 query，避免日志/Referer 泄露
		if strings.TrimSpace(c.Query("token")) != "" {
			c.Abort()
			utils.Fail(c, http.StatusUnauthorized, "use ticket instead of token")
			return
		}
		c.Abort()
		utils.Fail(c, http.StatusUnauthorized, "ticket is required")
		return
	}

	wt, err := ConsumeWSTicket(ticket)
	if err != nil || wt == nil {
		c.Abort()
		utils.Fail(c, http.StatusUnauthorized, "invalid or expired ticket")
		return
	}
	session, err := models.GetUserSessionByID(wt.SessionID)
	now := time.Now().Unix()
	if err != nil || session == nil || session.UserID != wt.UserID || session.AuthGuard != wt.Guard ||
		!session.IsActive || session.ExpiresAt <= now {
		c.Abort()
		utils.Fail(c, http.StatusUnauthorized, "session expired or revoked")
		return
	}

	// 同浏览器实例 ID：多标签重复登录时，只保留当前会话，踢掉同浏览器其它会话。
	browserID := strings.TrimSpace(c.Query("browser_id"))
	if browserID == "" {
		browserID = strings.TrimSpace(c.GetHeader("X-Browser-Id"))
	}
	if browserID != "" {
		if session.BrowserID == "" {
			_ = models.BindSessionBrowserID(session.ID, browserID)
			session.BrowserID = browserID
		}
		revokedIDs := make([]uint64, 0)
		if revoked, revErr := models.RevokeOtherSessionsByBrowserID(session.UserID, session.AuthGuard, browserID, strconv.FormatUint(session.ID, 10)); revErr == nil {
			revokedIDs = append(revokedIDs, revoked...)
		}
		// 兼容旧会话（browser_id 为空）：同 UA 的 web 多会话一并收口，避免多标签堆出 2~3 行。
		if revoked, revErr := models.RevokeSiblingWebSessionsByUA(session.UserID, session.AuthGuard, session.UserAgent, strconv.FormatUint(session.ID, 10)); revErr == nil {
			revokedIDs = append(revokedIDs, revoked...)
		}
		if len(revokedIDs) > 0 {
			DefaultHub().KickMany(revokedIDs, "browser_session_replaced")
		}
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(_ *http.Request) bool { return true }, // 上方已按 WS_CORS 完成校验。
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &Client{
		SessionID: session.ID,
		UserID:    session.UserID,
		conn:      conn,
		// 每条连接独立限流，心跳以外的业务消息最多每两秒一条。
		messages: newFixedWindowLimiter(1, 2*time.Second),
	}
	if !DefaultHub().Register(client) {
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "too_many_connections"), time.Now().Add(time.Second))
		_ = conn.Close()
		return
	}
	defer func() {
		DefaultHub().Unregister(client)
		_ = conn.Close()
	}()

	// 建连即写入心跳，确保短连接也能被管理端看见。
	_ = models.TouchSessionLastSeen(session.ID)
	var lastTouchMu sync.Mutex
	lastTouch := time.Now()
	touch := func() {
		lastTouchMu.Lock()
		defer lastTouchMu.Unlock()
		if time.Since(lastTouch) >= heartbeatPersistInterval {
			_ = models.TouchSessionLastSeen(session.ID)
			lastTouch = time.Now()
		}
	}

	// 读超时按「上报周期」动态换算的容忍窗口 +30s 兜底，避免管理端调整上报周期后连接被误判超时断开。
	readDeadline := func() time.Duration {
		return time.Duration(services.GetGlobalOnlinePresenceRuntimeConfig().GraceSeconds+30) * time.Second
	}
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(readDeadline()))
	conn.SetPongHandler(func(string) error {
		touch()
		return conn.SetReadDeadline(time.Now().Add(readDeadline()))
	})
	conn.SetPingHandler(func(message string) error {
		touch()
		client.writeMu.Lock()
		defer client.writeMu.Unlock()
		return conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second))
	})

	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		if !client.messages.allow("message") {
			client.forceClose("message_rate_limited")
			return
		}
		touch()
	}
}

func parsePresenceToken(token string) (*utils.Claims, string) {
	for _, guard := range []string{utils.UserAuthGuard, utils.AdminAuthGuard} {
		claims, err := utils.ParseTokenForGuard(token, guard)
		if err == nil {
			return claims, guard
		}
	}
	return nil, ""
}

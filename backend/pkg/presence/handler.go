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

	// 先只读取 browser ID；会话收口必须在 Upgrade + Register 成功后执行。
	// 否则新 WS 被代理/网络拒绝时，旧会话已经被提前撤销并踢下线。
	browserID := strings.TrimSpace(c.Query("browser_id"))
	if browserID == "" {
		browserID = strings.TrimSpace(c.GetHeader("X-Browser-Id"))
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

	// 同浏览器实例 ID：多标签重复登录时只保留当前会话。
	// 放在 Register 成功后，确保新连接已经真实可用才撤销旧会话。
	if browserID != "" {
		if session.BrowserID == "" {
			_ = models.BindSessionBrowserID(session.ID, browserID)
			session.BrowserID = browserID
		}
		revokedIDs := make([]uint64, 0)
		if revoked, revErr := models.RevokeOtherSessionsByBrowserID(session.UserID, session.AuthGuard, browserID, strconv.FormatUint(session.ID, 10)); revErr == nil {
			revokedIDs = append(revokedIDs, revoked...)
		}
		// 兼容升级前 browser_id 为空的旧会话，按 UA 做一次收口。
		if revoked, revErr := models.RevokeSiblingWebSessionsByUA(session.UserID, session.AuthGuard, session.UserAgent, strconv.FormatUint(session.ID, 10)); revErr == nil {
			revokedIDs = append(revokedIDs, revoked...)
		}
		if len(revokedIDs) > 0 {
			DefaultHub().KickMany(revokedIDs, "browser_session_replaced")
		}
	}

	// 建连即写入心跳，确保短连接也能被管理端看见。
	active, touchErr := models.TouchSessionLastSeen(session.ID)
	if touchErr != nil || !active {
		client.forceClose("session_revoked")
		return
	}
	var lastTouchMu sync.Mutex
	lastTouch := time.Now()
	touch := func() bool {
		lastTouchMu.Lock()
		defer lastTouchMu.Unlock()
		if time.Since(lastTouch) < heartbeatPersistInterval {
			return true
		}
		active, err := models.TouchSessionLastSeen(session.ID)
		if err != nil {
			// 短暂数据库故障不能把全体在线用户误踢；公告广播前的复核与下一次心跳会继续兜底。
			return true
		}
		if active {
			lastTouch = time.Now()
		}
		return active
	}

	// 读超时按「上报周期」动态换算的容忍窗口 +30s 兜底，避免管理端调整上报周期后连接被误判超时断开。
	readDeadline := func() time.Duration {
		return time.Duration(services.GetGlobalOnlinePresenceRuntimeConfig().GraceSeconds+30) * time.Second
	}
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(readDeadline()))
	conn.SetPongHandler(func(string) error {
		if !touch() {
			client.forceClose("session_revoked")
			return websocket.ErrCloseSent
		}
		return conn.SetReadDeadline(time.Now().Add(readDeadline()))
	})
	conn.SetPingHandler(func(message string) error {
		if !touch() {
			client.forceClose("session_revoked")
			return websocket.ErrCloseSent
		}
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
		if !touch() {
			client.forceClose("session_revoked")
			return
		}
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

package presence

import (
	"encoding/json"
	"fst/backend/app/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 表示一个已鉴权的 Presence WebSocket 连接。
type Client struct {
	SessionID uint64
	UserID    uint64
	conn      *websocket.Conn
	writeMu   sync.Mutex
	closeOnce sync.Once
	messages  *fixedWindowLimiter
}

func (c *Client) forceClose(reason string) {
	c.closeOnce.Do(func() {
		c.writeMu.Lock()
		// 同会话重连只关旧连接，勿发 force_logout，避免多标签/刷新误清登录态。
		if reason != "session_reconnected" {
			_ = c.conn.WriteJSON(map[string]string{"type": "force_logout", "reason": reason})
		}
		_ = c.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason), time.Now().Add(time.Second))
		c.writeMu.Unlock()
		_ = c.conn.Close()
	})
}

// Hub 是单机进程内连接表。多实例部署时需由 Redis 等消息总线同步 Kick。
type Hub struct {
	mu        sync.Mutex
	bySession map[uint64]*Client
	perUser   map[uint64]int
}

func NewHub() *Hub {
	return &Hub{bySession: make(map[uint64]*Client), perUser: make(map[uint64]int)}
}

// Register 注册连接；相同会话的新连接会替换并关闭旧连接。
func (h *Hub) Register(client *Client) bool {
	h.mu.Lock()
	old := h.bySession[client.SessionID]
	if old != nil {
		delete(h.bySession, client.SessionID)
		h.perUser[old.UserID]--
	}
	if h.perUser[client.UserID] >= 5 {
		h.mu.Unlock()
		return false
	}
	h.bySession[client.SessionID] = client
	h.perUser[client.UserID]++
	h.mu.Unlock()
	if old != nil {
		old.forceClose("session_reconnected")
	}
	return true
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if h.bySession[client.SessionID] == client {
		delete(h.bySession, client.SessionID)
		h.perUser[client.UserID]--
		if h.perUser[client.UserID] <= 0 {
			delete(h.perUser, client.UserID)
		}
	}
	h.mu.Unlock()
}

// Kick 强制下线一个会话。
func (h *Hub) Kick(sessionID uint64, reason ...string) {
	reasonText := "session_revoked"
	if len(reason) > 0 && reason[0] != "" {
		reasonText = reason[0]
	}
	h.mu.Lock()
	client := h.bySession[sessionID]
	h.mu.Unlock()
	if client != nil {
		client.forceClose(reasonText)
	}
}

func (h *Hub) KickMany(sessionIDs []uint64, reason ...string) {
	for _, sessionID := range sessionIDs {
		h.Kick(sessionID, reason...)
	}
}

// MarshalForceLogout 供测试或其它传输层复用退出消息格式。
func MarshalForceLogout(reason string) []byte {
	data, _ := json.Marshal(map[string]string{"type": "force_logout", "reason": reason})
	return data
}

// writeJSON 向单个连接写 JSON（带写锁）
func (c *Client) writeJSON(v interface{}) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_ = c.conn.WriteJSON(v)
}

// canReceiveBroadcast 在业务推送前复核 DB 会话状态。
// 主动撤销不一定发生在能直接访问 Hub 的控制器里；这里兜底防止已撤销/已过期的 WS
// 在下一个心跳到来前继续收到公告等广播。查询异常时不误踢所有正常用户，由心跳复核再次处理。
func (h *Hub) canReceiveBroadcast(client *Client) bool {
	active, err := models.IsUserSessionActiveByID(client.SessionID)
	if err != nil {
		return true
	}
	if active {
		return true
	}
	client.forceClose("session_revoked")
	return false
}

// BroadcastJSON 向所有在线连接广播一条业务消息（如公告发布）。单机 Hub；多实例需总线。
func (h *Hub) BroadcastJSON(v interface{}) {
	h.mu.Lock()
	clients := make([]*Client, 0, len(h.bySession))
	for _, c := range h.bySession {
		clients = append(clients, c)
	}
	h.mu.Unlock()
	for _, c := range clients {
		if !h.canReceiveBroadcast(c) {
			continue
		}
		c.writeJSON(v)
	}
}

// BroadcastToUserJSON 向指定用户的所有在线连接推送
func (h *Hub) BroadcastToUserJSON(userID uint64, v interface{}) {
	h.mu.Lock()
	clients := make([]*Client, 0)
	for _, c := range h.bySession {
		if c.UserID == userID {
			clients = append(clients, c)
		}
	}
	h.mu.Unlock()
	for _, c := range clients {
		if !h.canReceiveBroadcast(c) {
			continue
		}
		c.writeJSON(v)
	}
}

var defaultHub = NewHub()

// DefaultHub 返回全局单机 Hub。
func DefaultHub() *Hub { return defaultHub }

package presence

import (
	"fst/backend/app/models"
	"fst/backend/pkg/db"
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

// writeJSON 向单个连接写 JSON（带写锁）
func (c *Client) writeJSON(v interface{}) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_ = c.conn.WriteJSON(v)
}

// broadcastSessionCheckBatch 广播前批量复核会话的每批大小。
// 单条连接逐次查库在 5 万在线时会变成 5 万次 COUNT，必须按批查询。
const broadcastSessionCheckBatch = 500

// activeSessionIDs 批量返回仍未撤销且未过期的会话 ID。
// 查询异常时按“活跃”处理，保持原行为：不因一次 DB 抖动踢掉全体在线用户，由心跳继续兜底。
func activeSessionIDs(sessionIDs []uint64) map[uint64]bool {
	active := make(map[uint64]bool, len(sessionIDs))
	now := time.Now().Unix()
	for start := 0; start < len(sessionIDs); start += broadcastSessionCheckBatch {
		end := min(start+broadcastSessionCheckBatch, len(sessionIDs))
		chunk := sessionIDs[start:end]
		var found []uint64
		err := db.DB.Model(&models.UserSession{}).
			Where("id IN ? AND is_active = ? AND expires_at > ?", chunk, true, now).
			Pluck("id", &found).Error
		if err != nil {
			for _, id := range chunk {
				active[id] = true
			}
			continue
		}
		for _, id := range found {
			active[id] = true
		}
	}
	return active
}

// BroadcastJSON 向所有在线连接广播一条业务消息（如公告发布）。单机 Hub；多实例需总线。
func (h *Hub) BroadcastJSON(v interface{}) {
	h.mu.Lock()
	clients := make([]*Client, 0, len(h.bySession))
	ids := make([]uint64, 0, len(h.bySession))
	for _, c := range h.bySession {
		clients = append(clients, c)
		ids = append(ids, c.SessionID)
	}
	h.mu.Unlock()

	active := activeSessionIDs(ids)
	for _, c := range clients {
		if !active[c.SessionID] {
			c.forceClose("session_revoked")
			continue
		}
		c.writeJSON(v)
	}
}


var defaultHub = NewHub()

// DefaultHub 返回全局单机 Hub。
func DefaultHub() *Hub { return defaultHub }

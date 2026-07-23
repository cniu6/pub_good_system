package presence

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// WSTicket 短时一次性 Presence 票据（替代 URL ?token= JWT）
type WSTicket struct {
	Token     string
	UserID    uint64
	Guard     string
	SessionID uint64
	ExpiresAt time.Time
}

type ticketStore struct {
	mu   sync.Mutex
	data map[string]WSTicket
}

var globalTickets = &ticketStore{data: make(map[string]WSTicket)}

const wsTicketTTL = 60 * time.Second

// IssueWSTicket 签发一次性票据，绑定用户/guard/会话
func IssueWSTicket(userID uint64, guard string, sessionID uint64) (string, time.Time, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	id := hex.EncodeToString(buf)
	exp := time.Now().Add(wsTicketTTL)
	globalTickets.mu.Lock()
	defer globalTickets.mu.Unlock()
	// 顺带清理过期项，避免无限增长
	now := time.Now()
	for k, v := range globalTickets.data {
		if now.After(v.ExpiresAt) {
			delete(globalTickets.data, k)
		}
	}
	globalTickets.data[id] = WSTicket{
		Token:     id,
		UserID:    userID,
		Guard:     guard,
		SessionID: sessionID,
		ExpiresAt: exp,
	}
	return id, exp, nil
}

// ConsumeWSTicket 单次消费；成功返回票据信息，失败返回错误
func ConsumeWSTicket(ticket string) (*WSTicket, error) {
	ticket = trimTicket(ticket)
	if ticket == "" {
		return nil, fmt.Errorf("ticket required")
	}
	globalTickets.mu.Lock()
	defer globalTickets.mu.Unlock()
	item, ok := globalTickets.data[ticket]
	if !ok {
		return nil, fmt.Errorf("invalid ticket")
	}
	delete(globalTickets.data, ticket)
	if time.Now().After(item.ExpiresAt) {
		return nil, fmt.Errorf("ticket expired")
	}
	cp := item
	return &cp, nil
}

func trimTicket(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

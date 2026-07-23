package presence

import (
	"testing"
	"time"
)

func TestWSTicketConsumeOnce(t *testing.T) {
	id, exp, err := IssueWSTicket(7, "user", 99)
	if err != nil || id == "" || exp.Before(time.Now()) {
		t.Fatalf("issue failed: id=%q err=%v", id, err)
	}
	got, err := ConsumeWSTicket(id)
	if err != nil || got == nil || got.UserID != 7 || got.SessionID != 99 {
		t.Fatalf("consume failed: %+v err=%v", got, err)
	}
	if _, err := ConsumeWSTicket(id); err == nil {
		t.Fatal("ticket should be single-use")
	}
}

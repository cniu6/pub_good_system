package pay_balance

import (
	"fst/backend/app/services"
	"testing"
)

func TestPluginMeta(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "pay_balance" {
		t.Fatalf("Name() = %q, want pay_balance", p.Name())
	}
	if p.Version() == "" {
		t.Fatal("Version() should not be empty")
	}
	if p.Description() == "" {
		t.Fatal("Description() should not be empty")
	}
}

func TestPluginInitRegistersEpayChannel(t *testing.T) {
	services.ClearPaymentChannels()
	t.Cleanup(services.ClearPaymentChannels)

	p := NewPlugin()
	if err := p.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	ch, ok := services.GetPaymentChannel("epay")
	if !ok || ch == nil {
		t.Fatal("Init should register epay channel")
	}
	if ch.Type() != "epay" {
		t.Fatalf("channel Type() = %q", ch.Type())
	}
}

func TestPluginInitIdempotent(t *testing.T) {
	services.ClearPaymentChannels()
	t.Cleanup(services.ClearPaymentChannels)

	p := NewPlugin()
	if err := p.Init(); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if err := p.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}

	types := services.ListPaymentChannelTypes()
	if len(types) != 1 || types[0] != "epay" {
		t.Fatalf("expected only epay once, got %v", types)
	}
}

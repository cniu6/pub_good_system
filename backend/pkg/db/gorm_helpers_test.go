package db

import (
	"strings"
	"testing"
)

func TestCastToTextAndQuoteIdent(t *testing.T) {
	prev := activeDriver
	t.Cleanup(func() { activeDriver = prev })

	activeDriver = "mysql"
	if got := CastToText("money"); got != "CAST(money AS CHAR)" {
		t.Fatalf("mysql CastToText: %s", got)
	}
	if got := QuoteIdent("contact"); got != "`contact`" {
		t.Fatalf("mysql QuoteIdent: %s", got)
	}

	activeDriver = "postgres"
	if got := CastToText("money"); got != "CAST(money AS TEXT)" {
		t.Fatalf("postgres CastToText: %s", got)
	}
	if got := QuoteIdent("contact"); got != `"contact"` {
		t.Fatalf("postgres QuoteIdent: %s", got)
	}

	activeDriver = "sqlite"
	if got := CastToText("SUM(x)"); !strings.Contains(got, "AS TEXT") {
		t.Fatalf("sqlite CastToText: %s", got)
	}
	if got := QuoteIdent("type"); got != `"type"` {
		t.Fatalf("sqlite QuoteIdent: %s", got)
	}
}

package models

import "testing"

func TestAPILogRetentionDefaults(t *testing.T) {
	values := make(map[string]string, len(defaultSettings))
	for _, setting := range defaultSettings {
		values[setting.Key] = setting.Value
	}

	if got := values["api_log_max_count"]; got != "5000" {
		t.Fatalf("api_log_max_count = %q, want %q", got, "5000")
	}
	if got := values["api_log_cleanup_interval_seconds"]; got != "600" {
		t.Fatalf("api_log_cleanup_interval_seconds = %q, want %q", got, "600")
	}
	if got := values["api_log_per_user_limit_enabled"]; got != "false" {
		t.Fatalf("api_log_per_user_limit_enabled = %q, want %q", got, "false")
	}
}

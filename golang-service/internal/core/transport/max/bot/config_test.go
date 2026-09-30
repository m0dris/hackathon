package core_max_bot

import "testing"

func TestNewConfigRejectsUnprotectedWebhook(t *testing.T) {
	for _, tc := range []struct {
		name, token, secret, webhook string
	}{
		{"empty token", "", "valid-secret", "https://demo.example.org/webhook"},
		{"empty secret", "test-token", "", "https://demo.example.org/webhook"},
		{"short secret", "test-token", "abc", "https://demo.example.org/webhook"},
		{"invalid secret", "test-token", "secret with spaces", "https://demo.example.org/webhook"},
		{"HTTP webhook", "test-token", "valid-secret", "http://demo.example.org/webhook"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MAX_BOT_TOKEN", tc.token)
			t.Setenv("MAX_BOT_SECRET", tc.secret)
			t.Setenv("MAX_BOT_WEBHOOK_URL", tc.webhook)
			if _, err := NewConfig(); err == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
}

func TestNewConfigAcceptsConfiguredWebhook(t *testing.T) {
	t.Setenv("MAX_BOT_TOKEN", "test-token")
	t.Setenv("MAX_BOT_SECRET", "valid-secret")
	t.Setenv("MAX_BOT_WEBHOOK_URL", "https://demo.example.org/webhook")
	if _, err := NewConfig(); err != nil {
		t.Fatal(err)
	}
}

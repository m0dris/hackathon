package core_max_bot

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Token string `envconfig:"TOKEN" required:"true"`

	Secret string `envconfig:"SECRET" required:"true"`

	WebhookURL string `envconfig:"WEBHOOK_URL" required:"true"`

	AutoSubscribe bool `envconfig:"AUTO_SUBSCRIBE" default:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("MAX_BOT", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	if strings.TrimSpace(config.Token) == "" {
		return Config{}, fmt.Errorf("MAX_BOT_TOKEN must not be empty")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{5,256}$`).MatchString(config.Secret) {
		return Config{}, fmt.Errorf("MAX_BOT_SECRET must contain 5-256 letters, digits, underscores or hyphens")
	}
	webhook, err := url.Parse(config.WebhookURL)
	if err != nil || webhook.Scheme != "https" || webhook.Hostname() == "" || webhook.User != nil {
		return Config{}, fmt.Errorf("MAX_BOT_WEBHOOK_URL must be an HTTPS URL without credentials")
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get MAX bot config: %w", err))
	}

	return config
}

package core_max_bot

import (
	"fmt"

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

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get MAX bot config: %w", err))
	}

	return config
}

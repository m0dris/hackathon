package core_grpc_client

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr string `envconfig:"ADDR" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("GRPC_DB", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get gRPC client config: %w", err))
	}

	return config
}

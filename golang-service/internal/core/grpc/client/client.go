package core_grpc_client

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewConn(config Config) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		config.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial db microservice: %w", err)
	}

	return conn, nil
}

package core_http_request

import (
	"context"
	"fmt"
)

func GetIntFromContext(ctx context.Context, key string) (int, error) {
	value, ok := ctx.Value(key).(int)
	if !ok {
		return 0, fmt.Errorf("no value in context")
	}

	return value, nil
}

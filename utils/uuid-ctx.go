package utils

import (
	"context"
	"fmt"

	"example.com/m/middleware"
)

func GetUserUUIDFromCtx(ctx context.Context) (string, error) {
	val := ctx.Value(middleware.UserUUIDKey)
	uuidStr, ok := val.(string)
	if !ok || uuidStr == "" {
		return "", fmt.Errorf("user UUID missing in request context")
	}
	return uuidStr, nil
}
